from pathlib import Path
import os,subprocess,time,json
s=Path('/tmp/watch-file-check-cead');tree=Path('/tmp/watch-functions-1532/main-tree');p=tree/'src/compiler/sys.ts'
while not (s/'CENSUS_OK').exists():time.sleep(1)
raw=(s/'sys-main.ts').read_bytes();assert p.read_bytes()==raw;nl=b'\r\n' if b'\r\n' in raw else b'\n'
needle=b'    function watchFile(fileName: string, callback: FileWatcherCallback, defaultPollingInterval: PollingInterval): FileWatcher {'
logger=b'        _fs.appendFileSync(process.env.ADAMIC_DYNAMIC_COUNT!, JSON.stringify({ interval: String(defaultPollingInterval), type: typeof defaultPollingInterval }) + "\\n");'
assert raw.count(needle)==1;p.write_bytes(raw.replace(needle,needle+nl+logger));(s/'sys-count.ts').write_bytes(p.read_bytes())
env=dict(os.environ,ADAMIC_DYNAMIC_COUNT=str(s/'dynamic-lane.jsonl'),TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'tmp'),GOCACHE='/tmp/watch-file-check-1532/go-cache',npm_config_cache='/tmp/watch-functions-1532/npm-cache',STAGE3_CACHE=str(s/'cache'),ADAMIC_BUILD_CACHE_DIR=str(s/'cache'),PYTHONDONTWRITEBYTECODE='1')
cmd=['python3',str(s/'work/stage3/oracle/run.py'),str(tree),str(s/'count-oracle')];(s/'count-command.json').write_text(json.dumps(cmd))
try:
 with (s/'count-oracle.log').open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 (s/'count-exit.txt').write_text(str(r.returncode))
finally:p.write_bytes(raw)
