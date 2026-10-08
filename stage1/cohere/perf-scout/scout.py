#!/usr/bin/env python3
"""Step 44: prepare immutable snapshots, verify full bytes, then measure cold processes."""
import argparse
from concurrent.futures import ThreadPoolExecutor
import hashlib
import fcntl
import importlib.util
import json
import os
from pathlib import Path
import platform
import re
import statistics
import subprocess
import sys
import time

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
PIN = '050880ce59e30b356b686bd3144efe24f875ebc8'

USAGE_HELPER = None

def run(argv, output, env=None, cwd=REPO):
    output = Path(output)
    output.parent.mkdir(parents=True, exist_ok=True)
    with output.open('wb') as out, Path(str(output)+'.stderr').open('wb') as err:
        start = time.perf_counter()
        command=list(map(str,argv))
        if USAGE_HELPER is not None:command=[str(USAGE_HELPER),str(output)+'.usage',*command]
        process = subprocess.Popen(command, stdout=out, stderr=err, cwd=cwd, env=env)
        _, status, usage = os.wait4(process.pid, 0)
        process.returncode = os.waitstatus_to_exitcode(status)
        seconds = time.perf_counter()-start
    if process.returncode:
        raise RuntimeError(f'{argv}: exit {process.returncode}; see {output}.stderr')
    if USAGE_HELPER is not None:
        rss,user,system=Path(str(output)+'.usage').read_text().split()
    else:rss,user,system=usage.ru_maxrss,usage.ru_utime,usage.ru_stime
    return dict(seconds=seconds, max_rss_kib=int(rss), user=float(user), system=float(system),
                sha256=hashlib.sha256(output.read_bytes()).hexdigest(), bytes=output.stat().st_size)

def checked(actual, expected):
    if actual != expected:
        raise RuntimeError(f'byte mismatch: actual {hashlib.sha256(actual).hexdigest()}, expected {hashlib.sha256(expected).hexdigest()}')

def balanced_counts(text):
    match=re.fullmatch(r'adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\s*',text)
    if not match: raise ValueError('missing allocation counters')
    allocations,frees,retains,releases,peak,regions=map(int,match.groups())
    if allocations != frees or peak <= 0 or regions != 0:
        raise ValueError('unbalanced allocations, missing peak or unexpected regions')
    return dict(allocations=allocations,frees=frees,retains=retains,releases=releases,peak=peak,regions=regions)

def rows(path):
    values = [x for x in Path(path).read_text().splitlines() if x]
    if not values:
        raise ValueError(f'empty input: {path}')
    return values

def syntax_names(root):
    descriptors=[json.loads(p.read_text()) for p in sorted((root/'rules').glob('*/rule.json'))]
    if not descriptors:
        raise ValueError('no registered rules')
    return [d['name'] for d in descriptors if not d.get('typed')], descriptors

def machine():
    return dict(uname=platform.uname()._asdict(), cores=os.cpu_count(), affinity=sorted(os.sched_getaffinity(0)),
                quota=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
                cpu=next(x for x in Path('/proc/cpuinfo').read_text().splitlines() if x.startswith('model name')),
                load_1m=Path('/proc/loadavg').read_text().split()[0])

def overlay_build(directory, source, name):
    virtual=REPO/'cohere'/('adamic_scout_'+name+'.go')
    overlay=directory/(name+'-overlay.json')
    overlay.write_text(json.dumps({'Replace':{str(virtual):str(source)}}))
    run(['go','build','-overlay='+str(overlay),'-o',directory/name,virtual],directory/(name+'-build.log'),cwd=REPO/'cohere')

