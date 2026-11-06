package main
import("bytes";"testing")
func TestEncryption(t *testing.T){k:=bytes.Repeat([]byte{7},32);plain:=[]byte("precious files");a,e:=seal(plain,k);if e!=nil{t.Fatal(e)};b,e:=open(a,k);if e!=nil||!bytes.Equal(b,plain){t.Fatal("round trip",e)};a[len(a)-1]^=1;if _,e=open(a,k);e==nil{t.Fatal("tampering accepted")}}
func TestInvalidKey(t *testing.T){if _,e:=seal([]byte("x"),[]byte("short"));e==nil{t.Fatal("invalid AES key accepted")}}
