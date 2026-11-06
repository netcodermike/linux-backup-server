// An authenticated content-addressed encrypted backup service. See README before deployment.
package main
import (
 "bytes"
 "crypto/aes"
 "crypto/cipher"
 "crypto/rand"
 "crypto/sha256"
 "crypto/subtle"
 "encoding/hex"
 "encoding/json"
 _ "embed"
 "errors"
 "fmt"
 "io"
 "log"
 "net/http"
 "net/url"
 "os"
 "path/filepath"
 "regexp"
 "strings"
 "time"
)
var valid=regexp.MustCompile(`^[a-f0-9]{64}$`)
type Manifest struct { ID string `json:"id"`; Files map[string]string `json:"files"` }
func key() ([]byte,error){ b,e:=hex.DecodeString(os.Getenv("BACKUP_KEY"));if e!=nil||len(b)!=32{return nil,errors.New("BACKUP_KEY must be 64 hex characters")};return b,nil }
func seal(data,k []byte)([]byte,error){b,e:=aes.NewCipher(k);if e!=nil{return nil,e};g,e:=cipher.NewGCM(b);if e!=nil{return nil,e};n:=make([]byte,g.NonceSize());if _,e=rand.Read(n);e!=nil{return nil,e};return g.Seal(n,n,data,nil),nil}
func open(data,k []byte)([]byte,error){b,e:=aes.NewCipher(k);if e!=nil{return nil,e};g,e:=cipher.NewGCM(b);if e!=nil{return nil,e};if len(data)<g.NonceSize(){return nil,errors.New("short ciphertext")};return g.Open(nil,data[:g.NonceSize()],data[g.NonceSize():],nil)}
func atomic(path string,b []byte)error{
 if e:=os.MkdirAll(filepath.Dir(path),0700);e!=nil{return e};f,e:=os.CreateTemp(filepath.Dir(path),".tmp-");if e!=nil{return e};defer os.Remove(f.Name());if _,e=f.Write(b);e!=nil{f.Close();return e};if e=f.Close();e!=nil{return e};return os.Rename(f.Name(),path)
}
func serve()error{
 token:=os.Getenv("BACKUP_TOKEN");if len(token)<24{return errors.New("BACKUP_TOKEN must contain at least 24 characters")}
 root:=os.Getenv("BACKUP_DIR");if root==""{root="data"};if e:=os.MkdirAll(filepath.Join(root,"manifests"),0700);e!=nil{return e}
 mux:=http.NewServeMux()
 mux.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path=="/"&&r.Method=="GET"{w.Header().Set("Content-Type","text/html; charset=utf-8");io.WriteString(w,dashboard);return}
  w.Header().Set("WWW-Authenticate","Bearer")
  got:=strings.TrimPrefix(r.Header.Get("Authorization"),"Bearer ")
  if subtle.ConstantTimeCompare([]byte(got),[]byte(token))!=1{http.Error(w,"Unauthorized",401);return}
  // The web UI intentionally uses a password field rather than putting tokens in URLs.
  if r.URL.Path=="/manifests"&&r.Method=="GET"{
   entries,e:=os.ReadDir(filepath.Join(root,"manifests"));if e!=nil{http.Error(w,"Storage unavailable",500);return};ids:=[]string{};for _,f:=range entries{if valid.MatchString(f.Name()){ids=append(ids,f.Name())}};w.Header().Set("Content-Type","application/json");json.NewEncoder(w).Encode(ids);return
  }
  parts:=strings.Split(strings.Trim(r.URL.Path,"/"),"/");if len(parts)!=2||(parts[0]!="objects"&&parts[0]!="manifests")||!valid.MatchString(parts[1]){http.NotFound(w,r);return}
  path:=filepath.Join(root,parts[0],parts[1])
  if r.Method=="GET"{b,e:=os.ReadFile(path);if e!=nil{http.NotFound(w,r);return};w.Header().Set("Content-Type","application/octet-stream");w.Write(b);return}
  if r.Method!="PUT"{http.Error(w,"Use GET or PUT",405);return}
  body,e:=io.ReadAll(http.MaxBytesReader(w,r.Body,65<<20));if e!=nil{http.Error(w,"Object exceeds 65 MiB",413);return}
  // An existing content ID is immutable, preserving all referenced snapshots.
  if _,e=os.Stat(path);e==nil{w.WriteHeader(204);return}
  if e=atomic(path,body);e!=nil{http.Error(w,"Storage write failed",500);return};w.WriteHeader(201)
 })
 addr:=os.Getenv("BACKUP_ADDR");if addr==""{addr="127.0.0.1:8088"};log.Println("Vault listening on",addr)
 return (&http.Server{Addr:addr,Handler:mux,ReadHeaderTimeout:5*time.Second,ReadTimeout:60*time.Second,WriteTimeout:60*time.Second,MaxHeaderBytes:8192}).ListenAndServe()
}
var client=&http.Client{Timeout:90*time.Second}
func request(method,base,path string,body []byte)([]byte,error){
 u,e:=url.Parse(base);if e!=nil||u.Host==""||(u.Scheme!="http"&&u.Scheme!="https"){return nil,errors.New("invalid server URL")};u.Path=strings.TrimRight(u.Path,"/")+path
 req,e:=http.NewRequest(method,u.String(),bytes.NewReader(body));if e!=nil{return nil,e};req.Header.Set("Authorization","Bearer "+os.Getenv("BACKUP_TOKEN"));res,e:=client.Do(req);if e!=nil{return nil,e};defer res.Body.Close();if res.StatusCode>=300{return nil,fmt.Errorf("server status %d",res.StatusCode)};return io.ReadAll(io.LimitReader(res.Body,66<<20))
}
func backup(base,folder string)error{
 k,e:=key();if e!=nil{return e};root,e:=filepath.Abs(folder);if e!=nil{return e};info,e:=os.Stat(root);if e!=nil{return e};if !info.IsDir(){return errors.New("backup source must be a directory")}
 m:=Manifest{Files:map[string]string{}};salt:=make([]byte,32);if _,e=rand.Read(salt);e!=nil{return e};m.ID=hex.EncodeToString(salt)
 e=filepath.WalkDir(root,func(path string,d os.DirEntry,err error)error{
  if err!=nil{return err};if d.IsDir()||d.Type()&os.ModeSymlink!=0{return nil};info,e:=d.Info();if e!=nil{return e};if !info.Mode().IsRegular(){return nil};if info.Size()>64<<20{return fmt.Errorf("file too large: %s",path)}
  b,e:=os.ReadFile(path);if e!=nil{return e};h:=sha256.Sum256(b);id:=hex.EncodeToString(h[:]);encrypted,e:=seal(b,k);if e!=nil{return e};if _,e=request("PUT",base,"/objects/"+id,encrypted);e!=nil{return e};rel,e:=filepath.Rel(root,path);if e!=nil{return e};m.Files[filepath.ToSlash(rel)]=id;return nil
 });if e!=nil{return e};b,e:=json.Marshal(m);if e!=nil{return e};b,e=seal(b,k);if e!=nil{return e};if _,e=request("PUT",base,"/manifests/"+m.ID,b);e!=nil{return e};fmt.Println(m.ID);return nil
}
func restore(base,id,destination string)error{
 if !valid.MatchString(id){return errors.New("invalid snapshot ID")};k,e:=key();if e!=nil{return e};b,e:=request("GET",base,"/manifests/"+id,nil);if e!=nil{return e};b,e=open(b,k);if e!=nil{return e};var m Manifest;if e=json.Unmarshal(b,&m);e!=nil{return e}
 root,e:=filepath.Abs(destination);if e!=nil{return e};if e=os.Mkdir(root,0700);e!=nil{return errors.New("destination must not already exist and its parent must exist")}
 for name,hash:=range m.Files{
  clean:=filepath.Clean(filepath.FromSlash(name));if clean=="."||filepath.IsAbs(clean)||clean==".."||strings.HasPrefix(clean,".."+string(os.PathSeparator))||strings.Contains(clean,":")||!valid.MatchString(hash){return errors.New("unsafe manifest entry")}
  b,e=request("GET",base,"/objects/"+hash,nil);if e!=nil{return e};b,e=open(b,k);if e!=nil{return e};sum:=sha256.Sum256(b);if hex.EncodeToString(sum[:])!=hash{return errors.New("object hash mismatch")};if e=atomic(filepath.Join(root,clean),b);e!=nil{return e}
 };return nil
}
func main(){a:=os.Args[1:];var e error
 switch{case len(a)==1&&a[0]=="serve":e=serve();case len(a)==3&&a[0]=="backup":e=backup(a[1],a[2]);case len(a)==4&&a[0]=="restore":e=restore(a[1],a[2],a[3]);case len(a)==2&&a[0]=="list":var b []byte;b,e=request("GET",a[1],"/manifests",nil);if e==nil{fmt.Println(string(b))};default:fmt.Fprintln(os.Stderr,"Usage: vault serve | backup URL FOLDER | list URL | restore URL ID NEW_FOLDER");os.Exit(2)}
 if e!=nil{log.Fatal(e)}
}
//go:embed public/index.html
var dashboard string
