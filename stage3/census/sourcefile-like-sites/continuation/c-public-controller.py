import pathlib,sys,json,time,shutil
sys.dont_write_bytecode=True
from variants import invoke,core,checkpoint
p=pathlib.Path(__file__).resolve().parent
while not (p/'original-unstripped/public-signatures.json').exists():time.sleep(2)
for label,base in [('original','original'),('c','original'),('original-unstripped','original-unstripped'),('c-unstripped','original-unstripped')]:
 invoke(['node',str(p/'api-c.cjs'),label,base],p/label/'api-analysis.log',600)
 if not label.endswith('unstripped'):invoke(['python3',str(p/'lane-api-c.py'),label],p/label/'lane-api-check.log',600)
if json.loads((p/'c/lane-api-check.json').read_text())['errors']:
 root=pathlib.Path('/workspace/scratch/baobab-road2-c-public');assert not root.exists();shutil.copytree('/workspace/scratch/baobab-road2-c',root)
 f=root/'src/compiler/types.ts';before=f.read_bytes();needle=b'/** @internal */\r\nexport interface LineMapSource';assert before.count(needle)==1;f.write_bytes(before.replace(needle,b'export interface LineMapSource',1))
 (p/'c-public').mkdir(exist_ok=True)
 (p/'c-public/source-marker-diff.json').write_text(json.dumps({'only_change_from_c':'remove LineMapSource @internal marker','runtime_and_relation_types_identical':True},indent=2)+'\n')
 for label,strip in [('c-public','on'),('c-public-unstripped','off')]:
  invoke(['node',str(p/'emit-public-c.cjs'),str(root),label,strip],p/label/'public-emit.log',600)
  invoke(['node',str(p/'api-c.cjs'),label,'original' if strip=='on' else 'original-unstripped'],p/label/'api-analysis.log',600)
 invoke(['python3',str(p/'lane-api-c.py'),'c-public'],p/'c-public/lane-api-check.log',600)
 invoke(['node',str(p/'source-audit-c.cjs'),str(root),'c-public-exact','exact'],p/'c-public-exact/source-audit.log',600)
 core('c-public',root)
 checkpoint('road2-c-public-api')
print('c public workflow child exit=0',flush=True)
