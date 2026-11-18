from pathlib import Path
import os,time,subprocess,sys,json,gzip,hashlib
root=Path('/tmp/family5-evidence/c02');status=root/'whole-apply/status.json'
while not status.exists():time.sleep(5)
if json.loads(status.read_text())['exit_code']:sys.exit(1)
def run(name,cmd,env=None):
 code=subprocess.call([sys.executable,'/tmp/family5-proof.py',str(root/name)]+cmd,env=env)
 if code:sys.exit(code)
tree=Path('/workspace/family5-c02-tree')
run('stock',['node','/tmp/family5-stock.cjs',str(tree),str(root/'stock.json')])
run('stock-strict',['node','/tmp/family5-stock.cjs',str(tree),str(root/'stock-strict.json')],dict(os.environ,FAMILY5_STOCK_STRICT='1'))
r=json.loads((root/'stock.json').read_text());run('adamic-types',['/tmp/family5-tools/adamic','types']+[str(tree/p) for p in r['roots']])
node=tree/'node_modules';node.symlink_to('/tmp/family5-land/stage3/typescript/node_modules',target_is_directory=True)
os.chdir(tree);run('public-build',['node','/tmp/family5-land/stage3/typescript/node_modules/hereby/bin/hereby.js','local']);api=(tree/'built/local/typescript.d.ts').read_bytes();before=gzip.decompress(Path('/tmp/family5-evidence/setup/public-before.d.ts.gz').read_bytes());(root/'public-after.d.ts.gz').write_bytes(gzip.compress(api,mtime=0));node.unlink();same=api==before
(root/'public-comparison.json').write_text(json.dumps({'byte_identical':same,'before_sha256':hashlib.sha256(before).hexdigest(),'after_sha256':hashlib.sha256(api).hexdigest()},indent=2)+'\n');print(('PASS' if same else 'FAIL')+' public declaration byte equality',flush=True)
if not same:sys.exit(1)
os.chdir('/tmp/family5-land')
sys.exit(subprocess.call([sys.executable,'/tmp/family5-census.py','c02-after','program.ts','utilities.ts','parser.ts','checker.ts'],env=dict(os.environ,FAMILY5_TREE=str(tree))))
