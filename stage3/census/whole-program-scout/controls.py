"""Run each fixture in a fresh no-output process; require scan continuation."""
import json, os, pathlib, subprocess, sys
here=pathlib.Path(__file__).resolve().parent
binary=pathlib.Path(sys.argv[1]).resolve()
out=pathlib.Path(sys.argv[2]).resolve(); out.mkdir(parents=True,exist_ok=True)
rows=[]
for name in ('bind-stop','bind-clean','record-stop','record-clean'):
    source=here/'fixtures'/(name+'.ts')
    for mutant in (False,True):
        label=name+('-first-error-mutant' if mutant else '')
        env=dict(os.environ)
        if mutant: env['SCOUT_MUTANT_FIRST_ERROR']='1'
        run=subprocess.run([sys.executable,str(here/'measure.py'),str(out/label),str(binary),str(source)],env=env,text=True,capture_output=True)
        assert run.returncode==0,(label,run.returncode,run.stderr)
        result=json.loads((out/(label+'.json')).read_text())
        words=[s['Words'] for s in result['Stops'] or []]
        assert any('@ts-check' in w for w in words),(label,result)
        target='bind of a function that declares' if name.startswith('bind') else 'a record written into a graph slot'
        has=any(target in w for w in words)
        assert has==(name.endswith('-stop') and not mutant),(label,result)
        if not mutant and name.endswith('-clean'): assert result['Completed'],result
        if mutant: assert not result['Completed'],result
        rows.append(dict(name=label,target_stop=has,completed=result['Completed'],stops=len(words)))
(out/'controls-summary.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))
