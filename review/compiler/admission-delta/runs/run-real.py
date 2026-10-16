#!/usr/bin/env python3
"""Evidence runner: unchanged admission tool, real revisions, fresh Go caches."""
import collections
import json
import os
import pathlib
import shlex
import signal
import subprocess
import tempfile
import time

ROOT = pathlib.Path('/workspace/adamic')
OUT = ROOT / 'review/compiler/admission-delta/runs'
BASE = 'cf735d9fba9e38de6368575e5630e44375a86eaf'
HEAD = 'ced32bf9b1f410b8be07f078adb78365a20a9f5b'
TOOL = '/tmp/admission-real-tool'
SCRATCH = pathlib.Path(tempfile.mkdtemp(prefix='admission-real-', dir='/workspace'))
STATE = {'base': BASE, 'fixes_head': HEAD, 'scratch': str(SCRATCH), 'phases': [], 'runs': {}}
(SCRATCH / 'tmp').mkdir(mode=0o777)
ENV = dict(os.environ, GOMAXPROCS='4', GOPROXY='https://proxy.golang.org|direct', TMPDIR=str(SCRATCH/'tmp'))

def save():
    (OUT / 'timings.json').write_text(json.dumps(STATE, indent=2) + '\n')

def phase(name, command, cwd=ROOT, env=ENV, limit=600, result=None):
    record = {'phase': name, 'command': list(map(str, command)), 'cwd': str(cwd), 'status': 'running', 'wall_start': time.time(), 'limit_seconds': limit}
    STATE['phases'].append(record)
    save()
    began = time.monotonic()
    with (OUT / (name + '.log')).open('w') as log:
        output = open(result, 'w') if result else log
        try:
            process = subprocess.Popen(command, cwd=cwd, env=env, stdout=output, stderr=log, start_new_session=True)
            try:
                status = process.wait(timeout=limit)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                process.wait()
                status = 124
        finally:
            if result:
                output.close()
    record.update(status=status, wall_seconds=time.monotonic()-began, wall_end=time.time())
    save()
    print(name + ': exit ' + str(status) + ', wall ' + str(round(record['wall_seconds'],3)) + ' s', flush=True)
    return status

def compiler(label, sha):
    tree = SCRATCH / (label+'-checkout')
    assert phase(label+'-checkout', ['git','worktree','add','--detach',str(tree),sha]) == 0
    cohere_sha = subprocess.check_output(['git','rev-parse',sha+':cohere'],cwd=ROOT,text=True).strip()
    typescript_sha = subprocess.check_output(['git','rev-parse',cohere_sha+':TypeScript'],cwd=ROOT/'cohere',text=True).strip()
    assert phase(label+'-cohere', ['git','worktree','add','--detach',str(tree/'cohere'),cohere_sha],cwd=ROOT/'cohere') == 0
    assert phase(label+'-typescript', ['git','worktree','add','--detach',str(tree/'cohere/TypeScript'),typescript_sha],cwd=ROOT/'cohere/TypeScript') == 0
    cache = SCRATCH / (label+'-cold-cache')
    cache.mkdir()
    STATE[label+'_cache'] = {'path':str(cache),'entries_before_build':len(list(cache.iterdir()))}
    save()
    binary = SCRATCH / (label+'-adamic')
    assert phase(label+'-cold-build', ['go','build','-p','4','-o',str(binary),'./cmd/adamic'], cwd=tree, env=dict(ENV,GOCACHE=str(cache)),limit=900) == 0
    return binary

def observer(binary, role, trace):
    script = SCRATCH / ('observe-'+role+'-'+trace.stem)
    script.write_text('#!/bin/bash\nstarted=$EPOCHREALTIME\n'+shlex.quote(str(binary))+' "$@"\nstatus=$?\nprintf "%s\\t%s\\t%s\\t%s\\t%s\\t%s\\n" '+shlex.quote(role)+' "$1" "$2" "$started" "$EPOCHREALTIME" "$status" >> '+shlex.quote(str(trace))+'\nexit "$status"\n')
    script.chmod(0o755)
    return str(script)

def run(label, sha, base_binary, head_binary, budget=None):
    generated = OUT / (label+'.generated-manifest.json')
    assert phase(label+'-manifest', ['python3','cloud/admission-corpus/manifest.py','--sha',sha],result=generated,limit=60) == 0
    m = json.loads(generated.read_text())
    # These requested revisions do not contain the generator. Keep all corpus
    # entries untouched, and keep external generator provenance in timings.json.
    m['generator'] = ''
    execution = OUT / (label+'.execution-manifest.json')
    execution.write_text(json.dumps(m,indent=2)+'\n')
    trace = OUT / (label+'.commands.tsv')
    trace.write_text('')
    b = observer(base_binary,'base',trace)
    h = observer(head_binary,'head',trace)
    result = OUT / (label+'.json')
    command = [TOOL,'--base',BASE,'--head',sha,'--base-binary',b,'--head-binary',h,'--manifest',str(execution),'--json']
    if budget is not None:
        command += ['--budget',str(budget)]
    exit_code = phase(label+'-admission',command,limit=1800,result=result)
    r = json.loads(result.read_text())
    rows = [line.rstrip('\n').split('\t') for line in trace.read_text().splitlines()]
    classifications = [row for row in rows if row[1]=='c']
    backend_builds = [row for row in rows if row[1] in ['js','build']]
    def durations(rows):
        return {'commands':len(rows),'command_wall_sum_seconds':sum(float(x[4])-float(x[3]) for x in rows),'wall_span_seconds':max(float(x[4]) for x in rows)-min(float(x[3]) for x in rows)} if rows else {'commands':0,'command_wall_sum_seconds':0,'wall_span_seconds':0}
    STATE['runs'][label] = {'exit':exit_code,'base':r['base'],'head':r['head'],'verdict':r['verdict'],'admitted':r['admitted'],'sampled':r['sampling_size'],'omitted':r['omitted'],'classification':durations(classifications),'backend_builds':durations(backend_builds),'runtime_phase_wall_seconds':r['budget_used_seconds'],'classes':dict(collections.Counter(p['class'] for p in r['programs'])),'corpora':{c['name']:len(c['programs']) for c in r['corpora']},'disagreements':[p for p in r['programs'] if p['agree'] is False],'compiler_errors':[p for p in r['programs'] if p['class'].startswith('compiler-')]}
    save()
    print(label+': '+json.dumps({k:v for k,v in STATE['runs'][label].items() if k not in ['disagreements','compiler_errors']}),flush=True)
    return r

try:
    revision = subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip()
    blob = subprocess.check_output(['git','rev-parse',revision+':cloud/admission-corpus/manifest.py'],cwd=ROOT,text=True).strip()
    STATE['generator']={'revision':revision,'path':'cloud/admission-corpus/manifest.py','blob':blob,'exists_at_requested_heads':False}
    STATE['method']='Real compiler revisions built in isolated checkouts with separately empty GOCACHE directories, then supplied to the unchanged tool for phase instrumentation. Generated corpus entries are unmodified; execution manifests clear only absent-at-head generator metadata. Original generated manifests are preserved.'
    save()
    base_binary=compiler('main',BASE)
    run('main',BASE,base_binary,base_binary)
    head_binary=compiler('fixes',HEAD)
    run('fixes',HEAD,base_binary,head_binary,60)
finally:
    save()
