#!/usr/bin/env python3
"""Cold means the checker archive is absent; the Go build cache is warm."""
import json, os, pathlib, shutil, subprocess, time
root=pathlib.Path(__file__).resolve().parents[3]
reports=pathlib.Path(__file__).resolve().parent
scratch=pathlib.Path('/workspace/archive-once-cost')
rows=[]
for loop in range(1,4):
    slow=False
    for mode,flag in [('without','--gate-inputs-no-archive'),('with','--gate-inputs')]:
        cache=pathlib.Path('/home/agent/.cache/go-build')
        checker=pathlib.Path('/workspace/css-gate-tools/gate-inputs/checker')
        if checker.exists():shutil.rmtree(checker)
        for state in ['cold','warm']:
            env=os.environ.copy();env.update(ADAMIC_TOOLS='/workspace/css-gate-tools',GOMODCACHE='/tmp/adamic-gate/setup-modules-proof/with',GOCACHE=str(cache),GOFLAGS='-trimpath',XDG_CACHE_HOME='/tmp/adamic-gate/yaml-input-user-cache',ADAMIC_GATE_UNCACHED='0')
            log=reports/f'{loop}-{mode}-{state}.log'
            before=pathlib.Path('/proc/loadavg').read_text().strip()
            started=time.monotonic()
            with log.open('w') as out: result=subprocess.run(['bash','cloud/setup.sh',flag],cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
            elapsed=time.monotonic()-started
            row=dict(loop=loop,mode=mode,state=state,seconds=elapsed,returncode=result.returncode,load_before=before,load_after=pathlib.Path('/proc/loadavg').read_text().strip(),instrument=f'ADAMIC_TOOLS={env["ADAMIC_TOOLS"]} GOMODCACHE={env["GOMODCACHE"]} GOCACHE={cache} GOFLAGS=-trimpath XDG_CACHE_HOME={env["XDG_CACHE_HOME"]} ADAMIC_GATE_UNCACHED=0 bash cloud/setup.sh {flag}',log=log.name)
            rows.append(row);(reports/'timings.json').write_text(json.dumps(rows,indent=2)+'\n')
            print(json.dumps(row),flush=True)
            if result.returncode:raise SystemExit(result.returncode)
            slow=slow or elapsed>300

    if slow:break
