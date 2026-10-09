import pathlib,subprocess,os,json,time,difflib
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/internal-ir';os.chdir(root)
spec=[('P08','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','return &ir.Program{}, nil'),('P09','internal/ir/call_targets.go','func (p *Program) ClosureReadsArgumentsCount(call CallClosure) bool {','return false')]
originals={};menu=json.loads((out/'menu.json').read_text());meta=json.loads((out/'run-meta.json').read_text())
for mid,f,a,b in spec:
 p=root/f;s=p.read_text();originals[f]=s
 menu.append(dict(id=mid,file=f,line=s[:s.index(a)].count('\n')+1,before=a,after=b,kind='probe'))
 if f.endswith('lower.go'):c=s.replace('"context"','"context"\n "os"',1)
 else:c=s.replace('package ir\n','package ir\nimport "os"\n',1)
 c=c.replace(a,a+'\n if os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { '+b+' }',1);p.write_text(c)
 (out/(mid+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),c.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
(out/'menu.json').write_text(json.dumps(menu,indent=2)+'\n')
for mid,_,_,_ in spec:
 env=os.environ.copy();env['ADAMIC_MUTANT']=mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u022/cache/'+mid
 cmd='timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .';t=time.monotonic()
 with (out/(mid+'.log')).open('w') as log:r=subprocess.run(cmd,shell=True,env=env,stdout=log,stderr=subprocess.STDOUT)
 meta[mid]=dict(command=cmd,status=r.returncode,wall=time.monotonic()-t)
for f,s in originals.items():(root/f).write_text(s)
(out/'run-meta.json').write_text(json.dumps(meta,indent=2)+'\n')
