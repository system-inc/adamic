import pathlib,sys,json,shutil,hashlib,difflib
sys.dont_write_bytecode=True
from variants import invoke,core,checkpoint
p=pathlib.Path(__file__).resolve().parent
for label,root in [('c-unstripped','/workspace/scratch/baobab-road2-c'),('original-unstripped','/workspace/scratch/baobab-adapted')]:
 if not (p/label/'public-signatures.json').exists():invoke(['node',str(p/'emit-public-c.cjs'),root,label,'off'],p/label/'public-emit.log',600)
for label,base in [('original','original'),('original-unstripped','original-unstripped'),('c-unstripped','original-unstripped')]:
 invoke(['node',str(p/'api-c.cjs'),label,base],p/label/'api-analysis.log',600)
 if label in ['original','c-unstripped']:invoke(['python3',str(p/'lane-api-c.py'),label],p/label/'lane-api-check.log',600)
root=pathlib.Path('/workspace/scratch/baobab-road2-c-public');assert not root.exists();shutil.copytree('/workspace/scratch/baobab-road2-c',root)
f=root/'src/compiler/types.ts';before=f.read_bytes();needle=b'/** @internal */\r\nexport interface LineMapSource';assert before.count(needle)==1
# Replace the marker with a blank comment to preserve all source coordinates.
after=before.replace(needle,b'/**              */\r\nexport interface LineMapSource',1);f.write_bytes(after)
(p/'c-public').mkdir(exist_ok=True)
(p/'c-public/source-marker-diff.json').write_text(json.dumps({'only_change_from_c':'remove LineMapSource @internal tag, retaining a blank comment to preserve source coordinates','runtime_and_relation_types_identical':True,'before':hashlib.sha256(before).hexdigest(),'after':hashlib.sha256(after).hexdigest()},indent=2)+'\n')
(p/'c-public/types.ts.patch').write_text(''.join(difflib.unified_diff(before.decode().splitlines(True),after.decode().splitlines(True),fromfile='c/types.ts',tofile='c-public/types.ts')))
for label,strip in [('c-public','on'),('c-public-unstripped','off')]:
 invoke(['node',str(p/'emit-public-c.cjs'),str(root),label,strip],p/label/'public-emit.log',600)
 invoke(['node',str(p/'api-c.cjs'),label,'original' if strip=='on' else 'original-unstripped'],p/label/'api-analysis.log',600)
invoke(['python3',str(p/'lane-api-c.py'),'c-public'],p/'c-public/lane-api-check.log',600)
invoke(['node',str(p/'source-audit-c.cjs'),str(root),'c-public-exact','exact'],p/'c-public-exact/source-audit.log',600)
core('c-public',root)
checkpoint('road2-c-public-api')
print('c public workflow child exit=0',flush=True)
