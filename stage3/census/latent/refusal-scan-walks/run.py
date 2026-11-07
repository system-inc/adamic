"""Run scan-only census profiles sequentially on the same box, after building probes."""
import json, os, pathlib, subprocess, sys, time
repo, scratch, adapted, evidence = map(lambda p: pathlib.Path(p).resolve(), sys.argv[1:])
evidence.mkdir(parents=True, exist_ok=True)
probe = 'TestTask2bz5m7vRefusalScanProfile'
for mode in ('plain', 'counts'):
    for name in ('utilities', 'checker'):
        prefix = evidence / (mode + '-' + name)
        env = dict(os.environ, SCAN_ENTRY=str(adapted/'src/tsc/tsc.ts'),
                   SCAN_FILE=str(adapted/'src/compiler'/ (name+'.ts')),
                   SCAN_OUTPUT=str(prefix)+'.json', SCAN_COUNTS='1' if mode=='counts' else '0')
        for key in ('LATENT_FULL','LATENT_SPECULATIVE','LATENT_PROGRESS_FILE','LATENT_ONLY_FILE'):
            env.pop(key, None)
        if mode=='plain': env['SCAN_PPROF']=str(prefix)+'.pprof'
        else: env.pop('SCAN_PPROF',None)
        with open(str(prefix)+'.log','w') as log:
            start=time.monotonic()
            result=subprocess.run([str(scratch/('2bz-'+mode+'.test')), '-test.run=^'+probe+'$',
                                   '-test.count=1','-test.timeout=20m'],cwd=repo,env=env,
                                   stdout=log,stderr=subprocess.STDOUT,timeout=1210)
            log.write(f'\nprocess_wall_seconds={time.monotonic()-start:.6f}\n')
            if result.returncode: raise SystemExit(result.returncode)
        if mode=='plain':
            with open(str(prefix)+'-top.txt','w') as log:
                subprocess.run(['go','tool','pprof','-top','-cum','-nodecount=100',
                                str(scratch/'2bz-plain.test'),str(prefix)+'.pprof'],
                               cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
