#!/usr/bin/env python3
"""Run the unchanged oracle and lane guard on the already full-applied tree."""
import hashlib,json,os,pathlib,shutil,subprocess,sys
root=pathlib.Path(__file__).resolve().parents[3]
tree=pathlib.Path(sys.argv[1]).resolve();out=pathlib.Path(sys.argv[2]).resolve()
apply_result=pathlib.Path(sys.argv[3]).resolve();proof=json.loads(apply_result.read_text())
assert proof['exit']==0 and pathlib.Path(proof['tree']).resolve()==tree
assert (tree/'patch-set.md').exists()
out.mkdir(parents=True,exist_ok=False)
(out/'adapted-tree').symlink_to(tree,target_is_directory=True)
shutil.copyfile(proof['log'],out/'apply.log');shutil.copyfile(tree/'patch-set.md',out/'patch-set.md')
shutil.copyfile(apply_result,out/'apply-result.json')
def snapshot():
 return {str(p.relative_to(tree/'src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
before=snapshot();env=dict(os.environ,NODE_PATH=str(root/'stage3/api/node_modules'))
replay=subprocess.run(['node',str(pathlib.Path(__file__).with_name('adapt.cjs')),str(tree)],env=env,capture_output=True,text=True)
(out/'replay.log').write_text(replay.stdout+replay.stderr)
assert replay.returncode==0
result=json.loads(replay.stdout);assert result=={'files':0,'edits':0,'javascriptChanging':True},result
assert before==snapshot()
(out/'replay.json').write_text(json.dumps(dict(result,sourceBytesIdentical=True),indent=2)+'\n')
platform=json.loads(subprocess.check_output(['node','-p','JSON.stringify({os:process.platform,arch:process.arch,node:process.version})'],text=True))
execution={'commit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(),'apply_exit':proof['exit'],'platform':platform,'apply_provenance':'prior full stage3/apply.sh, exact tree reused; see apply-result.json','oracle_exit':None}
with (out/'oracle.log').open('w') as log:
 p=subprocess.run(['bash',str(root/'stage3/oracle/run.sh'),str(tree),str(out/'oracle')],cwd=root,stdout=log,stderr=subprocess.STDOUT)
execution['oracle_exit']=p.returncode
(out/'execution.json').write_text(json.dumps(execution,indent=2)+'\n')
sys.path.insert(0,str(root/'stage3/lane'))
from check import check_results,write_verdict
verdict=check_results(out)
raise SystemExit(write_verdict(out,verdict))