def prepare(directory, corpus):
    if subprocess.check_output(['git','-C',str(corpus),'rev-parse','HEAD'],text=True).strip()!=PIN:
        raise ValueError('wrong TypeScript pin')
    directory.mkdir(parents=True,exist_ok=True)
    run(['clang','-std=c11','-O2','-Wall','-Wextra','-Werror',HERE/'testdata/usage.c','-o',directory/'usage'],directory/'usage-build.log')
    digest=hashlib.sha256()
    for root in (REPO/'stage1/cohere/lint', REPO/'stage1/typescript', REPO/'stage1/cohere/tsprinter', REPO/'internal/native/runtime'):
        for path in sorted(root.rglob('*')):
            if path.is_file() and path.suffix in ('.ts','.a','.json','.c','.h') and 'validation' not in path.parts and 'performance' not in path.parts:
                digest.update(str(path.relative_to(REPO)).encode());digest.update(path.read_bytes())
    source_key=digest.hexdigest()
    keypath=directory/'source-key.txt'
    if keypath.exists() and keypath.read_text().strip()!=source_key:raise ValueError('source changed: prepare in a fresh directory')
    env=os.environ.copy();env.update(ADAMIC_SCOUT_OUTPUT=str(directory/'lint'),ADAMIC_TYPESCRIPT_SOURCE=str(corpus))
    replacement=HERE/'testdata/artifacts_test.go.txt'
    overlay=directory/'lint-overlay.json'
    overlay.write_text(json.dumps({'Replace':{str(REPO/'stage1/cohere/lint/scout_artifacts_test.go'):str(replacement)}}))
    if not (directory/'lint/upstream.txt').exists():
        run(['go','test','-count=1','-v','-timeout=30m','-overlay='+str(overlay),'-run=^TestScoutArtifacts$','./stage1/cohere/lint'],directory/'lint-build.log',env=env)
    # Preserve the real oracle, adding only an opt-in profiler around its entry point.
    oracle=(REPO/'stage1/cohere/lint/testdata/oracle.go').read_text()
    oracle=oracle.replace('import (','import (\n "runtime/pprof"',1).replace('func main() {','''func main() {
 if path:=os.Getenv("SCOUT_PPROF");path!="" {
 f,err:=os.Create(path);if err!=nil{panic(err)}
 if err:=pprof.StartCPUProfile(f);err!=nil{panic(err)}
 defer f.Close();defer pprof.StopCPUProfile()
 }
''',1)
    (directory/'lint-oracle.go').write_text(oracle)
    # Registry oracle adapters retain Go's internal import scope through an overlay.
    lint=directory/'lint'
    adapters=[lint/'.generated/registry.go', *sorted((lint/'rules').glob('*/oracle.go'))]
    replacements={str(REPO/'cohere/adamic_scout_lint.go'):str(directory/'lint-oracle.go')}
    for i,path in enumerate(adapters):replacements[str(REPO/f'cohere/adamic_scout_adapter_{i}.go')]=str(path)
    (directory/'oracle-overlay.json').write_text(json.dumps({'Replace':replacements}))
    run(['go','build','-overlay='+str(directory/'oracle-overlay.json'),'-o',lint/'oracle-profile',*replacements],directory/'oracle-profile-build.log',cwd=REPO/'cohere')
    names,descriptors=syntax_names(lint)
    (directory/'registered.json').write_text(json.dumps(descriptors,indent=2)+'\n')
    # Every compiler file, every owned syntax-rule fixture, every upstream captured syntax case.
    manifests={ 'compiler':[str(p)+'\t'+name for p in sorted((corpus/'src/compiler').rglob('*.ts')) for name in names],
                'owned':[], 'upstream':[] }
    for p in sorted((lint/'rules').glob('*/testdata/**/*')):
        if not p.is_file() or not p.name.endswith(('.ts.txt','.tsx.txt','.js.txt','.jsx.txt')):continue
        slug=p.relative_to(lint/'rules').parts[0]
        d=json.loads((lint/'rules'/slug/'rule.json').read_text())
        if d.get('typed'):continue
        relative=p.relative_to(lint/'rules'/slug/'testdata')
        target=directory/'owned-fixtures'/slug/str(relative)[:-4]
        target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(p.read_bytes())
        row=str(target)+'\t'+d['name']
        sidecar=REPO/'stage1/cohere/lint/rules'/slug/'testdata'/relative
        sidecar=sidecar.with_name(sidecar.name.rsplit('.',2)[0]+'.options.json')
        if sidecar.exists():row+='\t\t\tfalse\t'+json.dumps(json.loads(sidecar.read_text()),separators=(',',':'))
        manifests['owned'].append(row)
    for row in rows(lint/'upstream.txt'):
        f=row.split('\t')
        if len(f)>1 and f[1] in names and not row.endswith('unsupported-recovery'):manifests['upstream'].append(row)
    # Isolate shared parse/walk/serialization with one complete syntax rule.
    # The separate fused command measures every registered rule with a live program.
    manifests['compiler']=[str(p)+'\tno-var' for p in sorted((corpus/'src/compiler').rglob('*.ts'))]
    public_root=directory.parent.parent/'public-pins'
    public_files=[]
    for record in json.loads((HERE/'public-pins.json').read_text()):
        checkout=public_root/record['name']
        if not (checkout/'.git').exists(): raise ValueError('missing public checkout '+record['name'])
        if subprocess.check_output(['git','-C',checkout,'rev-parse','HEAD'],text=True).strip()!=record['sha']: raise ValueError('wrong public pin '+record['name'])
        tracked=subprocess.check_output(['git','-C',str(checkout),'ls-files','*.ts'],text=True).splitlines()
        candidates=[checkout/x for x in tracked if not x.endswith('.d.ts') and (checkout/x).is_file() and 200 <= (checkout/x).stat().st_size <= 10000]
        if not candidates: raise ValueError('no eligible .ts sources '+record['name'])
        candidate=sorted(candidates,key=lambda x:(x.stat().st_size,str(x)))[0]
        public_files.append(dict(**record,path=str(candidate.relative_to(checkout)),bytes=candidate.stat().st_size,sha256=hashlib.sha256(candidate.read_bytes()).hexdigest()))
        manifests.setdefault('public',[]).append(str(candidate)+'\tno-var')
    own_files=[p for p in subprocess.check_output(['git','ls-tree','-r','--name-only','9156bf5c579a44d687c9955d13e44f9ad8bbb6f8','--','stage1'],text=True).splitlines() if p.endswith('.ts')]
    manifests['repository']=[str(REPO/x)+'\tno-var' for x in own_files]
    (directory/'public-fixtures.json').write_text(json.dumps(public_files,indent=2)+'\n')
    for key,values in manifests.items():
        if not values:raise ValueError('empty '+key)
        path=directory/(key+'.txt');path.write_text('\n'.join(values)+'\n')
        run([lint/'oracle','--manifest',path,'--diagnostics'],directory/(key+'-diagnostics.txt'))
        flags=rows(directory/(key+'-diagnostics.txt'))
        if len(flags)!=len(values):raise ValueError('diagnostic row count')
        marked=[]
        for row,flag in zip(values,flags):
            f=row.split('\t')
            if flag=='1':
                f+=['']*max(0,7-len(f));f[6]='recovery'
            elif flag!='0':raise ValueError('bad diagnostic marker '+flag)
            marked.append('\t'.join(f))
        path.write_text('\n'.join(marked)+'\n')
    # The TypeScript corpus supplies real supported fragments via the existing independent selector.
    printer=directory/'printer';printer.mkdir(exist_ok=True)
    files=sorted((REPO/'stage3/drivers/tsc/corpus').rglob('*.ts'))+sorted((corpus/'src/compiler').rglob('*.ts'))
    (printer/'request.json').write_text(json.dumps({'Files':list(map(str,files)),'Directory':str(printer)}))
    mapping={str(REPO/'cohere/internal/format/javascript'/('adamic_'+name+'_test.go')):str(REPO/'stage1/cohere/tsprinter/testdata'/(name+'_side_test.go')) for name in ('expressions','statements')}
    (printer/'overlay.json').write_text(json.dumps({'Replace':mapping}))
    env['ADAMIC_TS_STATEMENT_REQUEST']=str(printer/'request.json')
    run(['go','test','-v','-count=1','-overlay='+str(printer/'overlay.json'),'-run=^TestAdamicStatementCorpus$','./internal/format/javascript'],printer/'select.log',env=env,cwd=REPO/'cohere')
    specs=json.loads((printer/'cases.json').read_text())
    real=[s for s in specs if s['Label'].startswith('/')]
    if not real:raise ValueError('no real printer cases')
    def escape(s):return s.replace('\\','\\\\').replace('\n','\\n').replace('\r','\\r').replace('\t','\\t')
    (printer/'cases.txt').write_text(''.join('>'+escape(s['Source'])+'\n' for s in real))
    (printer/'fixtures.json').write_text(json.dumps(real,indent=2)+'\n')
    overlay_build(directory,HERE/'testdata/printer_oracle.go.txt','printer-oracle')
    run(['go','run','./cmd/adamic','build',str(REPO/'stage1/cohere/tsprinter/statementsMain.ts'),'-o',printer/'native'],printer/'native-build.log')
    run(['go','run','./cmd/adamic','build',str(REPO/'stage1/cohere/tsprinter/statementsMain.ts'),'-o',printer/'counted','--count'],printer/'counted-build.log')
    # The native CLI saves C via `c`; use the same release flags plus debug symbols for attribution.
    run(['go','run','./cmd/adamic','c',str(REPO/'stage1/cohere/tsprinter/statementsMain.ts')],printer/'main.c')
    runtime=REPO/'internal/native/runtime'
    flags=['-std=c11','-O2','-g','-ffp-contract=off','-fno-optimize-sibling-calls','-I'+str(runtime)]
    run(['clang',*flags,printer/'main.c',*sorted(runtime.glob('*.c')),'-lm','-o',printer/'profiled'],printer/'profiled-build.log')
    keypath.write_text(source_key+'\n')
    (directory/'build.json').write_text(json.dumps(dict(base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
        cohere=subprocess.check_output(['git','-C','cohere','rev-parse','HEAD'],text=True).strip(),typescript=PIN,
        machine=machine(),native_flags=flags,registered=len(descriptors),syntax_rules=len(names)),indent=2)+'\n')

