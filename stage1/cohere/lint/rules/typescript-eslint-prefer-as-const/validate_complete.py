#!/usr/bin/env python3
"""Owned full-repair oracle; changes no shared Adamic harness or registry."""
import json, os, subprocess, tempfile, shutil, time
from pathlib import Path
root=Path(__file__).resolve().parents[5]
owned=Path(__file__).resolve().parent
lint=root/'stage1/cohere/lint'
cohere=(root/'cohere').resolve()
scratch=Path(tempfile.mkdtemp(prefix='wave14-complete-'))
evidence=owned/'evidence';evidence.mkdir(exist_ok=True)
log=(evidence/'complete.log').open('w',buffering=1)
rules=[('no-unnecessary-type-constraint','NoUnnecessaryTypeConstraint'),('prefer-as-const','PreferAsConst'),('prefer-enum-initializers','PreferEnumInitializers')]
def note(s):print(s,file=log,flush=True)
def run(args,cwd=root,env=None,output=None):
 target=output or scratch/('stdout-'+str(time.time_ns()))
 errors=scratch/('stderr-'+str(time.time_ns()))
 start=time.perf_counter()
 with target.open('wb') as out,errors.open('wb') as err:
  p=subprocess.run([str(a) for a in args],cwd=cwd,env=env,stdout=out,stderr=err)
 elapsed=time.perf_counter()-start
 if p.returncode or errors.stat().st_size:
  note('FAILED '+repr([str(a) for a in args])+f' exit={p.returncode}')
  note(errors.read_text());note(target.read_text()[:6000]);raise RuntimeError('command failed')
 data=target.read_bytes()
 if output is None:target.unlink()
 errors.unlink()
 return data,elapsed
