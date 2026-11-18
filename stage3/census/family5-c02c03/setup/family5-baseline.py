from pathlib import Path
import os,time,subprocess,sys,json,gzip
while Path('/proc/27409/cmdline').exists() and b'apply.py' in Path('/proc/27409/cmdline').read_bytes():time.sleep(5)
root=Path('/tmp/family5-evidence/setup')
def run(name,cmd,env=None):
 code=subprocess.call([sys.executable,'/tmp/family5-proof.py',str(root/name)]+cmd,env=env)
 if code:sys.exit(code)
run('baseline-stock-settled',['node','/tmp/family5-stock.cjs','/tmp/family5-base-tree',str(root/'baseline-stock-settled.json')])
env=dict(os.environ,FAMILY5_STOCK_STRICT='1');run('baseline-stock-strict',['node','/tmp/family5-stock.cjs','/tmp/family5-base-tree',str(root/'baseline-stock-strict.json')],env)
report=json.loads((root/'baseline-stock-settled.json').read_text());files=['/tmp/family5-base-tree/'+p for p in report['roots']]
run('baseline-adamic-types',['/tmp/family5-tools/adamic','types']+files)
node=Path('/tmp/family5-base-tree/node_modules');node.symlink_to('/tmp/family5-land/stage3/typescript/node_modules',target_is_directory=True)
os.chdir('/tmp/family5-base-tree')
run('baseline-public-build',['node','/tmp/family5-land/stage3/typescript/node_modules/hereby/bin/hereby.js','local'])
api=Path('built/local/typescript.d.ts').read_bytes();(root/'public-before.d.ts.gz').write_bytes(gzip.compress(api,mtime=0));node.unlink()
os.chdir('/tmp/family5-land')
sys.exit(subprocess.call([sys.executable,'/tmp/family5-census.py','c02-before','program.ts','utilities.ts','parser.ts','checker.ts']))