# Baseline plus a separate existing-batch-API process-start control.
def measure(directory, valgrind, rounds):
    global USAGE_HELPER
    USAGE_HELPER=directory/'usage'
    if not USAGE_HELPER.exists():raise ValueError('build testdata/usage.c as '+str(USAGE_HELPER))
    lint=directory/'lint';printer=directory/'printer'
    summary={'machine_before':machine(),'workloads':{}}
    for workload in ('owned','upstream','compiler','public','repository','printer'):
        isprinter=workload=='printer';input_path=printer/'cases.txt' if isprinter else directory/(workload+'.txt')
        items=rows(input_path)
        args=['--cases',str(input_path),'80'] if isprinter else ['--manifest',str(input_path)]
        commands={'Go':[str(directory/'printer-oracle' if isprinter else lint/'oracle'),*args],
                  'native':[str(printer/'native' if isprinter else lint/'scanner'),*args],
                  'Node':['node','--disable-warning=ExperimentalWarning',str(REPO/'oracle/node.mjs'),str(REPO/'stage1/cohere/tsprinter/statementsMain.ts' if isprinter else lint/'main.ts'),*args]}
        dest=directory/'measure'/workload;dest.mkdir(parents=True,exist_ok=True)
        answers={}
        for name,command in commands.items():
            run(command,dest/(name+'.answer'))
            stderr=Path(str(dest/(name+'.answer'))+'.stderr').read_bytes()
            if stderr:raise RuntimeError(f'{name} stderr: {stderr[:200]!r}')
            answers[name]=(dest/(name+'.answer')).read_bytes()
        checked(answers['native'],answers['Go']);checked(answers['Node'],answers['Go'])
        measurements={name:[] for name in commands}
        for r in range(rounds):
            # Rotate order, no build/profile work occurs between samples.
            names=list(commands);names=names[r%3:]+names[:r%3]
            for name in names:
                result=run(commands[name],dest/f'{r}-{name}')
                checked((dest/f'{r}-{name}').read_bytes(),answers['Go'])
                measurements[name].append(result)
        files=len(set(x.split('\t')[0] for x in items)) if not isprinter else len(set(s['Label'].split(':')[0] for s in json.loads((printer/'fixtures.json').read_text())))
        values=dict(files=files,rows=len(items),bytes=len(answers['Go']),sha256=hashlib.sha256(answers['Go']).hexdigest(),
                    findings=answers['Go'].count(b'\nrange ') if not isprinter else None,
                    samples=measurements,median={name:dict(seconds=statistics.median(x['seconds'] for x in samples),
                    wall_per_row=statistics.median(x['seconds'] for x in samples)/len(items),
                    amortized_wall_per_file=statistics.median(x['seconds'] for x in samples)/files,
                    max_rss_kib=max(x['max_rss_kib'] for x in samples)) for name,samples in measurements.items()})
        # Counts never substitute for the full-byte guard. Counted execution retains findings/fixes.
        run([str(printer/'counted' if isprinter else lint/'counted'),*args],dest/'counted')
        checked((dest/'counted').read_bytes(),answers['Go'])
        values['allocation_counts']=Path(str(dest/'counted')+'.stderr').read_text().strip()
        balanced_counts(values['allocation_counts'])
        # Profile a fixed first-file case set for affordable deterministic per-file evidence.
        if isprinter:
            fixture=json.loads((printer/'fixtures.json').read_text())[0]
            first_file=next((s['Label'].split(':')[0] for s in json.loads((printer/'fixtures.json').read_text()) if '/src/compiler/core.ts:' in s['Label']),fixture['Label'].split(':')[0])
            profitems=[row for row,s in zip(items,json.loads((printer/'fixtures.json').read_text())) if s['Label'].split(':')[0]==first_file]
        else:
            selected=next((row for row in items if row.split('\t')[0].endswith('/scanner.ts')),items[0])
            first_file=selected.split('\t')[0];profitems=[row for row in items if row.split('\t')[0]==first_file]
        profpath=dest/'profile-input.txt';profpath.write_text('\n'.join(profitems)+'\n')
        profargs=['--cases',str(profpath),'80'] if isprinter else ['--manifest',str(profpath)]
        if valgrind:
            env=os.environ.copy()
            run([valgrind,'--tool=callgrind','--callgrind-out-file='+str(dest/'callgrind.out'),str(printer/'profiled' if isprinter else lint/'profiled'),*profargs],dest/'profile-answer',env=env)
            spec=importlib.util.spec_from_file_location('accounting',REPO/'stage1/typescript/scanner/profile.py')
            mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
            cg=mod.summarize(dest/'callgrind.out')
            run([valgrind,'--tool=callgrind','--callgrind-out-file='+str(dest/'go.callgrind.out'),commands['Go'][0],*profargs],dest/'go-callgrind-answer',env={**env,'GOMAXPROCS':'1','GODEBUG':'asyncpreemptoff=1'})
            go_cg=mod.summarize(dest/'go.callgrind.out')
            checked((dest/'profile-answer').read_bytes(),(dest/'go-callgrind-answer').read_bytes())
            categories={}
            for name,cost in cg['self_all'].items():
                category='release' if name in ('adamic_release','release_last','free_one','let_go','list','deallocate','give','adamic_object_free_children','adamic_array_free_children') else 'retain' if name=='adamic_retain' else 'allocation' if name in ('adamic_allocate','malloc','free','realloc','calloc') else 'strings' if name.startswith('adamic_string_') or name in ('decode','unit_at','usable') else 'other'
                categories[category]=categories.get(category,0)+cost
            values['profile']=dict(go_instructions=go_cg['total'],go_self=go_cg['self'][:20],categories=categories,file=first_file,rows=len(profitems),instructions=cg['total'],instructions_per_row=cg['total']/len(profitems),self=cg['self'][:30],inclusive=cg['inclusive'][:30])
        env=os.environ.copy();env['SCOUT_PPROF']=str(dest/'go.pprof')
        run([str(directory/'printer-oracle' if isprinter else lint/'oracle-profile'),*args],dest/'go-profile-answer',env=env)
        checked((dest/'go-profile-answer').read_bytes(),answers['Go'])
        run(['go','tool','pprof','-top',dest/'go.pprof'],dest/'go-top.txt')
        summary['workloads'][workload]=values
        (directory/'measurements.json').write_text(json.dumps(summary,indent=2)+'\n')
    # Empty manifest isolates process/import/runtime start and shutdown, without a parse.
    start_dir=directory/'measure/startup';start_dir.mkdir(exist_ok=True)
    empty=start_dir/'empty.txt';empty.write_text('')
    summary['startup']={}
    for name,command in dict(native=[lint/'scanner'],Go=[lint/'oracle'],Node=['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',lint/'main.ts']).items():
        samples=[]
        for r in range(rounds):
            output=start_dir/f'{name}-{r}'
            samples.append(run([*command,'--manifest',empty,'--count'],output))
            checked(output.read_bytes(),b'0\n')
        summary['startup'][name]=samples
    # Typed pilot uses live native Go checker facts. Node replays the recorded facts; its timing is
    # deliberately not presented as a native-checker comparison.
    pilot=directory/'measure/checker';pilot.mkdir(exist_ok=True)
    witness=lint/'rules/no-unnecessary-boolean-literal-compare/testdata/witness.ts.txt'
    source=pilot/'witness.ts';source.write_bytes(witness.read_bytes())
    config=pilot/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True},'files':[str(source)]}))
    manifest=pilot/'pilot.txt';manifest.write_text('program '+str(config)+'\n'+str(source)+'\t@typescript-eslint/no-unnecessary-boolean-literal-compare\n')
    go=run([lint/'oracle','--manifest',manifest],pilot/'Go')
    native=run([lint/'scanner','--manifest',manifest,'--record',pilot/'transcript'],pilot/'native')
    node=run(['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',lint/'main.ts','--manifest',manifest,'--replay',pilot/'transcript'],pilot/'Node')
    checked((pilot/'native').read_bytes(),(pilot/'Go').read_bytes());checked((pilot/'Node').read_bytes(),(pilot/'Go').read_bytes())
    samples={'Go':[],'native':[],'Node_replay':[]}
    for r in range(rounds):
        for name,command in dict(Go=[lint/'oracle','--manifest',manifest],native=[lint/'scanner','--manifest',manifest],Node_replay=['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',lint/'main.ts','--manifest',manifest,'--replay',pilot/'transcript']).items():
            output=pilot/f'{name}-{r}';samples[name].append(run(command,output))
            checked(output.read_bytes(),(pilot/'Go').read_bytes())
    env={**os.environ,'ADAMIC_TSGO_PROFILE':str(pilot/'native.pprof'),'ADAMIC_TSGO_TIMING':'1'}
    run([lint/'scanner','--manifest',manifest],pilot/'native-profile',env=env)
    checked((pilot/'native-profile').read_bytes(),(pilot/'Go').read_bytes())
    run(['go','tool','pprof','-top',pilot/'native.pprof'],pilot/'native-top.txt')
    summary['checker']=dict(recording=native,Node_replay=node,samples=samples,profile_counters=Path(str(pilot/'native-profile')+'.stderr').read_text(),bytes=(pilot/'Go').stat().st_size)
    summary['other_checker_rules']={}
    for descriptor in json.loads((directory/'registered.json').read_text()):
        if not descriptor.get('typed') or descriptor['name']=='@typescript-eslint/no-unnecessary-boolean-literal-compare':continue
        slug=next(p.parent.name for p in (lint/'rules').glob('*/rule.json') if json.loads(p.read_text())['name']==descriptor['name'])
        dest=directory/'measure'/slug;dest.mkdir(exist_ok=True)
        source=dest/'witness.ts';source.write_bytes((lint/'rules'/slug/'testdata/witness.ts.txt').read_bytes())
        config=dest/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True},'files':[str(source)]}))
        manifest=dest/'pilot.txt';manifest.write_text('program '+str(config)+'\n'+str(source)+'\t'+descriptor['name']+'\n')
        run([lint/'oracle','--manifest',manifest],dest/'Go')
        run([lint/'scanner','--manifest',manifest,'--record',dest/'transcript'],dest/'native')
        run(['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',lint/'main.ts','--manifest',manifest,'--replay',dest/'transcript'],dest/'Node')
        for name in ('native','Node'):checked((dest/name).read_bytes(),(dest/'Go').read_bytes())
        samples={'Go':[],'native':[],'Node_replay':[]}
        for r in range(rounds):
            for name,command in dict(Go=[lint/'oracle','--manifest',manifest],native=[lint/'scanner','--manifest',manifest],Node_replay=['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',lint/'main.ts','--manifest',manifest,'--replay',dest/'transcript']).items():
                output=dest/f'{name}-{r}';samples[name].append(run(command,output))
                checked(output.read_bytes(),(dest/'Go').read_bytes())
        env={**os.environ,'ADAMIC_TSGO_PROFILE':str(dest/'native.pprof'),'ADAMIC_TSGO_TIMING':'1'}
        run([lint/'scanner','--manifest',manifest],dest/'native-profile',env=env)
        checked((dest/'native-profile').read_bytes(),(dest/'Go').read_bytes())
        run(['go','tool','pprof','-top',dest/'native.pprof'],dest/'native-top.txt')
        summary['other_checker_rules'][descriptor['name']]=dict(samples=samples,profile_counters=Path(str(dest/'native-profile')+'.stderr').read_text(),bytes=(dest/'Go').stat().st_size)
    # Before/after runs use the first 16 owned rows; every protocol byte must survive.
    items=rows(directory/'owned.txt')[:16]; dest=directory/'measure/batching';dest.mkdir(exist_ok=True)
    combined=dest/'batch.txt';combined.write_text('\n'.join(items)+'\n')
    samples={'before':[],'after':[]}
    for r in range(rounds):
        before=bytearray();wall=0;peak=0;per_file=[]
        for i,row in enumerate(items):
            path=dest/f'row-{i}.txt';path.write_text(row+'\n')
            output=dest/f'before-{r}-{i}'
            result=run([lint/'scanner','--manifest',path],output)
            data=output.read_bytes()
            # One-file runners number their sole case 0. Preserve the batch's case number.
            if not data.startswith(b'case 0\n'):raise ValueError('missing case boundary')
            before.extend(f'case {i}\n'.encode()+data[len(b'case 0\n'):])
            wall+=result['seconds'];peak=max(peak,result['max_rss_kib']);per_file.append(dict(row=row,**result))
        result=run([lint/'scanner','--manifest',combined],dest/f'after-{r}')
        checked((dest/f'after-{r}').read_bytes(),bytes(before))
        samples['before'].append(dict(seconds=wall,max_rss_kib=peak,per_file=per_file));samples['after'].append(result)
    result=run([lint/'oracle','--manifest',combined],dest/'Go')
    checked((dest/'Go').read_bytes(),bytes(before))
    run(['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',lint/'main.ts','--manifest',combined],dest/'Node')
    checked((dest/'Node').read_bytes(),bytes(before))
    summary['batching']=dict(rows=len(items),before_processes=len(items),after_processes=1,samples=samples,
        bytes=len(before),sha256=hashlib.sha256(before).hexdigest())
    summary['machine_after']=machine()
    (directory/'measurements.json').write_text(json.dumps(summary,indent=2)+'\n')
    print(json.dumps(summary,indent=2))

