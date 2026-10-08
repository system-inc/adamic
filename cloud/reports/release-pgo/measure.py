#!/usr/bin/env python3
"""Interleave ten pinned rounds, then count separate I1/D1/branch simulation events."""
import gzip
import json
import os
from pathlib import Path
import shlex
import subprocess
import sys

s = Path(sys.argv[1]).resolve()
tools = Path(sys.argv[2]).resolve()
modes = ['o2', 'thin', 'pgo', 'split']
os.environ['GOMAXPROCS'] = '1'
commands = {mode: ['taskset', '-c', '3', str(s/mode/'parse'), '--manifest', str(s/'held-out.txt'), '--count'] for mode in modes}
def run(args, stem):
    with (s/(stem+'.stdout')).open('wb') as out, (s/(stem+'.stderr')).open('wb') as err:
        subprocess.run(args, stdout=out, stderr=err, check=True)
    if (s/(stem+'.stdout')).read_bytes() != b'0\n':
        raise RuntimeError('MISCOMPILE: wrong parse output: '+stem)
for mode in modes:
    run(commands[mode], 'warm-'+mode)
    if (s/('warm-'+mode+'.stderr')).read_bytes():
        raise RuntimeError('MISCOMPILE: unexpected stderr')
rounds = []
for i in range(10):
    base = modes if (i//4)%2 == 0 else modes[::-1]
    order = base[i%4:] + base[:i%4]
    row = {'round': i+1, 'order': order}
    for mode in order:
        stem = s / ('timing-'+mode+'-'+str(i+1))
        load = os.getloadavg()
        with Path(str(stem)+'.log').open('wb') as log:
            subprocess.run([str(tools/'tools/usr/bin/hyperfine'), '--runs','1','--warmup','0','--shell','none','--show-output','--export-json',str(stem)+'.json',shlex.join(commands[mode])], stdout=log, stderr=subprocess.STDOUT, check=True)
        body = Path(str(stem)+'.log').read_text().split('\n',1)[1].split('  Time (',1)[0]
        if body != '0\n':
            raise RuntimeError('MISCOMPILE: timed output differs')
        r = json.loads(Path(str(stem)+'.json').read_text())['results'][0]
        row[mode] = dict(wall=r['times'][0],user=r['user'],system=r['system'],load_before=load,load_after=os.getloadavg())
    rounds.append(row)
    print('round', json.dumps(row), flush=True)
    (s/'timings.json').write_text(json.dumps(dict(rounds=rounds,best={mode:{metric:min(r[mode][metric] for r in rounds) for metric in ['wall','user','system']} for mode in modes}),indent=2)+'\n')
os.environ['VALGRIND_LIB'] = str(tools/'valgrind/usr/libexec/valgrind')
profiles = {}
def account(path):
    events=[];summary=[];footer=[];total=[];pending=False;function=False
    for line in path.read_text().splitlines():
        if line.startswith('events:'):
            events=line.split()[1:]; total=[0]*len(events)
        elif line.startswith('summary:'):summary=list(map(int,line.split()[1:]))
        elif line.startswith('totals:'):footer=list(map(int,line.split()[1:]))
        elif line.startswith('fn='):function=True
        elif line.startswith('calls='):pending=True
        elif function and line and line[0] in '0123456789+-*':
            values=list(map(int,line.split()[1:]))
            values += [0]*(len(events)-len(values))
            if pending:pending=False;continue
            if len(values)!=len(events):raise RuntimeError('event vector length differs')
            total=[a+b for a,b in zip(total,values)]
    def reconcile(expected):
        if total!=expected:raise RuntimeError('simulated event accounting differs')
    reconcile(footer)
    def reconcile_summary(expected):
        delta=[a-b for a,b in zip(expected,footer)]
        if delta[events.index('Ir')] not in [0,2] or any(v for i,v in enumerate(delta) if events[i]!='Ir'):
            raise RuntimeError('unexpected summary/self accounting difference')
    reconcile_summary(summary)
    summary_mutant=summary.copy();summary_mutant[events.index('Ir')]+=1
    try:reconcile_summary(summary_mutant)
    except RuntimeError:print('Ir summary +1 accounting mutant caught',flush=True)
    else:raise RuntimeError('summary accounting mutant survived')
    mutant=footer.copy();mutant[events.index('I1mr')]+=1
    try:reconcile(mutant)
    except RuntimeError:print('I1 event +1 accounting mutant caught',flush=True)
    else:raise RuntimeError('accounting mutant survived')
    return dict(events=events,summary=dict(zip(events,summary)),self_totals=dict(zip(events,footer)),summary_minus_self=dict(zip(events,[a-b for a,b in zip(summary,footer)])))
for mode in modes:
    path=s/(mode+'.callgrind')
    with (s/(mode+'-simulation.stdout')).open('wb') as out, (s/(mode+'-simulation.stderr')).open('wb') as err:
        subprocess.run(['taskset','-c','3',str(tools/'valgrind/usr/bin/valgrind'),'--tool=callgrind','--cache-sim=yes','--branch-sim=yes','--I1=32768,8,64','--D1=32768,8,64','--LL=268435456,1,64','--callgrind-out-file='+str(path),str(s/mode/'parse'),'--manifest',str(s/'held-out.txt'),'--count'],stdout=out,stderr=err,check=True)
    if (s/(mode+'-simulation.stdout')).read_bytes()!=b'0\n':
        raise RuntimeError('MISCOMPILE: simulated-run output differs')
    profiles[mode]=account(path)
    with gzip.open(str(path)+'.gz','wb') as z:z.write(path.read_bytes())
    (s/'profiles.json').write_text(json.dumps(profiles,indent=2)+'\n')
    print(mode,json.dumps(profiles[mode]),flush=True)
