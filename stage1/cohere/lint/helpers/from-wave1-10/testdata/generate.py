"""Parse every consumer's Go test strings, then generate independent scanner controls."""
import json,subprocess,itertools
from pathlib import Path
here=Path(__file__).resolve().parent
root=here.parents[5]
ledger=json.loads((here.parents[1]/'readiness.json').read_text())
symbol='github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.addWhitespaceAroundMathOperators'
rules=[r['rule'] for r in ledger['remaining'] if symbol in r['remaining_helpers']]
files=list((root/'cohere/internal/lint/rules/tailwind').glob('*_test.go'))
# The actual rule descriptor's exported name identifies its upstream test file.
names={'enforce-canonical-classes':'enforce_canonical_classes_test.go','enforce-consistent-class-order':'enforce_consistent_class_order_test.go','enforce-consistent-variant-order':'enforce_consistent_variant_order_test.go','enforce-shorthand-classes':'enforce_shorthand_classes_test.go','no-conflicting-classes':'no_conflicting_classes_test.go','no-unknown-classes':'no_unknown_classes_test.go'}
selected=[root/'cohere/internal/lint/rules/tailwind'/names[r.split('/')[1]] for r in rules]
for f in selected: assert f.exists(), f
scratch=Path('/tmp/wave10-helper-literals.go');scratch.write_text((here/'literals.go.txt').read_text())
binary='/tmp/wave10-helper-literals'
subprocess.run(['go','build','-o',binary,str(scratch)],cwd=root,check=True)
raw=json.loads(subprocess.check_output([binary,*map(str,selected)],cwd=root))
values=[]; coverage=[]
for rule,file in zip(rules,selected):
 strings=raw[str(file)]
 coverage.append({'rule':rule,'file':str(file.relative_to(root)),'literals':len(strings)})
 values.extend(strings)
 # Exercise scanner branches even when the original class string contains no math function.
 values.extend('calc('+s+')' for s in strings)
functions=['calc','min','max','clamp','mod','rem','sin','cos','tan','asin','acos','atan','atan2','pow','sqrt','hypot','log','exp','round','CALC','foo','foo-calc','mycalc','']
operands=['0','1px','2%','-3.4e-2','3E+4','var(--x)','(2+3)','😀','é','\u00a0','\u0085','\ufeff','a','e','',')','--x','1abc','1aé','calc(2+3)']
for name,left,op,right in itertools.product(functions,operands,['+','-','*','/',','],operands): values.append(name+'('+left+op+right+')')
values.extend(['calc(1,  2)','calc(1\t+2)','calc(1\n+2)','calc((1+2)*3)','calc(1+var(--x-y))','calc(1 + 2)','calc(1\r\n-2)','calc(1é-2)','calc(1😀-2)', 'calc(1e -2)', 'calc(1e\u00a0-2)', 'calc(1,)', 'calc)1+2(', 'calc(1) +2', 'calc(1%+2)', 'calc(é+é)', 'remnant'])
values=list(dict.fromkeys(values))
(here/'cases.json').write_text(json.dumps(values,ensure_ascii=True)+'\n')
(here/'coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
print('cases',len(values),'consumers',len(coverage),'literals',sum(r['literals'] for r in coverage))
