#!/usr/bin/env python3
"""Measure the five already validated builds on the committed benchmark corpus."""
import argparse
import gzip
import hashlib
import json
import os
from pathlib import Path
import shlex
import subprocess


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--work',type=Path,required=True)
    parser.add_argument('--hyperfine',required=True)
    parser.add_argument('--valgrind',required=True)
    parser.add_argument('--valgrind-lib',required=True)
    args=parser.parse_args();work=args.work.resolve()
    os.environ['GOMAXPROCS']='1'
    os.environ['VALGRIND_LIB']=args.valgrind_lib
    modes=['o2','thin','profile','go-plain','go-profile']
    builds=json.loads((work/'builds.json').read_text())
    commands={m:['taskset','-c','3',str(work/m),'--manifest',str(work/'compiler.txt'),'--count'] for m in modes}
    def check(command,stem):
        with (work/(stem+'.stdout')).open('wb') as out,(work/(stem+'.stderr')).open('wb') as err:
            subprocess.run(command,stdout=out,stderr=err,check=True)
        if (work/(stem+'.stdout')).read_bytes()!=b'0\n' or (work/(stem+'.stderr')).read_bytes():
            raise RuntimeError('MISCOMPILE: wrong output: '+stem)
    for mode in modes:
        if hashlib.sha256((work/mode).read_bytes()).hexdigest()!=builds[mode]['binary_sha256']:
            raise RuntimeError('validated binary changed: '+mode)
        check(commands[mode],'warm-'+mode)
    rounds=[]
    for i in range(10):
        base=modes if i<5 else modes[::-1]
        order=base[i%5:]+base[:i%5]
        row={'round':i+1,'order':order}
        for mode in order:
            stem=work/('timing-'+mode+'-'+str(i+1))
            load=os.getloadavg()
            with Path(str(stem)+'.log').open('wb') as log:
                subprocess.run([args.hyperfine,'--runs','1','--warmup','0','--shell','none','--show-output','--export-json',str(stem)+'.json',shlex.join(commands[mode])],stdout=log,stderr=subprocess.STDOUT,check=True)
            body=Path(str(stem)+'.log').read_text().split('\n',1)[1].split('  Time (',1)[0]
            if body!='0\n':raise RuntimeError('MISCOMPILE: timed output differs: '+mode)
            result=json.loads(Path(str(stem)+'.json').read_text())['results'][0]
            row[mode]={'wall':result['times'][0],'user':result['user'],'system':result['system'],'load_before':load,'load_after':os.getloadavg()}
        rounds.append(row)
        (work/'timings.json').write_text(json.dumps({'rounds':rounds,'best':{m:{k:min(r[m][k] for r in rounds) for k in ['wall','user','system']} for m in modes}},indent=2)+'\n')
        print('round',json.dumps(row),flush=True)
    instructions={}
    for mode in modes:
        profile=work/(mode+'.callgrind')
        check(['taskset','-c','3',args.valgrind,'--tool=callgrind','--cache-sim=yes','--branch-sim=yes','--I1=32768,8,64','--D1=32768,8,64','--LL=268435456,1,64','--callgrind-out-file='+str(profile),str(work/mode),'--manifest',str(work/'compiler.txt'),'--count'],'instructions-'+mode)
        events=[];summary=[];footer=[]
        for line in profile.read_text().splitlines():
            if line.startswith('events:'):events=line.split()[1:]
            elif line.startswith('summary:'):summary=list(map(int,line.split()[1:]))
            elif line.startswith('totals:'):footer=list(map(int,line.split()[1:]))
        if len(events)!=13 or len(summary)!=13 or len(footer)!=13 or summary[0]-footer[0] not in [0,2] or any(a!=b for a,b in zip(summary[1:],footer[1:])):
            raise RuntimeError('instruction summary/footer accounting differs: '+mode)
        instructions[mode]={'instructions':summary[0],'self_total':footer[0],'I1_misses':summary[events.index('I1mr')],'events':dict(zip(events,summary)),'binary_sha256':builds[mode]['binary_sha256']}
        with gzip.open(str(profile)+'.gz','wb') as zipped:zipped.write(profile.read_bytes())
        (work/'instructions.json').write_text(json.dumps(instructions,indent=2)+'\n')
        print(mode,'instructions',summary[0],flush=True)
    for mode in modes:
        if hashlib.sha256((work/mode).read_bytes()).hexdigest()!=builds[mode]['binary_sha256']:raise RuntimeError('measured binary changed: '+mode)
    final=json.loads((work/'timings.json').read_text())
    print('stage 1 bar',instructions['profile']['instructions']<=5000000000 and final['best']['profile']['user']<=0.6,flush=True)

if __name__=='__main__':main()
