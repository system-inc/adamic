import pathlib,json,difflib,subprocess,time,os
repo=pathlib.Path('/workspace/adamic');r=repo/'review/test-audit/stage3-fixtures-fixtures'
# Standalone Lower entry probe. Remove now-unused imports and the entire body.
p=repo/'internal/lower/lower.go';s=p.read_text();a=s.index('func Lower(');start=s.index('{',a);end=s.index('\ntype lowering struct');v=s[:start+1]+'\n return nil, nil\n}\n'+s[end:];v=v.replace('\n\t"fmt"','').replace('\n\t"path/filepath"','');(r/'P1.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),v.splitlines(True),fromfile='a/internal/lower/lower.go',tofile='b/internal/lower/lower.go')));p.write_text(v)
try:
 with (r/'logs/P1-vet.log').open('w') as f:rc=subprocess.run(['go','vet','./internal/lower/'],stdout=f,stderr=subprocess.STDOUT).returncode
 if rc:raise RuntimeError('empty lower vet')
finally:p.write_text(s)
# Native C entry, zero string at entry. All other emitter functions stay intact.
p=repo/'internal/native/emit.go';s=p.read_text();v=s.replace('return cProgram(program, -1)','return ""',1);(r/'P2.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),v.splitlines(True),fromfile='a/internal/native/emit.go',tofile='b/internal/native/emit.go')));p.write_text(v)
try:
 with (r/'logs/P2-vet.log').open('w') as f:rc=subprocess.run(['go','vet','./internal/native/'],stdout=f,stderr=subprocess.STDOUT).returncode
 if rc:raise RuntimeError('empty native vet')
 env=os.environ.copy();env.update(ADAMIC_BUILD_CACHE_DIR='/tmp/u163/cache/P2',ADAMIC_STAGE3_BUILD_STORE='/tmp/u163/hooks/P2');cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage3/fixtures/','-run','.'];t=time.monotonic()
 with (r/'logs/P2.log').open('w') as f:rc=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT).returncode
 (r/'native-probe-run.json').write_text(json.dumps(dict(id='P2',command=cmd,wall=time.monotonic()-t,rc=rc),indent=2))
finally:p.write_text(s)
