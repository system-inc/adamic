"""Focused gates and deliberate guard mutations, with persistent test logs."""
import json, os, pathlib, subprocess
root=pathlib.Path(__file__).resolve().parents[4]
evidence=pathlib.Path(__file__).resolve().parent
env=os.environ.copy();env['GOMAXPROCS']='4';env['ADAMIC_GATE_UNCACHED']='0'
def run(name,command,extra=None,expected=None):
 run_env=env.copy();run_env.update(extra or {})
 log=evidence/(name+'.log')
 with log.open('wb') as output:
  outcome=subprocess.run(command,cwd=root,env=run_env,stdout=output,stderr=output)
 data=log.read_text()
 if expected is None:
  if outcome.returncode:raise SystemExit(f'failed {log}')
 elif outcome.returncode==0 or expected not in data:
  raise SystemExit(f'wrong mutant failure {log}')
 print(f'{name}: '+('PASS' if expected is None else 'mutant caught: '+expected),flush=True)
base=['go','test','./stage1/cohere/lint','-count=1','-timeout','30m','-v']
run('parity',base+['-run','^(TestSplitWholeParity|TestSelectedRuleParity|TestSelectedRuleUncachedAnswers|TestWholeBuildSelection)$'],{'ADAMIC_NATIVE_SPLIT':'1'})
run('helper-parity',base+['-run','^TestSplitWholeParity$','-args','-rule-helper-change'])
run('cache-guards',base+['-run','^(TestLintCacheInvalidation|TestLintCacheMutants|TestLintCacheBypassAndIntegrity)$'])
run('ordinary-uncached-all-rule',base+['-run','^TestRulesAgree$'],{'ADAMIC_GATE_UNCACHED':'1','ADAMIC_NATIVE_SPLIT':'1'})
ordinary=(evidence/'ordinary-uncached-all-rule.log').read_text()
if 'split=false jobs=0' not in ordinary or 'split=true' in ordinary:raise SystemExit('ordinary witness used the wrong compiler mode')
run('filtered-oracle',['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(functions|closures|generic_functions)\\.a$','-count=1','-timeout','30m','-v'],{'ADAMIC_GATE_UNCACHED':'1','ADAMIC_NATIVE_SPLIT':'0'})
run('vet',['go','vet','./stage1/cohere/lint','./cmd/adamic-lint-check'])
run('native-split-guards',['go','test','./internal/native','-count=1','-v','-run','^(TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals|TestUnitSystemHeaderProvenance)$'])
run('native-mutants',['bash','internal/native/clang_units_evidence/run-mutants.sh'])
run('native-header-mutants',['bash','internal/native/clang_units_evidence/run-system-header-mutants.sh'])
for name in ('flags','state','literal','determinism'):
 (evidence/('native-'+name+'-mutant.log')).write_bytes(pathlib.Path('/tmp/adamic-clang-'+name+'-reproduced.log').read_bytes())
for name in ('flatten','header-key','flags','blanket-warning'):
 (evidence/('native-header-'+name+'-mutant.log')).write_bytes(pathlib.Path('/tmp/adamic-header-'+name+'-reproduced.log').read_bytes())
mutants = [('inherited-split-switch', 'split_test.go', 'func init() { os.Unsetenv("ADAMIC_NATIVE_SPLIT") }', 'func init() {}', '^TestWholeBuildSelection$', 'inherited split switch can replace the whole-file witness', {'ADAMIC_NATIVE_SPLIT': '1'}), ('uncached-selection', 'cache_test.go', ' && os.Getenv("ADAMIC_GATE_UNCACHED") != "1"', '', '^TestWholeBuildSelection$', 'uncached build options:', {}), ('parity-output', 'lint_test.go', 'source := native.C(lowered)', 'source := native.C(lowered)\n if !options.Split { source = "#include <stdio.h>\\n" + strings.Replace(source, "int main(int argc, char **argv) {", "int main(int argc, char **argv) {\\nputs(\\"whole mutant\\");", 1) }', '^TestSplitWholeParity/no-var$', 'split and whole-file outputs differ', {})]
for name,file,original,replacement,pattern,message,extra in mutants:
 source=(root/'stage1/cohere/lint'/file).read_text()
 if source.count(original)!=1:raise SystemExit('mutant anchor mismatch '+name)
 path=pathlib.Path('/tmp')/('adamic-lint-'+name+'.go');path.write_text(source.replace(original,replacement))
 overlay=path.with_suffix('.json');overlay.write_text(json.dumps({'Replace':{str(root/'stage1/cohere/lint'/file):str(path)}}))
 command=['go','test','-overlay='+str(overlay),'./stage1/cohere/lint','-count=1','-timeout','30m','-v','-run',pattern]
 run(name+'-mutant',command,extra,message)

# Observed compiler activity, independent of finding comparisons: removing the
# two goroutine launches must preserve findings but fail this overlap check.
def compilation_overlap(records):
 groups={}
 for record in records:
  if '-E' in record['argv']:
   groups.setdefault(str(pathlib.Path(record['cwd']).parent),[]).append(record['time'])
 if len(groups)!=2:return False
 windows=[(min(times),max(times)) for times in groups.values()]
 return max(start for start,end in windows)<min(end for start,end in windows)
for row in json.loads((evidence/'measurements.json').read_text()):
 if row['compiler']=='split' and row['mode']!='warm-unchanged':
  assert compilation_overlap(json.loads((evidence/row['trace']).read_text())),row['trace']
source=(root/'stage1/cohere/lint/rule_check_test.go').read_text()
assert source.count('go func() { defer pending.Done();')==2
path=pathlib.Path('/tmp/adamic-lint-serial-build.go')
path.write_text(source.replace('go func() { defer pending.Done();','func() { defer pending.Done();'))
overlay=path.with_suffix('.json');overlay.write_text(json.dumps({'Replace':{str(root/'stage1/cohere/lint/rule_check_test.go'):str(path)}}))
trace=pathlib.Path('/tmp/adamic-lint-split-measure/clang-events.jsonl');trace.write_text('')
wrapper=pathlib.Path('/tmp/adamic-lint-split-measure/bin/clang')
assert wrapper.is_file(),'run measure.py first to install the recording wrapper'
command=['go','test','-overlay='+str(overlay),'./stage1/cohere/lint','-run','^TestRule$','-count=1','-timeout','30m','-v','-args','-rule','no-var','-rule-byte-change']
run('serial-build-mutant',command,{'PATH':str(wrapper.parent)+':'+env['PATH']})
records=[json.loads(line) for line in trace.read_text().splitlines()]
(evidence/'serial-build-mutant.clang.json').write_text(json.dumps(records,indent=2)+'\n')
try:
 assert compilation_overlap(records),'correct and mutant compilation windows do not overlap'
except AssertionError as failure:
 with (evidence/'serial-build-mutant.log').open('a') as output:output.write('EXPECTED TRACE TEST FAILURE: '+str(failure)+'\n')
 print('serial-build: mutant caught: '+str(failure),flush=True)
else:raise SystemExit('serial-build mutant survived overlap test')
