import pathlib,json,subprocess,time,difflib
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-json-audit';plan=json.loads(pathlib.Path('/tmp/u102/plan.json').read_text());base={x['file']:subprocess.check_output(['git','show','origin/main:'+x['file']],cwd=root,text=True) for x in plan};records=[]
# Restore original files before independently replaying each standalone diff.
for f,s in base.items():(root/f).write_text(s)
for f in ['internal/lower/audit_u102.go','internal/native/audit_u102.go','stage1/cohere/json/audit_u102_test.go']:(root/f).unlink(missing_ok=True)
for x in plan:
 f=x['file'];s=x['source']
 if x['id'] in ['W04','P09']:head,tail=s.split('func TestCachedWidthMutantIsCaught',1);s=head+'func TestCachedWidthMutantIsCaught'+tail.replace('input, expected := protocol(cases, answers)','input, _ := protocol(cases, answers)',1)
 (root/f).write_text(s)
 subprocess.run(['gofmt','-w',str(root/f)],check=True)
 s=(root/f).read_text();x['source']=s
 patch=''.join(difflib.unified_diff(base[f].splitlines(True),s.splitlines(True),fromfile='a/'+f,tofile='b/'+f));(out/(x['id']+'.diff')).write_text(patch)
 # Apply check after restoring proves the diff applies independently to clean origin/main source.
 (root/f).write_text(base[f]);check=subprocess.run(['git','apply','--check',str(out/(x['id']+'.diff'))],cwd=root,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
 (root/f).write_text(s)
 package='./'+str(pathlib.Path(f).parent)+'/'
 cmd=['timeout','90','go','vet',package];t=time.monotonic()
 with (out/(x['id']+'-vet.log')).open('w') as log:p=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT)
 records.append(dict(id=x['id'],apply_check_exit=check.returncode,apply_check_output=check.stdout,command=cmd,exit=p.returncode,wall_seconds=time.monotonic()-t))
 (root/f).write_text(base[f]);pathlib.Path('/tmp/u102/validation.json').write_text(json.dumps(records,indent=2));print(x['id'],check.returncode,p.returncode,round(records[-1]['wall_seconds'],2),flush=True)
 if p.returncode:raise SystemExit('Fix standalone compile before continuing')
pathlib.Path('/tmp/u102/plan.json').write_text(json.dumps(plan))