def encoding_measure(directory, valgrind, rounds):
    global USAGE_HELPER
    USAGE_HELPER=directory/'usage'
    baseline=directory/'lint'; optimized=directory/'encoding';printer=directory/'printer'
    result={'machine_before':machine(),'workloads':{}}
    for workload in ('owned','upstream','compiler','public','repository'):
        manifest=directory/(workload+'.txt');args=['--manifest',manifest]
        dest=directory/'encoding-measure'/workload;dest.mkdir(parents=True,exist_ok=True)
        commands=dict(Go=[baseline/'oracle',*args],before=[baseline/'scanner',*args],after=[optimized/'scanner',*args],
            Node_before=['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',baseline/'main.ts',*args],
            Node_after=['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',optimized/'main.ts',*args])
        want=(directory/'measure'/workload/'Go.answer').read_bytes()
        samples={name:[] for name in commands}
        for r in range(rounds):
            names=list(commands);names=names[r%len(names):]+names[:r%len(names)]
            for name in names:
                output=dest/f'{name}-{r}';samples[name].append(run(commands[name],output))
                checked(output.read_bytes(),want)
                if Path(str(output)+'.stderr').stat().st_size:raise ValueError('encoding sample stderr '+name)
        run([optimized/'counted',*args],dest/'counted');checked((dest/'counted').read_bytes(),want)
        result['workloads'][workload]=dict(rows=len(rows(manifest)),bytes=len(want),sha256=hashlib.sha256(want).hexdigest(),
            samples=samples,median={k:statistics.median(x['seconds'] for x in v) for k,v in samples.items()},
            allocation_counts=Path(str(dest/'counted')+'.stderr').read_text())
    for workload in result['workloads'].values(): balanced_counts(workload['allocation_counts'])
    # All three registered typed witnesses use precisely the baseline recorded facts on Node.
    for name in ('checker','no-redundant-type-constituents','prefer-find'):
        pilot=directory/'measure'/name
        want=(pilot/'Go').read_bytes()
        dest=directory/'encoding-measure'/name;dest.mkdir(exist_ok=True)
        run([optimized/'scanner','--manifest',pilot/'pilot.txt'],dest/'native')
        run(['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',optimized/'main.ts','--manifest',pilot/'pilot.txt','--replay',pilot/'transcript'],dest/'Node')
        for side in ('native','Node'):checked((dest/side).read_bytes(),want)
    if valgrind:
        dest=directory/'encoding-measure/compiler';manifest=directory/'measure/compiler/profile-input.txt'
        run([valgrind,'--tool=callgrind','--callgrind-out-file='+str(dest/'callgrind.out'),optimized/'profiled','--manifest',manifest],dest/'profile-answer')
        checked((dest/'profile-answer').read_bytes(),(directory/'measure/compiler/profile-answer').read_bytes())
        spec=importlib.util.spec_from_file_location('accounting',REPO/'stage1/typescript/scanner/profile.py')
        mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
        profile=mod.summarize(dest/'callgrind.out')
        result['profile']=dict(instructions=profile['total'],self=profile['self'][:30],inclusive=profile['inclusive'][:30])
    result['machine_after']=machine()
    (directory/'encoding-measurements.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))

def no_skips(answer):
    if re.search(rb'(?m)^(?:skipped|refused) ',answer):
        raise ValueError('workload has skipped/refused rows')

def fused(directory, rounds):
    global USAGE_HELPER
    USAGE_HELPER=directory/'usage'
    dest=directory/'fused';dest.mkdir(exist_ok=True)
    files=[x.split('\t')[0] for x in rows(directory/'compiler.txt')]
    config=dest/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True},'files':files}))
    manifest=dest/'manifest.txt';manifest.write_text('program '+str(config)+'\n'+''.join(path+'\tall\n' for path in files))
    before=directory/'lint';after=directory/'encoding'
    run([before/'oracle','--manifest',manifest],dest/'Go')
    want=(dest/'Go').read_bytes()
    run([before/'scanner','--manifest',manifest,'--record',dest/'transcript'],dest/'record')
    checked((dest/'record').read_bytes(),want)
    no_skips(want)
    commands=dict(Go=[before/'oracle','--manifest',manifest],before=[before/'scanner','--manifest',manifest],after=[after/'scanner','--manifest',manifest],
        Node_before=['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',before/'main.ts','--manifest',manifest,'--replay',dest/'transcript'],
        Node_after=['node','--disable-warning=ExperimentalWarning',REPO/'oracle/node.mjs',after/'main.ts','--manifest',manifest,'--replay',dest/'transcript'])
    report={'machine_before':machine(),'files':len(files),'rules':len(json.loads((directory/'registered.json').read_text())),'bytes':len(want),'findings':want.count(b'\nrange '),'samples':{name:[] for name in commands}}
    for r in range(rounds):
        names=list(commands);names=names[r%len(names):]+names[:r%len(names)]
        for name in names:
            output=dest/f'{name}-{r}';report['samples'][name].append(run(commands[name],output));checked(output.read_bytes(),want)
    for side,root in (('before',before),('after',after)):
        output=dest/(side+'-counted');run([root/'counted','--manifest',manifest],output);checked(output.read_bytes(),want)
        report[side+'_counts']=Path(str(output)+'.stderr').read_text()
        balanced_counts(report[side+'_counts'])
    env={**os.environ,'ADAMIC_TSGO_PROFILE':str(dest/'native.pprof'),'ADAMIC_TSGO_TIMING':'1'}
    run([after/'scanner','--manifest',manifest],dest/'profile-answer',env=env);checked((dest/'profile-answer').read_bytes(),want)
    run(['go','tool','pprof','-top',dest/'native.pprof'],dest/'native-top.txt')
    report['profile_counters']=Path(str(dest/'profile-answer')+'.stderr').read_text()
    report['machine_after']=machine();(directory/'fused.json').write_text(json.dumps(report,indent=2)+'\n')
    print(json.dumps(report,indent=2))

