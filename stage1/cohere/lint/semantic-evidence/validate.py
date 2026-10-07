"""Focused gates and a real native-output mutant of semantic parity."""
import json,os,pathlib,subprocess
root=pathlib.Path(__file__).resolve().parents[4];p=pathlib.Path(__file__).resolve().parent
env=os.environ.copy();env['GOMAXPROCS']='4';env['ADAMIC_GATE_UNCACHED']='0'
def run(name,command,expected=None,extra=None):
 e=env.copy();e.update(extra or {})
 with (p/(name+'.log')).open('wb') as output:r=subprocess.run(command,cwd=root,env=e,stdout=output,stderr=output)
 text=(p/(name+'.log')).read_text()
 if expected:
  assert r.returncode==1 and expected in text,(name,r.returncode,text)
 else:assert r.returncode==0,(name,r.returncode,text)
 print(name+': '+('PASS' if not expected else 'expected mutant failure caught'),flush=True)
run('normal-parity',['go','test','./stage1/cohere/lint','-run','^(TestSplitWholeParity|TestSelectedRuleParity|TestSelectedRuleUncachedAnswers|TestWholeBuildSelection)$','-count=1','-timeout','30m','-v'])
run('all-rule-uncached',['go','test','./stage1/cohere/lint','-run','^TestRulesAgree$','-count=1','-timeout','30m','-v'],extra={'ADAMIC_GATE_UNCACHED':'1','ADAMIC_NATIVE_SPLIT':'1'})
run('vet',['go','vet','./stage1/cohere/lint','./cmd/adamic-lint-check'])
source=(root/'stage1/cohere/lint/lint_test.go').read_text();anchor='source := native.C(lowered)';assert source.count(anchor)==1
replacement=anchor+'\n if *ruleSemanticChange { source = "#include <stdio.h>\\n" + strings.Replace(source, "int main(int argc, char **argv) {", "int main(int argc, char **argv) {\\nputs(\\\"native parity mutant\\\");", 1) }'
mutant=pathlib.Path('/tmp/adamic-lint-semantic-native-mutant.go');mutant.write_text(source.replace(anchor,replacement))
overlay=mutant.with_suffix('.json');overlay.write_text(json.dumps({'Replace':{str(root/'stage1/cohere/lint/lint_test.go'):str(mutant)}}))
run('native-parity-mutant',['go','test','-overlay='+str(overlay),'./stage1/cohere/lint','-run','^TestRule$','-count=1','-timeout','30m','-v','-args','-rule','no-var','-rule-semantic-change'],expected='semantic edit Node/native outputs differ')
# The source itself is a semantic mutant: all nine samples must be rejected by
# the independent Go oracle after the three port runtimes agree.
rows=json.loads((p/'measurements.json').read_text());assert len(rows)==27
for row in rows:
 if row['mode']=='warm-semantic':
  text=(p/row['log']).read_text()
  assert row['compiled_objects']>0 and row['exit']==1
  assert 'EXPECTED SEMANTIC EDIT FAILURE: Node JavaScript native agree and disagree with Go' in text
  assert 'semantic edit Node JavaScript native byte-identical:' in text
print('nine semantic edits changed native objects, agreed on all port runtimes and were rejected by Go',flush=True)
