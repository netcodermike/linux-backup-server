"""End-to-end integration check using the built vault executable and synthetic data."""
import json, os, pathlib, secrets, socket, subprocess, tempfile, time, urllib.request, urllib.error
binary=str(pathlib.Path('vault.exe' if os.name=='nt' else 'vault').resolve())
with tempfile.TemporaryDirectory() as tmp:
    root=pathlib.Path(tmp);source=root/'source';source.mkdir();(source/'file.txt').write_text('first revision')
    with socket.socket() as s:s.bind(('127.0.0.1',0));port=s.getsockname()[1]
    base=f'http://127.0.0.1:{port}'
    env=dict(os.environ,BACKUP_TOKEN=secrets.token_hex(32),BACKUP_KEY=secrets.token_hex(32),BACKUP_DIR=str(root/'data'),BACKUP_ADDR=f'127.0.0.1:{port}')
    server=subprocess.Popen([binary,'serve'],env=env,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
    def run(*args):return subprocess.check_output([binary,*args],env=env,text=True).strip()
    try:
        for attempt in range(100):
            try:run('list',base);break
            except subprocess.CalledProcessError:time.sleep(.05)
        else:raise AssertionError('server did not start')
        first=run('backup',base,str(source));run('backup',base,str(source))
        assert len(list((root/'data/objects').iterdir()))==1,'unchanged file not deduplicated'
        (source/'file.txt').write_text('second revision');second=run('backup',base,str(source))
        run('restore',base,first,str(root/'first'));run('restore',base,second,str(root/'second'))
        assert (root/'first/file.txt').read_text()=='first revision'
        assert (root/'second/file.txt').read_text()=='second revision'
        assert first in json.loads(run('list',base))
        try:urllib.request.urlopen(base+'/manifests');raise AssertionError('missing token accepted')
        except urllib.error.HTTPError as e:assert e.code==401
        print('Encrypted backup, deduplication, authentication, and restore checks passed')
    finally:server.terminate();server.wait(timeout=10)
