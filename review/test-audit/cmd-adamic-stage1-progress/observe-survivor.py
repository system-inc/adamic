from pathlib import Path
import tempfile,subprocess,os,json,time,shlex
p=Path('review/test-audit/cmd-adamic-stage1-progress'); scratch=Path(tempfile.mkdtemp(prefix='u010-observe-')); source=Path('cmd/adamic-stage1-progress/main.go').read_text(); assert source.count('func main() {')==1
(scratch/'implementation.go').write_text(source.replace('func main() {','func unusedCommandMain() {',1));(scratch/'coverage.go').write_text(Path('cmd/adamic-stage1-progress/coverage.go').read_text()); observer='package main\nimport ("encoding/json"; "os")\nfunc main() { json.NewEncoder(os.Stdout).Encode(firstParagraph("# Heading\\n\\none\\ntwo\\n\\nnext\\n")) }\n'; (scratch/'observer.go').write_text(observer); (p/'observer.go.txt').write_text(observer); records=[]
for id in ['', 'M10']:
 env=os.environ.copy();env['ADAMIC_MUTANT']=id;command=['timeout','90','go','run',str(scratch/'implementation.go'),str(scratch/'coverage.go'),str(scratch/'observer.go')];start=time.monotonic();log=p/('survivor-'+('after' if id else 'before')+'.log')
 with log.open('w') as out:result=subprocess.run(command,env=env,stdout=out,stderr=subprocess.STDOUT)
 records.append(dict(selector=id,command='ADAMIC_MUTANT='+shlex.quote(id)+' '+shlex.join(command)+' > '+str(log)+' 2>&1',exit=result.returncode,wall_seconds=time.monotonic()-start));print(id or 'baseline',result.returncode,log.read_text().strip(),flush=True)
 if result.returncode:raise RuntimeError('observer failed')
(p/'survivor-observation.json').write_text(json.dumps(records,indent=2)+'\n')
