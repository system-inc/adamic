import subprocess,pathlib,json,os
root=pathlib.Path('review/compiler/chain-slices')
rows=json.loads((root/'member-patch-checks.json').read_text())
env=dict(os.environ,GIT_INDEX_FILE='/tmp/chain-slices-main.index')
subprocess.run(['git','read-tree','origin/main'],env=env,check=True)
for r in rows:
 patch=subprocess.check_output(['git','diff','--binary',r['base'],r['source'],'--','.',':!review',':!docs',':!internal/oracle/counts.md'])
 p=subprocess.run(['git','apply','--cached','--check'],input=patch,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
 r['patch_check_exit']=p.returncode;r['diagnostic']=p.stderr.decode()
 r['check_scope']='branch net excluding review, docs and counts; isolated origin/main index; stacked source may include its dependencies'
(root/'member-patch-checks.json').write_text(json.dumps(rows,indent=2)+'\n')
print('\n'.join(f"{r['member']}: {'clean' if r['patch_check_exit']==0 else 'conflicts'}" for r in rows))
