#!/usr/bin/env python3
import json,os,re,random,subprocess,tempfile,shutil,time
from pathlib import Path
owned=Path(__file__).resolve().parent
root=owned.parents[4];cohere=root/'cohere'
scratch=Path(tempfile.mkdtemp(prefix='helper14-'));evidence=owned/'evidence';evidence.mkdir(exist_ok=True)
log=(evidence/'comparison.log').open('w',buffering=1)
def note(s):print(s,file=log,flush=True)
def run(args,cwd=root,env=None,target=None,allow_failure=False):
 out=target or scratch/('out-'+str(time.time_ns()));err=scratch/('err-'+str(time.time_ns()))
 with out.open('wb') as o,err.open('wb') as e:p=subprocess.run(list(map(str,args)),cwd=cwd,env=env,stdout=o,stderr=e)
 if not allow_failure and (p.returncode or err.stat().st_size):
  note('FAILED '+repr(args)+f' exit={p.returncode}');note(err.read_text());note(out.read_text()[:5000]);raise RuntimeError('command failed')
 return out.read_bytes(),p.returncode,err.read_bytes()
def overlay(mapping,name):
 p=scratch/(name+'.json');p.write_text(json.dumps({'Replace':{str(k):str(v) for k,v in mapping.items()}}));return p
try:
 note('scratch='+str(scratch))
 original=cohere/'internal/lint/rules/tailwind/collapse/variant.go'
 instrumented=scratch/'variant.go';text=original.read_text();anchor='func leadingInteger(value string) (int, bool) {';assert text.count(anchor)==1
 instrumented.write_text(text.replace(anchor,anchor+'\n slot14Capture(value)'))
 capture_overlay=overlay({original:instrumented,cohere/'internal/lint/rules/tailwind/collapse/slot14_capture.go':owned/'capture.go.txt'},'capture')
 tests=[('enforce-canonical-classes','Test(EnforceCanonicalClasses|Canonical)'),('enforce-consistent-class-order','TestEnforceConsistentClassOrder'),('enforce-consistent-variant-order','TestEnforceConsistentVariantOrder'),('enforce-shorthand-classes','TestEnforceShorthandClasses'),('no-conflicting-classes','Test(NoConflictingClasses|ConflictingClassesPropose)'),('no-unknown-classes','Test(NoUnknownClasses|UnknownIgnore|KnownRootWithUnknown|UnknownClassFixtures)')]
 inputs=[];coverage={}
 for name,pattern in tests:
  capture=scratch/(name+'.jsonl');env=dict(os.environ,SLOT14_CAPTURE=str(capture))
  output,code,errors=run(['go','test','-overlay='+str(capture_overlay),'./internal/lint/rules/tailwind','-run','^'+pattern,'-count=1','-v','-timeout=10m'],cwd=cohere,env=env,target=evidence/(name+'.log'),allow_failure=True)
  live=[json.loads(l) for l in capture.read_text().splitlines()] if capture.exists() else []
  # Every consumer test file contributes all raw string literals as source controls.
  testfile=cohere/'internal/lint/rules/tailwind'/(name.replace('-','_')+'_test.go')
  source=testfile.read_text();raw=re.findall(r'`([^`]*)`',source)
  tokens=[m.group(0) for value in raw for m in re.finditer(r'[+\-]?[0-9]+(?:[.][0-9]+)?[A-Za-z%]*',value)]
  controls=raw+tokens;inputs+=live+controls
  coverage[name]={'upstream_exit':code,'live_calls':len(live),'source_controls':len(controls),'source':str(testfile.relative_to(root))}
  note(name+f': upstream exit={code}, actual leadingInteger calls={len(live)}, consumer source controls={len(controls)}')
 boundaries=['','+','-','0','-0','+40rem',' \t\n\r-40rem','\v40','\f40','\u00a040','\ufeff40','9223372036854775807','9223372036854775808','-9223372036854775808','18446744073709551615','18446744073709551616','-18446744073709551617','9'*1000,'00000000000000000000040rem','40.5rem','0x10','１２','40😀']
 rng=random.Random(1407)
 controls=boundaries+[rng.choice(['','+','-',' \t','\v','\u00a0'])+''.join(rng.choice('0123456789') for _ in range(rng.randrange(1,130)))+rng.choice(['rem','px','','.5','x']) for _ in range(4000)]
 inputs+=controls
 inputfile=scratch/'inputs.json';inputfile.write_text(json.dumps(inputs,ensure_ascii=True))
 (evidence/'coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
 oracle_output=scratch/'go-output';env=dict(os.environ,SLOT14_INPUT=str(inputfile),SLOT14_OUTPUT=str(oracle_output))
 oracle_overlay=overlay({cohere/'internal/lint/rules/tailwind/collapse/slot14_oracle_test.go':owned/'oracle_test.go.txt'},'oracle')
 run(['go','test','-overlay='+str(oracle_overlay),'./internal/lint/rules/tailwind/collapse','-run','^TestSlot14LeadingInteger$','-count=1','-timeout=10m'],cwd=cohere,env=env,target=evidence/'oracle.log')
 builder=scratch/'builder';build_overlay=overlay({root/'slot14_build.go':owned/'build.go.txt'},'build')
 run(['go','build','-overlay='+str(build_overlay),'-o',builder,root/'slot14_build.go'])
 runner=scratch/'runner.a';runner.write_text((owned/'withdrawn_runner.a.txt').read_text().replace('../options_json.ts','./options_json.ts'))
 shutil.copy2(owned.parent/'options_json.ts',scratch/'options_json.ts')
 (scratch/'leading_integer.a').write_text((owned/'withdrawn_leading_integer.a.txt').read_text())
 native=scratch/'native';js=scratch/'emitted.mjs'
 run([builder,runner,native,js,'sanitized'])
 node=['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs']
 want=oracle_output.read_bytes()
 for side,args in [('Node',node+[runner,inputfile]),('native',[native,inputfile]),('emitted JavaScript',node+[js,inputfile])]:
  got,_,_=run(args);assert got==want,(side,'byte mismatch');note(side+f': exact {len(inputs)} inputs, {len(want)} bytes')
 shutil.copy2(owned.parent/'options_json.ts',scratch/'options_json.ts')
 mutant=scratch/'mutant';mutant.mkdir();(mutant/'runner.a').write_text(runner.read_text().replace('./options_json.ts','../options_json.ts'))
 spec=json.loads((owned/'mutant.json').read_text());text=(owned/'withdrawn_leading_integer.a.txt').read_text();assert text.count(spec['from'])==1
 (mutant/'leading_integer.a').write_text(text.replace(spec['from'],spec['to']))
 mutant_native=scratch/'mutant-native';mutant_js=scratch/'mutant.mjs'
 run([builder,mutant/'runner.a',mutant_native,mutant_js,'sanitized'])
 for side,args in [('Node',node+[mutant/'runner.a',inputfile]),('native',[mutant_native,inputfile]),('emitted JavaScript',node+[mutant_js,inputfile])]:
  got,_,_=run(args);assert got!=want;note(side+': compiling exit-zero clean-stderr sign mutant caught only by Go comparison')
 note('PASS bounded helper four-way comparison and semantic mutant')
 gaps=[name for name,row in coverage.items() if row['upstream_exit']!=0 or row['live_calls']==0]
 note('Unpassed upstream/live consumer coverage: '+repr(gaps))
except Exception as e:note('FAIL '+repr(e));raise
finally:log.close()