try:
 note('scratch='+str(scratch))
 # Instrument copies of upstream test files, not the shared test harness.
 replacements={}
 for name,var in rules:
  original=cohere/'internal/lint/rules/typescript'/ (name.replace('-','_')+'_test.go')
  copy=scratch/original.name
  copy.write_text(original.read_text().replace('rule_testing.Run(', 'wave14CaptureRun('))
  replacements[str(original)]=str(copy)
 capture_helper=scratch/'capture_test.go'
 capture_helper.write_text('''package typescript
import ("testing"; "github.com/system-inc/cohere/internal/lint/rule"; rule_testing "github.com/system-inc/cohere/internal/lint/testing")
func wave14CaptureRun(t *testing.T, subject rule.Rule, fileName, source string) rule_testing.Result {
 result:=rule_testing.Run(t,subject,fileName,source)
 rule_testing.RecordAssertedCase(t,result)
 return result
}
''')
 replacements[str(cohere/'internal/lint/rules/typescript/wave14_capture_test.go')]=str(capture_helper)
 overlay=scratch/'capture-overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
 capture=scratch/'capture';env=dict(os.environ,COHERE_DOCS_CAPTURE=str(capture))
 out,elapsed=run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/typescript','-run','^('+'|'.join('Test'+v for _,v in rules)+')','-count=1','-v','-timeout=10m'],cwd=cohere,env=env,output=evidence/'upstream-complete.log')
 note(f'original upstream assertions PASS {elapsed:.3f}s')
 rows_by_name={name:[] for name,_ in rules};seen=set()
 for file in sorted(capture.glob('*.jsonl')):
  for line in file.read_text().splitlines():
   row=json.loads(line);key=(row['rule'],row['file'],row['source'])
   if key in seen:continue
   seen.add(key);name=row['rule'].split('/')[-1]
   path=scratch/f'case-{len(seen)}{Path(row["file"]).suffix}'
   path.write_text(row['source'])
   rows_by_name[name].append(f'{path}\t{row["rule"]}\t{row["file"]}')
 for name,var in rules:
  witness=lint/'rules'/('typescript-eslint-'+name)/'testdata/witness.ts.txt'
  semantic='/repository/witness.tsx' if name=='no-unnecessary-type-constraint' else '/repository/witness.ts'
  rows_by_name[name].append(f'{witness}\t@typescript-eslint/{name}\t{semantic}')
 # Build owned compiler helper and independent upstream oracle through virtual files.
 def build_go(source,virtual,cwd,target):
  overlay=scratch/(target.name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(cwd/virtual):str(source)}}))
  run(['go','build','-overlay='+str(overlay),'-o',target,cwd/virtual],cwd=cwd)
 builder=scratch/'builder';build_go(owned/'build.go.txt','wave14_owned_build.go',root,builder)
 oracle=scratch/'oracle';build_go(owned/'complete_oracle.go.txt','wave14_owned_oracle.go',cohere,oracle)
 runner=owned/'complete_runner.a';binary=scratch/'native';js=scratch/'emitted.mjs'
 run([builder,runner,binary,js,'sanitized'])
 node=['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs']
 def manifest(rows,label):
  p=scratch/(label+'.manifest');p.write_text('\n'.join(rows)+'\n');return p
 def compare(rows,label,source=runner,native=binary,emitted=js):
  p=manifest(rows,label);want,_=run([oracle,p])
  for side,args in [('Node',node+[source,p]),('native',[native,p]),('emitted JavaScript',node+[emitted,p])]:
   got,_=run(args)
   if got!=want:
    a=got.decode().splitlines();b=want.decode().splitlines()
    for i in range(max(len(a),len(b))):
     left=a[i] if i<len(a) else '<EOF>';right=b[i] if i<len(b) else '<EOF>'
     if left!=right:note(f'{label} {side} DIFFERENCE line {i+1}: {left!r} != {right!r}');break
    raise RuntimeError('comparison failed')
  count,_=run([oracle,p,'--count']);note(f'{label}: all four identical {len(want)} bytes; findings={count.decode().strip()}; files={len(rows)}')
  return p,int(count)
 for name,var in rules:compare(rows_by_name[name],'fixtures-'+name)
 compiler=Path(os.environ.get('ADAMIC_TYPESCRIPT_SOURCE','/tmp/lint-wave1-14-typescript-pinned'))
 pin,_=run(['git','rev-parse','HEAD'],cwd=compiler)
 assert pin.decode().strip()=='050880ce59e30b356b686bd3144efe24f875ebc8'
 files=sorted((compiler/'src/compiler').rglob('*.ts'))+sorted(p for p in Path(os.environ.get('ADAMIC_STAGE1_CORPUS',str(root/'stage1'))).rglob('*') if p.suffix in ('.a','.ts'))
 note(f'corpus files={len(files)} compiler=77 stage1={len(files)-77}')
 corpus={name:[f'{p}\t@typescript-eslint/{name}\t{p}' for p in files] for name,_ in rules}
 for name,var in rules:compare(corpus[name],'corpus-'+name)
 for name,var in rules:
  variant=scratch/('mutant-'+name)
  shutil.copytree(root/'stage1',variant/'stage1',ignore=shutil.ignore_patterns('evidence','.generated'))
  changed=variant/'stage1/cohere/lint/rules'/('typescript-eslint-'+name)
  mutation=json.loads((changed/'mutant.json').read_text())
  file=changed/mutation['file'];text=file.read_text();assert text.count(mutation['from'])==1
  file.write_text(text.replace(mutation['from'],mutation['to']))
  source=variant/'stage1/cohere/lint/rules/typescript-eslint-prefer-as-const/complete_runner.a'
  mutant_binary=scratch/('mutant-'+name+'-native');mutant_js=scratch/('mutant-'+name+'.mjs')
  run([builder,source,mutant_binary,mutant_js,'sanitized'])
  p=manifest([rows_by_name[name][-1]],'mutant-'+name);want,_=run([oracle,p])
  for side,args in [('Node',node+[source,p]),('native',[mutant_binary,p]),('emitted JavaScript',node+[mutant_js,p])]:
   got,_=run(args);assert got!=want,f'mutant survived {name} {side}'
   note(f'{name} {mutation["name"]}: compiling exit-0 mutant caught only by comparison on {side}')
 release=scratch/'release';run([builder,runner,release,scratch/'release.mjs','release'])
 for name,var in rules:
  rows=corpus[name];p=manifest(rows,'timing-'+name);count,_=run([oracle,p,'--count']);note(f'{name} natural findings={count.decode().strip()}')
  if int(count)==0:
   witness=lint/'rules'/('typescript-eslint-'+name)/'testdata/witness.ts.txt'
   positive=scratch/(name+'-positive.tsx');positive.write_text(witness.read_text()*500)
   rows=rows+[f'{positive}\t@typescript-eslint/{name}\t{positive}'];p=manifest(rows,'timing-'+name)
   note('timing uses 500 witness copies as a synthetic supplement')
  best={};expected=None
  for round in range(5):
   for side,args in [('Go',[oracle,p,'--count']),('native',[release,p,'--count']),('Node',node+[runner,p,'--count'])]:
    got,elapsed=run(args)
    if expected is None:expected=got
    assert got==expected
    best[side]=min(best.get(side,float('inf')),elapsed)
    note(f'{name} round={round+1} {side} seconds={elapsed:.6f} count={got.decode().strip()}')
  for side in ('native','Node','Go'):note(f'{name} {side} {int(expected)/best[side]:.2f} findings/s best={best[side]:.6f}s count={int(expected)} files={len(rows)}')
 note('PASS complete owned four-way repair comparison, corpus, mutants and throughput')
except Exception as error:
 note('FAIL '+repr(error));raise
finally:log.close()
