from pathlib import Path
import subprocess,json,os,time,difflib
root=Path.cwd(); ev=root/'review/test-audit/cloud'; src=root/'cloud/lint-wave-check.py'
base=subprocess.check_output(['git','show','HEAD:cloud/lint-wave-check.py'],text=True)
changes=[('M1',73,'require(not status, "evidence"','require(status, "evidence"','flip condition'),('M2',108,'self.git("fetch", "--prune", "--no-recurse-submodules", "origin",','self.git("fetch", "--no-recurse-submodules", "origin",','change option'),('M3',338,'for backend in ("Node", "native"):','for backend in ("Node",):','change constant'),('P1',405,'def main():\n','def main():\n    return 0\n','empty-answer probe')]
(ev/'diffs').mkdir(exist_ok=True)
plan=[];validations=[]
for mid,line,old,new,kind in changes:
    assert base.count(old)==1
    altered=base.replace(old,new)
    diff=''.join(difflib.unified_diff(base.splitlines(True),altered.splitlines(True),fromfile='a/cloud/lint-wave-check.py',tofile='b/cloud/lint-wave-check.py'))
    path=ev/'diffs'/f'{mid}.diff';path.write_text(diff)
    plan.append({'id':mid,'file':'cloud/lint-wave-check.py','line':line,'old':old,'new':new,'kind':kind})
    start=time.monotonic()
    with (ev/f'{mid}-compile.log').open('w') as out:
        for cmd in (['git','apply','--check',str(path)],['git','apply',str(path)],['python3','-B','-c',"from pathlib import Path; compile(Path('cloud/lint-wave-check.py').read_text(), 'cloud/lint-wave-check.py', 'exec'); print('Python compile: PASS')"]):
            r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT);assert r.returncode==0,cmd
    subprocess.run(['git','restore','--source=HEAD','--','cloud/lint-wave-check.py'],check=True)
    validations.append({'id':mid,'exit':0,'seconds':time.monotonic()-start})
(ev/'plan.json').write_text(json.dumps(plan,indent=2));(ev/'compile-status.json').write_text(json.dumps(validations,indent=2))
# Install a single selector in the production source only.
switched=base
switched=switched.replace(changes[0][2],'require(status if os.getenv("ADAMIC_MUTANT") == "M1" else not status, "evidence"')
switched=switched.replace(changes[1][2],'self.git("fetch", *(("--prune",) if os.getenv("ADAMIC_MUTANT") != "M2" else ()), "--no-recurse-submodules", "origin",')
switched=switched.replace(changes[2][2],'for backend in (("Node",) if os.getenv("ADAMIC_MUTANT") == "M3" else ("Node", "native")):')
switched=switched.replace('def main():\n','def main():\n    if os.getenv("ADAMIC_MUTANT") == "P1":\n        return 0\n')
patch=''.join(difflib.unified_diff(base.splitlines(True),switched.splitlines(True),fromfile='a/cloud/lint-wave-check.py',tofile='b/cloud/lint-wave-check.py'))
(ev/'selector.diff').write_text(patch)
subprocess.run(['git','apply',str(ev/'selector.diff')],check=True)
statuses=[]
try:
    for mid in ['control','M1','M2','M3','P1']:
        env=dict(os.environ);env.pop('ADAMIC_MUTANT',None)
        if mid!='control':env['ADAMIC_MUTANT']=mid
        cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./cloud/','-run','.']
        start=time.monotonic()
        with (ev/f'{mid}.log').open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
        status={'id':mid,'exit':r.returncode,'wall':time.monotonic()-start,'command':('ADAMIC_MUTANT='+mid+' ' if mid!='control' else '')+' '.join(cmd)+' > '+str(ev/f'{mid}.log')+' 2>&1'}
        statuses.append(status);(ev/'matrix-status.json').write_text(json.dumps(statuses,indent=2)); print(json.dumps(status),flush=True)
finally:
    subprocess.run(['git','restore','--source=HEAD','--','cloud/lint-wave-check.py'],check=True)
