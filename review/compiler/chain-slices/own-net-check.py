import subprocess,json,pathlib,os
root=pathlib.Path('review/compiler/chain-slices')
rows=json.loads((root/'member-patch-checks.json').read_text())
env=dict(os.environ,GIT_INDEX_FILE='/tmp/chain-slices-main.index')
subprocess.run(['git','read-tree','5e33a17b'],env=env,check=True)
bases={'step24-parser-main':'bafb9ef4','per-backend-stops':'b7c8d7a8','search-shrink':'2391c655','optional-presence-next':'2391c655','eep-presence':'897d0e79'}
for row in rows:
 if row['member'] in bases: row['base']=subprocess.check_output(['git','rev-parse',bases[row['member']]],text=True).strip()
 scope=['internal/native/tsgo.go'] if row['member']=='tsgo.go-cache' else ['.',':!review',':!docs',':!internal/oracle/counts.md']
 patch=subprocess.check_output(['git','diff','--binary',row['base'],row['source'],'--',*scope])
 p=subprocess.run(['git','apply','--cached','--check'],input=patch,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
 row['patch_check_exit']=p.returncode;row['diagnostic']=p.stderr.decode()
 row['files']=subprocess.check_output(['git','diff','--name-only',row['base'],row['source'],'--',*scope],text=True).splitlines()
 row['check_scope']='Own source range where identified; optional-calls still includes inherited lowering-a and requires extraction. Counts/docs/review excluded; chain fixes audited separately, not included in this applicability check.'
(root/'own-net-patch-checks.json').write_text(json.dumps(rows,indent=2)+'\n')
print('\n'.join(f"{r['member']}: {'clean' if r['patch_check_exit']==0 else 'conflicts'} ({len(r['files'])} files)" for r in rows))
