import pathlib,subprocess,time,json,difflib,os
p=pathlib.Path('/tmp/u123/evidence');f=pathlib.Path('internal/lower/lower.go');old=f.read_text();header='func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {';new=old.replace(header,header+'\n\tif true { return nil, nil }',1)
(p/'P1.diff').write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+str(f),tofile='b/'+str(f))))
try:
 f.write_text(new)
 start=time.monotonic()
 with (p/'P1-vet.log').open('w') as log:r=subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT)
 result={'vet_exit':r.returncode,'vet_seconds':time.monotonic()-start}
 start=time.monotonic();env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u123/cache/P1')
 with (p/'P1-matrix.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/rules/no-unsafe-optional-chaining/','-run','.'],stdout=log,stderr=subprocess.STDOUT,env=env)
 result.update(exit=r.returncode,wall_seconds=time.monotonic()-start);(p/'probe-timings.json').write_text(json.dumps(result,indent=2));print(result,flush=True)
finally:f.write_text(old)
