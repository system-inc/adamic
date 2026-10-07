#!/usr/bin/env python3
"""Rule-owned four-backend comparison, corpus capture, repairs, timing and mutants."""
import argparse
import difflib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import time

repository = Path(__file__).resolve().parents[5]
owned = Path(__file__).resolve().parent
cohere = (repository / 'cohere').resolve()
parser = argparse.ArgumentParser()
parser.add_argument('--work', default='/tmp/wave109-validation')
parser.add_argument('--compiler', default='/tmp/wave109-typescript')
parser.add_argument('--slug', action='append')
parser.add_argument('--skip-corpus', action='store_true')
args = parser.parse_args()
work = Path(args.work); work.mkdir(parents=True, exist_ok=True)
slugs = args.slug or ['typescript-no-non-null-asserted-optional-chain', 'typescript-no-non-null-assertion', 'typescript-no-this-alias']

def run(command, label, cwd=repository, env=None, check=True):
    output = work / (label + '.stdout.log'); error = work / (label + '.stderr.log')
    started = time.monotonic()
    with output.open('wb') as out, error.open('wb') as err:
        result = subprocess.run(command, cwd=cwd, stdout=out, stderr=err, env=env)
    elapsed = time.monotonic() - started
    if check and (result.returncode != 0 or error.stat().st_size):
        raise RuntimeError(f'{label}: exit {result.returncode}\n{error.read_text()}\n{output.read_text()[-2000:]}')
    return output.read_bytes(), elapsed, result.returncode

def equal(actual, expected, label):
    if actual != expected:
        (work / (label + '.diff.log')).write_text(''.join(difflib.unified_diff(expected.decode().splitlines(True), actual.decode().splitlines(True), fromfile='Go', tofile=label)))
        raise AssertionError(label + ' differs; see ' + str(work / (label + '.diff.log')))

# Capture tests without changing rule bodies or repository files.
capture = work / 'capture'; capture.mkdir(exist_ok=True)
harness = cohere / 'internal/lint/testing/rule_testing.go'
source = harness.read_text()
anchor = 'return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
assert source.count(anchor) == 1
side = work / 'rule_testing.go'
side.write_text(source.replace(anchor, 'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\nRecordAssertedCase(t, result)\nreturn result'))
replacements = {str(harness): str(side)}
if 'react-no-unescaped-entities' in slugs:
    capture_source = cohere / 'internal/lint/testing/docs_capture.go'
    capture_side = work / 'docs_capture.go'
    wire = capture_source.read_text().replace('"encoding/json"', '"encoding/json"\n"reflect"').replace('json.Marshal(result.capture.options)', 'json.Marshal(wave109WireOptions(result.capture.options))')
    wire += r'''
// Reflection reads data only; no rule listener, decoder or diagnostic is changed.
func wave109WireOptions(value any) any {
 v:=reflect.ValueOf(value);if v.Kind()==reflect.Pointer {v=v.Elem()}
 if !v.IsValid() || v.Kind()!=reflect.Struct || v.Type().Name()!="NoUnescapedEntitiesOptions" {return value}
 f:=v.FieldByName("Forbid");if f.IsNil(){return map[string]any{}}
 entries:=make([]any,0);f=f.Elem()
 for i:=0;i<f.Len();i++ {e:=f.Index(i);c:=e.FieldByName("character").String();if !e.FieldByName("hasAlternatives").Bool(){entries=append(entries,c);continue};a:=e.FieldByName("alternatives");alts:=make([]string,0);for j:=0;j<a.Len();j++ {alts=append(alts,a.Index(j).String())};entries=append(entries,map[string]any{"char":c,"alternatives":alts})}
 return map[string]any{"forbid":entries}
}
'''
    capture_side.write_text(wire)
    replacements[str(capture_source)] = str(capture_side)
