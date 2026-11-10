import pathlib, subprocess, sys, json
root=pathlib.Path(__file__).resolve().parents[3]; out=pathlib.Path(__file__).parent
label=sys.argv[1]
def git(*args):return subprocess.check_output(['git',*args],cwd=root,text=True).strip()
subprocess.run(['git','add','--',str(out.relative_to(root))],cwd=root,check=True)
tree=git('write-tree')
prev=pathlib.Path('/workspace/overload-checkpoint-sha')
parent=prev.read_text().strip() if prev.exists() else git('rev-parse','1d84074f')
sha=git('commit-tree',tree,'-p',parent,'-m','census checkpoint: '+label)
log=out/('checkpoint-'+label+'.log.txt')
with log.open('w') as f:
 code=subprocess.run(['git','push','origin',sha+':refs/heads/typescript/overload-callback-census-checkpoints'],cwd=root,stdout=f,stderr=subprocess.STDOUT).returncode
print(json.dumps(dict(checkpoint=label,sha=sha,child_exit=code)),flush=True)
if code==0:prev.write_text(sha+'\n')
sys.exit(code)
