import pathlib,json,difflib,subprocess,os,time
repo=pathlib.Path('/workspace/adamic');r=repo/'review/test-audit/stage3-fixtures-fixtures';cases=[('M1','internal/lower/expression.go','return ir.BooleanConstant{Value: node.Kind == ast.KindTrueKeyword}, nil','return ir.BooleanConstant{Value: node.Kind != ast.KindTrueKeyword}, nil','flip condition'),('M2','internal/lower/diagnostics.go',"stage 0 can't lower %s yet","stage 0 cannot lower %s yet",'change constant'),('M3','internal/native/emit.go','bodies.WriteString("\\treturn 0;\\n}\\n")','bodies.WriteString("\\treturn 1;\\n}\\n")','change constant')]
plan=[];original={}
for mid,file,a,b,menu in cases:
 p=repo/file;s=p.read_text();original[p]=s;assert s.count(a)==1;line=s[:s.index(a)].count('\n')+1;plan.append(dict(id=mid,file=file,line=line,change=b,old=a,menu=menu));v=s.replace(a,b);(r/(mid+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),v.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
# Freeze all menu mutations before any outcome runs.
(r/'plan.json').write_text(json.dumps(plan,indent=2))
# Standalone diffs each compile; source changes are restored before switch installation.
for mid,file,a,b,menu in cases:
 p=repo/file;s=original[p];p.write_text(s.replace(a,b))
 try:
  with (r/'logs'/f'{mid}-vet.log').open('w') as out:rc=subprocess.run(['go','vet','./'+str(pathlib.Path(file).parent)+'/'],stdout=out,stderr=subprocess.STDOUT).returncode
  if rc:raise RuntimeError(mid+' standalone vet failed')
 finally:p.write_text(s)
# Selector plumbing only; standalone diffs above contain only fixed-menu changes.
for mid,file,a,b,menu in cases:
 p=repo/file;s=original[p]
 if mid=='M1':new='if os.Getenv("ADAMIC_MUTANT") == "M1" { '+b+' }; '+a
 elif mid=='M2':
  a='return fmt.Sprintf("%s: stage 0 can\'t lower %s yet", n.Where, n.What)';b='return fmt.Sprintf("%s: stage 0 cannot lower %s yet", n.Where, n.What)';new='if os.Getenv("ADAMIC_MUTANT") == "M2" { '+b+' }; '+a
 else:new='if os.Getenv("ADAMIC_MUTANT") == "M3" { '+b+' } else { '+a+' }'
 s=s.replace(a,new).replace('import (','import (\n "os"',1);p.write_text(s)
# Lower empty-entry probe is not a mutation kill.
p=repo/'internal/lower/lower.go';original[p]=p.read_text();p.write_text(original[p].replace('import (','import (\n "os"',1).replace('files := program.Files()','if os.Getenv("ADAMIC_MUTANT") == "P1" { return nil, nil }; files := program.Files()',1))
(r/'switch.diff').write_text(subprocess.check_output(['git','diff'],cwd=repo,text=True))
results=[]
try:
 for mid in ('M1','M2','M3','P1'):
  env=os.environ.copy();env.update(ADAMIC_MUTANT=mid,ADAMIC_BUILD_CACHE_DIR='/tmp/u163/cache/'+mid,ADAMIC_STAGE3_BUILD_STORE='/tmp/u163/hooks/switch');cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage3/fixtures/','-run','.'];start=time.monotonic()
  with (r/'logs'/f'{mid}.log').open('w') as out:rc=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT).returncode
  results.append(dict(id=mid,command=cmd,wall=time.monotonic()-start,rc=rc));(r/'matrix-runs.json').write_text(json.dumps(results,indent=2))
finally:
 for p,s in original.items():p.write_text(s)
