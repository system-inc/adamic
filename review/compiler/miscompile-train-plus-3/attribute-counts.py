from pathlib import Path
import subprocess,json,re,difflib,tempfile,time
out=Path(__file__).parent
artifacts=out/'count-attribution'; artifacts.mkdir(exist_ok=True)
def rows(text):
 result={}
 for line in text.splitlines():
  fields=[s.strip() for s in line.split('|')[1:-1]]
  if len(fields)==7 and all(s.isdigit() for s in fields[1:]): result[fields[0]]=list(map(int,fields[1:]))
 return result
old=rows(subprocess.check_output(['git','show','51aa3a96:internal/oracle/counts.md'],text=True)); new=rows(Path('internal/oracle/counts.md').read_text())
assert old.keys()==new.keys()
pattern=re.compile(r'adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)')
results=[]
with tempfile.TemporaryDirectory(prefix='adamic-count-attribution-') as directory:
 for index,path in enumerate(p for p in new if old[p]!=new[p]):
  result=dict(fixture=path,before=old[path],after=new[path],delta=[b-a for a,b in zip(old[path],new[path])],cause_commit='6a056842ff91f8d260aa086675050a61a5e2f2f4')
  codes={}
  for label,cli,want in [('base','/tmp/adamic-train-plus-3-base-cli',old[path]),('fixed','/tmp/adamic-train-plus-3-fixed-cli',new[path])]:
   command=[cli,'c',path]
   code=subprocess.run(command,capture_output=True,text=True,timeout=40); assert code.returncode==0,code.stderr
   codes[label]=code.stdout
   (artifacts/(str(index)+'-'+label+'.c.txt')).write_text(code.stdout)
   binary=str(Path(directory)/(str(index)+'-'+label))
   with (artifacts/(str(index)+'-'+label+'-build.log')).open('w') as log:
    build=subprocess.run([cli,'build',path,'-o',binary,'--count'],stdout=log,stderr=subprocess.STDOUT,timeout=40)
   assert build.returncode==0
   run=subprocess.run([binary],capture_output=True,text=True,timeout=30)
   (artifacts/(str(index)+'-'+label+'-run.log')).write_text(run.stdout+run.stderr)
   match=pattern.search(run.stderr); assert match,(path,label,run.stderr)
   measured=list(map(int,match.groups())); assert measured==want,(path,label,measured,want)
   result[label+'_measured']=measured
  diff=''.join(difflib.unified_diff(codes['base'].splitlines(True),codes['fixed'].splitlines(True),fromfile='51aa3a96',tofile='6a056842'))
  (artifacts/(str(index)+'.patch')).write_text(diff)
  result['removed_member_helper_calls']=codes['base'].count('checked_view_members(')-codes['fixed'].count('checked_view_members(')
  result['new_uniqueness_checks']=codes['fixed'].count('->heap.references == 1')-codes['base'].count('->heap.references == 1')
  assert result['removed_member_helper_calls']>0,(path,'no named member helper removal in C')
  result['cause']='Removed unchecked member-helper identity calls after readiness; existing borrow, consumed-parameter and uniqueness plans operate on the restored original operand. Direct base and fixed counted runs match both table rows.'
  results.append(result)
  (out/'count-attribution.json').write_text(json.dumps(results,indent=2)+'\n')
  print({k:v for k,v in result.items() if k in ['fixture','delta','removed_member_helper_calls','new_uniqueness_checks']},flush=True)
assert len(results)==38
print('PASS: all 38 moved rows directly reproduced on base and fixed compilers; no rows added or removed',flush=True)
