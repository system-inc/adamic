from pathlib import Path
import subprocess,shutil,json
root=Path(__file__).resolve().parents[3];out=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch/oracle-counts-merge-proof');scratch.mkdir(parents=True,exist_ok=True)
def git(*args):return subprocess.check_output(['git',*args],cwd=scratch,text=True).strip()
assert not (scratch/'.git').exists();git('init','-q');git('config','user.name','Counts merge proof');git('config','user.email','counts-proof@example.invalid')
source=root/'internal/oracle/testdata/counts';target=scratch/'internal/oracle/testdata/counts';shutil.copytree(source,target);git('add','.');git('commit','-qm','Recorded per-fixture migration baseline');base=git('rev-parse','HEAD')
keys=['dedication/dedication.a','internal/oracle/testdata/borrow_chain_walk.a'];commits=[];expected={}
for i,key in enumerate(keys):
 git('checkout','-q','--detach',base);p=target/'runtime'/(key+'.counts');cells=p.read_text().split('|');cells[2]=' '+str(int(cells[2])+1)+' ';p.write_text('|'.join(cells));relative=str(p.relative_to(scratch));expected[relative]=p.read_text();git('add',relative);git('commit','-qm','Move only '+key+' counts');commits.append(git('rev-parse','HEAD'))
command=['git','merge-tree','--write-tree','--merge-base='+base,*commits];result=subprocess.run(command,cwd=scratch,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=30);(out/'merge-proof-output.txt').write_text(result.stdout);assert result.returncode==0;tree=result.stdout.splitlines()[0]
for path,value in expected.items():assert git('show',tree+':'+path)==value.strip()
changed=git('diff','--name-only',base,tree).splitlines();assert set(changed)==set(expected)
record=dict(base=base,sibling_commits=commits,parents=[git('rev-parse',c+'^') for c in commits],merged_tree=tree,exit=result.returncode,changed_files=changed,both_edits_preserved=True);assert record['parents']==[base,base];(out/'merge-proof.json').write_text(json.dumps(record,indent=2)+'\n');print('PASS: two sibling commits, two distinct fixture files, clean merge, both edits preserved')