def fused_profile(directory, valgrind):
    global USAGE_HELPER
    if not valgrind: raise ValueError('--valgrind required')
    USAGE_HELPER=directory/'usage'
    dest=directory/'fused'
    manifest=dest/'profile-input.txt'
    row=next(x for x in rows(dest/'manifest.txt') if x.split('\t')[0].endswith('/scanner.ts'))
    manifest.write_text(rows(dest/'manifest.txt')[0]+'\n'+row+'\n')
    spec=importlib.util.spec_from_file_location('accounting',REPO/'stage1/typescript/scanner/profile.py')
    mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
    report={"machine_before":machine(),"profile_environment":{"GOMAXPROCS":"1","GODEBUG":"asyncpreemptoff=1"}}
    for side,binary in (('Go',directory/'lint/oracle'),('before',directory/'lint/profiled'),('after',directory/'encoding/profiled')):
        env={**os.environ,'GOMAXPROCS':'1','GODEBUG':'asyncpreemptoff=1'}
        output=dest/(side+'-callgrind-answer')
        run([valgrind,'--tool=callgrind','--callgrind-out-file='+str(dest/(side+'.callgrind.out')),binary,'--manifest',manifest],output,env=env)
        if side=='Go': want=output.read_bytes();no_skips(want)
        else: checked(output.read_bytes(),want)
        cg=mod.summarize(dest/(side+'.callgrind.out'))
        report[side]=cg
    report['machine_after']=machine()
    (directory/'fused-profile.json').write_text(json.dumps(report,indent=2)+'\n')


