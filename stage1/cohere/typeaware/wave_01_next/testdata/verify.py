"""Isolated native/production-Go agreement, sanitizer, mutant and handle checks."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import time
from controls import controls

parser = argparse.ArgumentParser()
parser.add_argument('--repository', type=Path, required=True)
parser.add_argument('--typescript', type=Path, required=True)
parser.add_argument('--artifacts', type=Path, required=True)
args = parser.parse_args()
repo = args.repository.resolve(); corpus = args.typescript.resolve(); out = args.artifacts.resolve(); out.mkdir(parents=True, exist_ok=True)
source = repo / 'stage1/cohere/typeaware/wave_01_next'
records = []; sequence = 0

def run(name, command, *, cwd=repo, env=None, expected=0):
    global sequence
    sequence += 1; stem = out / ('%03d-%s' % (sequence, name)); started = time.perf_counter_ns()
    with Path(str(stem)+'.stdout').open('wb') as stdout, Path(str(stem)+'.stderr').open('wb') as stderr:
        process = subprocess.run(list(map(str, command)), cwd=cwd, env={**os.environ, **(env or {})}, stdout=stdout, stderr=stderr)
    data = Path(str(stem)+'.stdout').read_bytes(); error = Path(str(stem)+'.stderr').read_bytes()
    records.append({'name': name, 'command': list(map(str, command)), 'code': process.returncode, 'wall_ns': time.perf_counter_ns()-started, 'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest(), 'stderr': error.decode(errors='replace')})
    (out/'records.json').write_text(json.dumps(records, indent=2)+'\n')
    if process.returncode != expected: raise RuntimeError('%s exits %s, expected %s: %s' % (name, process.returncode, expected, error[-2000:]))
    return data, error

def build(name, archive, entry=source/'suite.a', sanitize=False):
    binary = out/name; command=[out/'adamic', 'build', entry, '-o', binary, '--tsgo', archive]
    if sanitize: command.append('--sanitize')
    run(name+'-build', command); return binary

def archive(name, sanitize=False, overlay=None):
    target=out/(name+'.a'); command=['go','build','-buildmode=c-archive','-o',target]
    if overlay: command += ['-overlay', overlay]
    command += ['./bridge/tsgo/archive']
    env={'CC':'clang','CGO_CFLAGS':'-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all'} if sanitize else None
    run(name+'-archive',command,env=env);return target

def compare(name, binary, config, manifest):
    want, _ = run(name+'-go',[oracle,config,manifest]); got, error = run(name+'-native',[binary,config,manifest])
    if error: raise RuntimeError(name+' native stderr')
    if got != want:
        at=next((at for at,(a,b) in enumerate(zip(got,want)) if a!=b), min(len(got),len(want)))
        raise RuntimeError('%s mismatch at byte %d: native %r Go %r' % (name,at,got[at:at+200],want[at:at+200]))
    print(name, 'bytes',len(got),got.splitlines()[-1].decode(),flush=True);return want

roots=controls(repo,out);config=out/'tsconfig.json';manifest=out/'controls.manifest'
# Cross-file called functions use TypeScript input fixtures, not native rule implementations.
(out/'Printer.ts').write_text("export async function renderRow(item:string){console.log(item);}\nexport async function storeRow(item:string){return item.length;}\n")
extra=["import {renderRow} from './Printer';export async function f(items:string[]){for(const item of items){await renderRow(item);}}", "import {storeRow} from './Printer';export async function f(items:string[]){for(const item of items){await storeRow(item);}}", "export async function f(items:string[]){for(const item of items){for(;(await Promise.resolve(false));){} }}", "export async function f(items:string[]){for(const item of items){for(await Promise.resolve();;)break;}}"]
for index,text in enumerate(extra):
    path=out/('extra-%02d.a'%index);path.write_text(text+'\nexport {};\n'); roots.append(path)
for path in sorted((repo/'cohere/internal/lint/rules/nexus/testdata/performance_no_independent_await_in_loop').rglob('*.ts')):
    target=out/('ahra-'+path.parent.name+'-'+path.stem+'.a');target.write_bytes(path.read_bytes());roots.append(target)
manifest.write_text('\n'.join(map(str,roots))+'\n')
run('stage0',['go','build','-o',out/'adamic','./cmd/adamic'])
normal=archive('checker');binary=build('native',normal)
virtual=repo/'cohere/adamic_wave01next.go';overlay=out/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/oracle.go')}}));oracle=out/'oracle'
run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],cwd=repo/'cohere')
truth=compare('controls',binary,config,manifest)
for rule in ['nexus/correctness-require-child-process-error-listener','nexus/correctness-require-response-status-check','nexus/performance-no-independent-await-in-loop']:
    if ('\t'+rule+'\t').encode() not in truth: raise RuntimeError('no positive control for '+rule)
for strict in [False]:
    cfg=json.loads(config.read_text());cfg['compilerOptions']['strict']=strict;relaxed=out/'relaxed.json';relaxed.write_text(json.dumps(cfg));compare('relaxed',binary,relaxed,manifest)
san=archive('checker-asan',True);asan=build('native-asan',san,sanitize=True);compare('controls-asan',asan,config,manifest)
for name,file,before,after in [('child','child_process_error_listener.a',"return syntaxTree.node(event).text === 'error';","return syntaxTree.node(event).text !== 'error';"),('response','response_status_check.a',"if(member === 'body')", "if(member === 'body-mutant')"),('independent','independent_await_in_loop.a','ruleRunner.byte(end),', 'ruleRunner.byte(end + 1),')]:
    directory=out/('mutant-'+name+'-source');directory.mkdir(exist_ok=True)
    for original in source.glob('*.a'):
        text=original.read_text()
        if original.name==file:
            if text.count(before)!=1: raise RuntimeError('nonunique mutant '+name)
            text=text.replace(before,after)
        text=re.sub(r"from '([^']+)'", lambda m: m.group(0) if m.group(1)=='adamic' or m.group(1).startswith('./') else "from '%s'" % (original.parent/m.group(1)).resolve(),text)
        (directory/original.name).write_text(text)
    mutant=build('mutant-'+name,normal,directory/'suite.a');got,error=run(name+'-mutant-run',[mutant,config,manifest])
    if error or got==truth: raise RuntimeError('mutant not caught cleanly '+name)
    first=next((at for at,(a,b) in enumerate(zip(got,truth)) if a!=b),min(len(got),len(truth)))
    print('mutant',name,'caught only by byte comparison at',first,flush=True)
for name,pin,base in [('compiler','050880ce59e30b356b686bd3144efe24f875ebc8',corpus),('repository',None,repo)]:
    if pin:
        actual,_=run('compiler-pin',['git','-C',corpus,'rev-parse','HEAD'])
        if actual.decode().strip()!=pin:raise RuntimeError('wrong compiler pin')
    lines=(repo/('stage1/cohere/typeaware/validation-coverage/'+name+'.manifest')).read_text().splitlines()
    paths=[base/line for line in lines if line];subject=out/(name+'.manifest');subject.write_text('\n'.join(map(str,paths))+'\n')
    cfg=corpus/'src/compiler/tsconfig.json' if name=='compiler' else repo/'tsconfig.json'
    compare(name,binary,cfg,subject);compare(name+'-asan',asan,cfg,subject)
    for round_number in range(3):
        want,error=run(name+'-timed-go-'+str(round_number),[oracle,cfg,subject],env={'ADAMIC_TSGO_TIMING':'1'})
        got,error=run(name+'-timed-native-'+str(round_number),[binary,cfg,subject],env={'ADAMIC_TSGO_TIMING':'1'})
        if want!=got:raise RuntimeError('timed agreement failed')
original=repo/'bridge/tsgo/archive/main.go';text=original.read_text();before='delete(programs.live, uint64(handle))'
if text.count(before)!=1:raise RuntimeError('nonunique handle mutant')
copy=out/'released-registry.go';copy.write_text(text.replace(before,'// Mutant retains the released handle.'));overlay=out/'released-overlay.json';overlay.write_text(json.dumps({'Replace':{str(original):str(copy)}}));retained=archive('checker-retained',overlay=overlay)
for index,question in enumerate(['declaration-context\nnode','syntax-flow-graph','reference-access','loop-header']):
    text='for(;;){}' if question=='loop-header' else 'function f(){return;}'
    kind='ForStatement' if question=='loop-header' else 'FunctionDeclaration';input_file=out/'probe-input.a';input_file.write_text(text)
    entry=out/'released.a';entry.write_text("import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,%d,'%s',%s));\n"%(len(text),kind,json.dumps(question)))
    probe=build('released-'+str(index),normal,entry);data,error=run('released-run-'+str(index),[probe,config,input_file],expected=70)
    if error!=b'adamic: panic: invalid or released checker handle\n':raise RuntimeError('wrong released failure')
    mutant=build('released-mutant-'+str(index),retained,entry);data,error=run('released-mutant-run-'+str(index),[mutant,config,input_file])
    if error:raise RuntimeError('retained handle mutant stderr')
    print('released',question.replace('\n',':'),'panic 70; retained mutant exits 0',flush=True)
run('checker-tests',['go','test','./bridge/tsgo/checker/...','-count=1'])
run('bridge-tests',['go','test','./bridge/tsgo','-count=1'])
print('PASS all three claimed rules',flush=True)
