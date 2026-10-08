#!/usr/bin/env python3
"""Independent full-repair parity for owned rules; no shared files are changed."""
import argparse, json, os, re, shutil, subprocess, time
from pathlib import Path
owned=Path(__file__).resolve().parent
repo=owned.parents[4]
a=argparse.ArgumentParser()
a.add_argument('--scratch',type=Path,required=True)
a.add_argument('--typescript',type=Path)
a.add_argument('--mutants',action='store_true')
a.add_argument('--benchmark',action='store_true')
args=a.parse_args()
s=args.scratch.resolve();s.mkdir(parents=True,exist_ok=True)
subjects=[('no-async-promise-executor','NoAsyncPromiseExecutor'),('no-case-declarations','NoCaseDeclarations'),('no-compare-neg-zero','NoCompareNegZero')]
cohere=repo/'cohere'

def run(command,label,cwd=repo,env=None,allow_failure=False):
    out=s/(label+'.stdout.txt');err=s/(label+'.stderr.txt')
    start=time.perf_counter()
    with out.open('wb') as stdout,err.open('wb') as stderr:
        result=subprocess.run([str(x) for x in command],cwd=cwd,env=env,stdout=stdout,stderr=stderr)
    duration=time.perf_counter()-start
    if (result.returncode or err.stat().st_size) and not allow_failure:
        raise SystemExit(f'{label}: exit={result.returncode}, stdout={out}, stderr={err}')
    return result.returncode,out.read_bytes(),duration

# Unchanged Go parser/listeners, formatter and converging fixer. Only serialization
# is expanded in a scratch source to expose all independently computed repairs.
source=(repo/'stage1/cohere/lint/testdata/oracle.go').read_text()
start=source.index('\t\trepair, replacement, suggestion :=')
end=source.index('\n\t}\n\tresult, err := edit.FixText',start)
source=source.replace('if len(file.Diagnostics()) != 0 {', 'if len(file.Diagnostics()) != 0 && !(len(fields) > 6 && fields[6] == "recovery") {')
source=source[:start]+'''
        fmt.Fprintf(out, "shape %d %d %s fixes=%d suggestions=%d\\n", start, end, d.Message.Id, len(d.Fixes), len(d.Suggestions))
        for _, fix := range d.Fixes {
            fmt.Fprintf(out, "fix %d %d\\t%s\\n", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
        }
        for _, suggestion := range d.Suggestions {
            fmt.Fprintf(out, "suggestion %s\\t%s\\n", suggestion.Message.Id, written(suggestion.Message.Description))
            for _, fix := range suggestion.Fixes {
                fmt.Fprintf(out, "edit %d %d\\t%s\\n", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
            }
        }'''+source[end:]
# The rule test API permits recovered invalid syntax. The production edit engine
# correctly rejects it. Preserve raw rule output only when there is no automatic
# fix to apply; valid sources still use the untouched converging fixer.
source=source.replace('\tresult, err := edit.FixText', '''
    if len(fields) > 6 && fields[6] == "recovery" {
        parsed := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName:path, Path:tspath.Path(path)}, source, core.ScriptKindTS)
        if len(parsed.Diagnostics()) != 0 {
            for _, diagnostic := range diagnostics { if len(diagnostic.Fixes) != 0 { panic("invalid recovered input offers an automatic fix") } }
            fmt.Fprintf(out, "fixed\\t%s\\n", written(source))
            return len(diagnostics)
        }
    }
    result, err := edit.FixText''',1)
(s/'oracle.go').write_text(source)
(s/'selection.go').write_text('package main\nimport "github.com/system-inc/cohere/internal/lint/rule"\ntype registered struct {subject rule.Rule; options func([]string) any}\nfunc registeredRules() []registered{return []registered{'+','.join('{oracle'+name+'(),oracle'+name+'Options}' for _,name in subjects)+'}}\n')
virtual=[(str(cohere/'adamic_owned_oracle.go'),str(s/'oracle.go')),(str(cohere/'adamic_owned_selection.go'),str(s/'selection.go'))]
virtual +=[(str(cohere/('adamic_owned_'+slug.replace('-','_')+'.go')),str(owned.parent/slug/'oracle.go')) for slug,_ in subjects]
(s/'oracle-overlay.json').write_text(json.dumps({'Replace':dict(virtual)}))
run(['go','build','-overlay='+str(s/'oracle-overlay.json'),'-o',s/'go-oracle']+[v for v,_ in virtual],'oracle-build',cohere)