overlay = work / 'capture-overlay.json'; overlay.write_text(json.dumps({'Replace': replacements}))
filters = [json.loads((owned.parent / slug / 'rule.json').read_text())['upstreamTest'] for slug in slugs]
environment = dict(os.environ, COHERE_DOCS_CAPTURE=str(capture))
packages=sorted(set(json.loads((owned.parent / slug / 'rule.json').read_text())['upstreamPackage'] for slug in slugs))
run(['go','test','-overlay='+str(overlay)]+['./internal/lint/rules/'+name for name in packages]+['-run','^('+'|'.join(filters)+')','-count=1','-v','-timeout=15m'], 'upstream', cwd=cohere, env=environment)
records = {}
for file in capture.glob('*.jsonl'):
    for line in file.read_text().splitlines():
        row = json.loads(line)
        records[(row['rule'], row['source'], json.dumps(row.get('options'), sort_keys=True), row['file'])] = row

summary = {}
for slug in slugs:
    directory = owned.parent / slug
    desc = json.loads((directory / 'rule.json').read_text())
    subject = desc['name']
    case_dir = work / slug; case_dir.mkdir(exist_ok=True)
    rows = []
    for number, key in enumerate(sorted(records)):
        row = records[key]
        if row['rule'] != subject: continue
        if subject in ['structure/tailwind-no-physical-direction','@eslint-community/eslint-comments/require-description'] and ('<div className=' in row['source'] or '<div>' in row['source']):
            (case_dir/'blocked-jsx.ts.txt').write_text(row['source']);continue
        fixture = case_dir / f'upstream-{number}.ts.txt'; fixture.write_text(row['source'])
        rows.append(str(fixture)+'\t'+('/captured/'+row['file'] if not row['file'].startswith('/') else row['file'])+'\t'+json.dumps(row.get('options'), separators=(',',':')))
    for fixture in sorted((directory/'testdata').glob('*.ts.txt')):
        rows.append(str(fixture)+('\t/source.tsx\t' if subject.startswith('react/') else '\t/source.ts\t'))
        if subject.endswith('/no-this-alias'):
            for options in [{'AllowedNames':['self']}, {'ReportDestructuring':True}, {'ReportDestructuring':True,'AllowedNames':['x','self']}]:
                rows.append(str(fixture)+'\t/source.ts\t'+json.dumps(options,separators=(',',':')))
            rows.append(str(fixture)+'\t/source.js\t')
    manifest = case_dir/'cases.txt'; manifest.write_text('\n'.join(rows)+'\n')
    virtual = cohere / ('wave109_'+slug.replace('-','_')+'.go')
    overlay = case_dir/'overlay.json'; overlay.write_text(json.dumps({'Replace': {str(virtual):str(directory/'testdata/comparison.go.txt')}}))
    oracle = case_dir/'oracle'
    run(['go','build','-overlay='+str(overlay),'-o',str(oracle),str(virtual)],slug+'-oracle-build',cwd=cohere)
    binary = case_dir/'native'
    _,_,build_code=run(['go','run',str(owned/'build_probe.go'),str(directory/'probe.ts'),str(binary)],slug+'-build',check=False)
    commands = {'Go':[str(oracle)], 'Node':['node','--disable-warning=ExperimentalWarning',str(repository/'oracle/node.mjs'),str(directory/'probe.ts')], 'JavaScript':['node','--disable-warning=ExperimentalWarning',str(repository/'oracle/node.mjs'),str(binary)+'.mjs'], 'native':[str(binary)]}
    expected,_,_=run(commands['Go']+[str(manifest)],slug+'-cases-Go')
    backends=['Node'] if build_code else ['Node','JavaScript','native']
    for backend in backends:
        observed,_,_=run(commands[backend]+[str(manifest)],slug+'-cases-'+backend)
        equal(observed,expected,slug+'-cases-'+backend)
    print(slug, 'findings and full suggestions identical',len(rows),'cases',len(expected),'bytes',flush=True)
    summary[slug]={'cases':len(rows),'bytes':len(expected)}
    # Same witness, compiled mutant: every backend must finish cleanly and disagree with Go.
    mutant = case_dir/'mutant'; mutant.mkdir(exist_ok=True)
    change = json.loads((directory/'mutant.json').read_text())
    for file in directory.rglob('*.ts'):
        text = file.read_text()
        if file.name == change.get('file','rule.ts'):
            assert text.count(change['from']) == 1
            text = text.replace(change['from'],change['to'],1)
        def rewrite(match):
            target = match.group(2)
            if target.startswith('.') and not (file.parent/target).resolve().is_relative_to(directory):
                target = str((file.parent/target).resolve())
            return match.group(1)+target+match.group(3)
        text = re.sub(r"(from\s+['\"])([^'\"]+)(['\"])", rewrite, text)
        destination=mutant/file.relative_to(directory);destination.parent.mkdir(parents=True,exist_ok=True);destination.write_text(text)
    mutant_binary = case_dir/'mutant-native'
    if not build_code:
        run(['go','run',str(owned/'build_probe.go'),str(mutant/'probe.ts'),str(mutant_binary)],slug+'-mutant-build')
    mutation_commands = {'Node':['node','--disable-warning=ExperimentalWarning',str(repository/'oracle/node.mjs'),str(mutant/'probe.ts')], 'JavaScript':['node','--disable-warning=ExperimentalWarning',str(repository/'oracle/node.mjs'),str(mutant_binary)+'.mjs'], 'native':[str(mutant_binary)]}
    for backend,command in mutation_commands.items():
        if backend not in backends: continue
        observed,_,_=run(command+[str(manifest)],slug+'-mutant-'+backend)
        assert observed != expected, (slug,backend,'survived')
        print(slug,'mutant',change['name'],'caught only by comparison on',backend,flush=True)
    if build_code:
        print(slug,'native/JavaScript lowering blocked; retained source, Node comparison and mutant',flush=True)
        summary[slug]['build_blocked']='see build stderr'
    if args.skip_corpus: continue
    compiler = Path(args.compiler)
    pin = subprocess.check_output(['git','rev-parse','HEAD'],cwd=compiler,text=True).strip()
    assert pin == '050880ce59e30b356b686bd3144efe24f875ebc8'
    sources = sorted((compiler/'src/compiler').rglob('*.ts'))
    stage = sorted(file for file in (repository/'stage1').rglob('*') if file.suffix in ['.ts','.a'])
    all_sources = sources + stage
    snapshots=case_dir/'sources';snapshots.mkdir(exist_ok=True)
    frozen=[]
    for number,file in enumerate(all_sources):
        snapshot=snapshots/(str(number)+'.ts.txt');snapshot.write_bytes(file.read_bytes());frozen.append(str(snapshot)+'\t'+str(file)+'\t')
    corpus_manifest=case_dir/'corpus.txt';corpus_manifest.write_text('\n'.join(frozen)+'\n')
    oracle_output,_,code=run(commands['Go']+[str(corpus_manifest)],slug+'-corpus-Go',check=False)
    if code:
        print(slug,'Go corpus refused; see stderr log',flush=True)
        summary[slug]['corpus_blocked']='Go parse diagnostic'
        continue
    for backend in backends:
        observed,_,code=run(commands[backend]+[str(corpus_manifest)],slug+'-corpus-'+backend,check=False)
        if code:
            print(slug,backend,'corpus refused; see stderr log',flush=True)
            summary[slug]['corpus_blocked']=backend+' parser refusal'
            break
        equal(observed,oracle_output,slug+'-corpus-'+backend)
    else:
        print(slug,'compiler and stage1 identical',len(all_sources),'files',flush=True)
        summary[slug]['corpus_files']=len(all_sources)
    # Throughput is count-only over the pinned compiler, including process startup.
    compiler_manifest=case_dir/'compiler.txt';compiler_manifest.write_text('\n'.join(frozen[:len(sources)])+'\n')
    best={};counts=None
    for round_number in range(3):
        for backend in (['Go','Node'] if build_code else ['Go','Node','native']):
            output,elapsed,code=run(commands[backend]+[str(compiler_manifest),'--count'],slug+f'-rate-{round_number}-'+backend,check=False)
            if code:
                print(slug,backend,'throughput blocked',flush=True)
                continue
            if counts is None:counts=output
            equal(output,counts,slug+'-rate-'+backend)
            best[backend]=min(best.get(backend,float('inf')),elapsed)
    summary[slug]['rates']={backend:{'seconds':elapsed,'findings':int(counts or b'0'),'findings_per_second':int(counts or b'0')/elapsed} for backend,elapsed in best.items()}
    print(slug,'rates',summary[slug]['rates'],flush=True)
(work/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
