import pathlib,subprocess,os,time,json
s=pathlib.Path('/tmp/watch-functions-1532')
# Wait for the baseline producer, without overlapping writes to its output.
while pathlib.Path('/proc/1281').exists(): time.sleep(1)
tree=s/'main-tree'
assert (tree/'patch-set.md').exists(), 'baseline apply did not finish'
p=tree/'src/compiler/sys.ts';raw=p.read_bytes();text=raw.decode();newline='\r\n' if '\r\n' in text else '\n'
needle='            watchFile,'+newline+'            watchDirectory,'
assert text.count(needle)==1
replacement='''            watchFile: (fileName, callback, pollingInterval, options) => {
                _fs.appendFileSync(process.env.ADAMIC_WATCH_COUNT!, JSON.stringify({ interval: String(pollingInterval), type: typeof pollingInterval }) + "\\n");
                return watchFile(fileName, callback, pollingInterval as PollingInterval, options);
            },
            watchDirectory,'''.replace('\n',newline)
(s/'sys-count-before.ts').write_bytes(raw);p.write_bytes(text.replace(needle,replacement).encode())
(s/'count-instrumentation.diff').write_text(subprocess.check_output(['git','diff','--no-index',str(s/'sys-count-before.ts'),str(p)],text=True) if False else replacement)
env=dict(os.environ,ADAMIC_WATCH_COUNT=str(s/'lane-intervals.jsonl'),TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'tmp'),GOCACHE=str(s/'go-cache'),npm_config_cache=str(s/'npm-cache'),STAGE3_CACHE=str(s/'stage3-cache'),ADAMIC_BUILD_CACHE_DIR=str(s/'build-cache'))
cmd=['python3',str(s/'work/stage3/oracle/run.py'),str(tree),str(s/'count-oracle')]
(s/'count-command.json').write_text(json.dumps(cmd))
with (s/'count-oracle.log').open('w') as log:
 r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
(s/'count-exit.txt').write_text(str(r.returncode))
