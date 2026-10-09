import pathlib,subprocess,time,json,difflib,os
root=pathlib.Path('/workspace/adamic');os.chdir(root);p=root/'review/test-defend/internal-lower-enum_flags';plan=json.loads((p/'defense-plan.json').read_text());m=next(x for x in plan if x['id']=='D3');s=root/m['file'];base=s.read_text();assert base==subprocess.check_output(['git','show','HEAD:'+m['file']],text=True)
for suffix in ['.log','.diff']:
 (p/('D3'+suffix)).rename(p/('D3-inactive'+suffix))
m['new']=m['new'].replace('strings.HasPrefix(initializer.Text(), "0x")','strings.HasPrefix(sourceExpression(initializer), "0x")');m['selector_repair']='initializer.Text normalizes radix spelling; corrected to existing sourceExpression helper.';plan[2]=m
(p/'defense-plan.json').write_text(json.dumps(plan,indent=2));changed=base.replace(m['old'],m['new']);(p/'D3.diff').write_text(''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
runs=json.loads((p/'matrix-runs.json').read_text());runs['D3-inactive']=runs.pop('D3')
def run(n,c,env=None):
 t=time.monotonic()
 with (p/(n+'.log')).open('w') as f:r=subprocess.run(c,stdout=f,stderr=subprocess.STDOUT,env=env)
 runs[n]={'exit':r.returncode,'wall_seconds':time.monotonic()-t,'command':c};return r.returncode
try:
 assert run('D3-apply',['git','apply','--check',str(p/'D3.diff')])==0
 s.write_text(changed);assert run('D3-vet',['go','vet','./internal/lower/'])==0
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/enum-defend/cache/D3-source-spelling'
 run('D3',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],env)
finally:s.write_text(base);(p/'matrix-runs.json').write_text(json.dumps(runs,indent=2))
print('D3 corrected matrix complete')