# Capture every exercised upstream rule test input, preserving its filename suffix.
harness=cohere/'internal/lint/testing/rule_testing.go'
original=harness.read_text();anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
assert original.count(anchor)==1
(s/'capture.go').write_text(original.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\nRecordAssertedCase(t, result)\nreturn result',1))
(s/'capture-overlay.json').write_text(json.dumps({'Replace':{str(harness):str(s/'capture.go')}}))
env=os.environ.copy();env['COHERE_DOCS_CAPTURE']=str(s/'capture')
status,_,_=run(['go','test','-overlay='+str(s/'capture-overlay.json'),'./internal/lint/rules/core','-count=1','-v','-timeout=5m','-run','^Test('+ '|'.join(name for _,name in subjects)+')'],'upstream',cohere,env,True)
if status or not (s/'upstream.stdout.txt').read_text().rstrip().endswith('s') or '\nPASS\n' not in (s/'upstream.stdout.txt').read_text():
    raise SystemExit('upstream tests did not pass')
unique={}
for file in (s/'capture').glob('*.jsonl'):
    for line in file.read_text().splitlines():
        row={key.capitalize():value for key,value in json.loads(line).items()}
        key=(row['Rule'],row['File'],row['Source'])
        unique[key]=row
rows=[]
counts={slug:0 for slug,_ in subjects}
for i,(key,row) in enumerate(sorted(unique.items())):
    name=row['Rule'];slug=name
    if slug not in counts:continue
    suffix=Path(row['File']).suffix or '.ts'
    path=s/('case-'+str(i)+suffix);path.write_text(row['Source'])
    rows.append(str(path)+'\t'+name+'\t\t\tfalse\tnull\trecovery');counts[slug]+=1
for slug,_ in subjects:
    path=s/(slug+'-witness.ts');path.write_bytes((owned.parent/slug/'testdata/witness.ts.txt').read_bytes())
    rows.append(str(path)+'\t'+slug)
# Supplement cases absent from imported fixtures, including byte offsets and file gates.
manifest=s/'manifest.txt';manifest.write_text('\n'.join(rows)+'\n')
print('captured upstream cases '+json.dumps(counts),flush=True)

virtual_build=str(repo/'adamic_owned_build.go')
(s/'build-overlay.json').write_text(json.dumps({'Replace':{virtual_build:str(owned/'build.go.txt')}}))
def build(driver,directory,label):
    run(['go','run','-overlay='+str(s/'build-overlay.json'),virtual_build,driver,directory],label)
build(owned/'driver.a',s/'built','port-build')
node=['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs']
def commands(directory,driver):
    return [('Node',node+[driver]),('emitted',node+[directory/'driver.mjs']),('sanitized',[directory/'sanitized'])]
gaps=[];comparable=[]
for number,row in enumerate(rows):
    text=Path(row.split('\t')[0]).read_text()
    if text=='new Promise(@dec async () => {})':
        one=s/('gap-'+str(number)+'.manifest');one.write_text(row+'\n')
        run([s/'go-oracle','--manifest',one],'gap-'+str(number)+'-Go')
        for side,command in commands(s/'built',owned/'driver.a'):
            status,_,_=run(command+['--manifest',one],'gap-'+str(number)+'-'+side,allow_failure=True)
            stderr=(s/('gap-'+str(number)+'-'+side+'.stderr.txt')).read_text()
            if status!=70 or 'parser slice' not in stderr:raise SystemExit('decorator recovery gap was not an explicit refusal on '+side)
        gaps.append(row)
    else:comparable.append(row)
(s/'gaps.json').write_text(json.dumps(gaps,indent=2)+'\n')
manifest.write_text('\n'.join(comparable)+'\n')
print('EXPLICIT DECORATOR RECOVERY GAP '+str(len(gaps))+' rows; comparing '+str(len(comparable)),flush=True)
_,want,_=run([s/'go-oracle','--manifest',manifest],'go-output')
for side,command in commands(s/'built',owned/'driver.a'):
    _,observed,_=run(command+['--manifest',manifest],side+'-output')
    if observed!=want:
        position=next((i for i,(a,b) in enumerate(zip(want,observed)) if a!=b),min(len(want),len(observed)))
        raise SystemExit(f'{side} parity differs at byte {position}; see output logs')
    print(f'PASS {side}: {len(comparable)} cases, {len(want)} identical bytes including complete repairs',flush=True)

if args.typescript:
    pin=subprocess.check_output(['git','-C',str(args.typescript),'rev-parse','HEAD'],text=True).strip()
    assert pin=='050880ce59e30b356b686bd3144efe24f875ebc8'
    corpus=sorted((args.typescript/'src/compiler').rglob('*.ts'))
    corpus+=sorted((repo/'stage1').rglob('*.a'))+sorted((repo/'stage1').rglob('*.ts'))
    corpus=[p for p in corpus if '/.generated/' not in str(p) and '/gaps/' not in str(p)]
    full=s/'corpus-manifest.txt'
    full.write_text(''.join(str(path)+'\t'+slug+'\n' for path in corpus for slug,_ in subjects))
    _,want,_=run([s/'go-oracle','--manifest',full],'corpus-Go')
    for side,command in commands(s/'built',owned/'driver.a'):
        _,observed,_=run(command+['--manifest',full],'corpus-'+side)
        if observed!=want:raise SystemExit(side+' corpus parity differs')
    print(f'PASS corpus: {len(corpus)} files x 3 rules, {len(want)} identical bytes',flush=True)

if args.mutants:
    for slug,_ in subjects:
        directory=s/('mutant-'+slug);root=directory/'rules';root.mkdir(parents=True)
        for child in [slug for slug,_ in subjects]+['typescript-no-unused-expressions']:
            dest=root/child;dest.mkdir()
            for file in (owned.parent/child).glob('*.a'):
                content=file.read_text()
                def external(match):
                    spec=match.group(1);target=(file.parent/spec).resolve()
                    if not target.is_relative_to(owned.parent):return "from '"+str(target)+"'"
                    return match.group(0)
                content=re.sub(r"from '([^']+)'",external,content)
                (dest/file.name).write_text(content)
        mutation=json.loads((owned.parent/slug/'mutant.json').read_text())
        target=root/slug/mutation['file'];content=target.read_text()
        assert content.count(mutation['from'])==1
        target.write_text(content.replace(mutation['from'],mutation['to'],1))
        driver=root/owned.name/'driver.a';build(driver,directory/'built','build-'+mutation['name'])
        witness=s/(slug+'-witness.ts');one=directory/'manifest.txt';one.write_text(str(witness)+'\t'+slug+'\n')
        _,expected,_=run([s/'go-oracle','--manifest',one],mutation['name']+'-Go')
        for side,command in commands(directory/'built',driver):
            _,actual,_=run(command+['--manifest',one],mutation['name']+'-'+side)
            if actual==expected:raise SystemExit('mutant survived '+mutation['name']+' '+side)
            print('CAUGHT output-only '+mutation['name']+' '+side,flush=True)

if args.benchmark:
    for slug,_ in subjects:
        source={'no-async-promise-executor':'new Promise(async () => {const x = %d;});\n','no-case-declarations':'switch(x) {case %d: const a = 1; break;}\n','no-compare-neg-zero':'x%d === -0;\n'}[slug]
        path=s/(slug+'-planted.ts');path.write_text(''.join(source%i for i in range(1000)))
        manifest=s/(slug+'-benchmark.txt')
        compiler=sorted((args.typescript/'src/compiler').rglob('*.ts')) if args.typescript else []
        manifest.write_text(''.join(str(p)+'\t'+slug+'\n' for p in compiler+[path]))
        best={};expected=None
        for round in range(3):
            for side,command in [('Go',[s/'go-oracle']),('native',[s/'built/release']),('Node',node+[owned/'driver.a'])]:
                _,actual,duration=run(command+['--manifest',manifest,'--count'],slug+f'-round{round}-'+side)
                if expected is None:expected=actual
                assert actual==expected
                best[side]=min(best.get(side,duration),duration)
        count=int(expected)
        print('THROUGHPUT '+slug+' '+json.dumps({'findings':count,'files':len(compiler)+1,'per_second':{side:count/duration for side,duration in best.items()}}),flush=True)
print('PASS complete owned validation; see LANDING.md for shared harness coverage',flush=True)
