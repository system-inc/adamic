#!/usr/bin/env python3
"""Run the unchanged unified lint gate; preserve its events and upstream capture."""
import collections, glob, hashlib, json, os, pathlib, subprocess, sys, time
repository = pathlib.Path.cwd()
output = pathlib.Path(sys.argv[1]).resolve()
output.mkdir(parents=True, exist_ok=True)
profile = pathlib.Path(sys.argv[2]).resolve()
profile.mkdir(parents=True, exist_ok=True)
environment = dict(os.environ, ADAMIC_LINT_BENCH='1', ADAMIC_TYPESCRIPT_SOURCE='/workspace/wave-06-typescript', ADAMIC_LINT_PROFILE_DIR=str(profile), ADAMIC_LINT_PROFILE_SNAPSHOTS=str(profile), GOFLAGS='-buildvcs=false', GOPROXY='https://proxy.golang.org|direct')
command = ['go', 'test', '-json', '-count=1', '-timeout=60m', './stage1/cohere/lint']
started = time.monotonic()
loads = []
records = {}
metadata = {'command':command,'branch':subprocess.check_output(['git','branch','--show-current'],text=True).strip(),'head':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'nproc':subprocess.check_output(['nproc'],text=True).strip(),'cpu.max':pathlib.Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'inputs':{k:v for k,v in environment.items() if k.startswith('ADAMIC_') or k in ('GOMAXPROCS','GOFLAGS','GOPROXY')},'loadStart':list(os.getloadavg())}
(output/'command.json').write_text(json.dumps(metadata,indent=2)+'\n')
with (output/'gate.jsonl').open('wb') as log:
    process = subprocess.Popen(command, env=environment, stdout=log, stderr=subprocess.STDOUT)
    while process.poll() is None:
        loads.append([round(time.monotonic()-started,3),*os.getloadavg()])
        for path in glob.glob(str(pathlib.Path(environment.get('TMPDIR', '/tmp'))/'lint-shared-*'/'upstream-*'/'capture'/'*.jsonl')):
            try:
                for line in pathlib.Path(path).read_text().splitlines():
                    row = json.loads(line)
                    if (repository/'stage1/cohere/lint/rules/no-duplicate-type-constituents/rule.json').exists() and row.get('rule') == '@typescript-eslint/no-duplicate-type-constituents':
                        records[line] = row
            except (OSError, json.JSONDecodeError): pass
        if records: (output/'upstream-capture.json').write_text(json.dumps(list(records.values()),indent=2)+'\n')
        time.sleep(2)
status = process.returncode
counts = collections.Counter(); top = collections.Counter(); skips=[]; failures=[]; mutants=[]; typed=[]
for line in (output/'gate.jsonl').read_text().splitlines():
    try: event=json.loads(line)
    except json.JSONDecodeError: continue
    action=event.get('Action'); name=event.get('Test',''); message=event.get('Output','')
    if name and action in ('pass','fail','skip'):
        counts[action]+=1
        if '/' not in name: top[action]+=1
        if action=='skip': skips.append(name)
        if action=='fail': failures.append(name)
    if 'duplicate text verdict suppressed caught on' in message: mutants.append(message.strip())
    if name=='TestRulesAgree' and ('typed runtime time:' in message or 'live Go, native, Node and emitted JavaScript replay identical:' in message): typed.append(message.strip())
unique={(r.get('rule'),r.get('file'),r.get('source'),json.dumps(r.get('options'),sort_keys=True)) for r in records.values()}
metadata.update(exitCode=status,wallSeconds=time.monotonic()-started,counts=dict(counts),topLevel=dict(top),skips=skips,failures=failures,loadEnd=list(os.getloadavg()),upstreamUnique=len(unique),mutants=mutants)
(output/'summary.json').write_text(json.dumps(metadata,indent=2)+'\n')
(output/'load.json').write_text(json.dumps(loads)+'\n')
(output/'typed-runtime.log').write_text('\n'.join(typed)+'\n')
(output/'mutant.log').write_text('\n'.join(mutants)+'\n')
print(json.dumps(metadata))
sys.exit(status)