def fetch(directory):
    pins=json.loads((HERE/'public-pins.json').read_text())
    directory.mkdir(parents=True,exist_ok=True)
    def one(record):
        repo=directory/record['name'];repo.mkdir(exist_ok=True)
        try:
            run(['git','init','-q',repo],repo/'init.log')
            run(['git','-C',repo,'fetch','--depth=1','--filter=blob:none','https://github.com/'+record['repo']+'.git',record['sha']],repo/'fetch.log')
            run(['git','-C',repo,'-c','core.hooksPath=/dev/null','-c','filter.lfs.smudge=','-c','filter.lfs.required=false','checkout','--detach',record['sha']],repo/'checkout.log')
            record={**record,'status':'ready','tracked_ts':len(subprocess.check_output(['git','-C',repo,'ls-files','*.ts']).splitlines())}
        except Exception as e:record={**record,'status':'failed','error':str(e)}
        return record
    with ThreadPoolExecutor(max_workers=4) as pool:results=list(pool.map(one,pins))
    (directory/'fetch-results.json').write_text(json.dumps(results,indent=2)+'\n')
    print(json.dumps(results,indent=2))

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode',choices=['prepare','prepare-encoding','measure','encoding','fused','fused-profile','fetch'])
    parser.add_argument('directory',type=Path)
    parser.add_argument('--corpus',type=Path)
    parser.add_argument('--valgrind')
    parser.add_argument('--rounds',type=int,default=5)
    args=parser.parse_args();directory=args.directory.resolve()
    if args.mode=='prepare':
        if args.corpus is None:parser.error('--corpus required')
        directory.mkdir(parents=True,exist_ok=True)
        with (directory/'.prepare.lock').open('w') as lock:
            fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
            prepare(directory,args.corpus.resolve())
    elif args.mode=='prepare-encoding':
        overlay=directory/'lint-overlay.json'
        env={**os.environ,'ADAMIC_SCOUT_OUTPUT':str(directory/'encoding'),'ADAMIC_SCOUT_ENCODER':str(HERE/'encode.ts')}
        run(['go','test','-count=1','-v','-timeout=30m','-overlay='+str(overlay),'-run=^TestScoutEncoding$','./stage1/cohere/lint'],directory/'encoding-build.log',env=env)
    elif args.mode=='measure':
        if args.rounds<1:parser.error('--rounds must be positive')
        measure(directory,args.valgrind,args.rounds)
    elif args.mode=='encoding':encoding_measure(directory,args.valgrind,args.rounds)
    elif args.mode=='fused':fused(directory,args.rounds)
    elif args.mode=='fused-profile':fused_profile(directory,args.valgrind)
    else:fetch(directory)
if __name__=='__main__':main()
