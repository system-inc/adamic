import pathlib,json,subprocess,time,os
p=pathlib.Path('review/test-defend/internal-lower-module_namespace');s=json.loads((p/'scope.json').read_text());m=next(r for r in json.loads((p/'mutant-plan.json').read_text()) if r['mutant']=='D06');f=pathlib.Path(m['file']);original=f.read_text();assert original==subprocess.check_output(['git','show',s['base']+':'+str(f)],text=True)
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^TestEnumInitializationReach$'];t=time.monotonic()
try:
 f.write_text(original.replace(m['old'],m['new'],1))
 with (p/'logs/D06-enum-whole.log').open('w') as log:code=subprocess.call(cmd,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-module-namespace/cache/D06'),stdout=log,stderr=log)
 (p/'D06-enum-whole.json').write_text(json.dumps({'command':cmd,'env':{'ADAMIC_BUILD_CACHE_DIR':'/tmp/defend-module-namespace/cache/D06'},'exit':code,'wall_seconds':time.monotonic()-t},indent=2))
finally:f.write_text(original)
