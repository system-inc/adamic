from pathlib import Path
import json,collections,re
r=Path(__file__).resolve().parent
rows=json.loads((r/'main-results.json').read_text())+json.loads((r/'supplement-results.json').read_text())
fixes={132:('literal-key reads','fx6-key-read'),133:('named callable producer','fx7-callable-producers'),134:('unrelated same-name fields','fx6-valid-programs'),135:('assignable callable producers','fx7-callable-producers'),136:('destructured contracts','fx6-valid-programs'),137:('tuple length object view','fx6-valid-programs (narrow refusal only)'),138:('tuple object recognition','fx7-wrong-aborts'),139:('read diagnostic locations','fx6-key-read routing; fx6-cleanup tests'),140:('checked failure type text','fx7-depth-texts'),141:('dead code/comments','fx6-cleanup'),142:('unsupported descriptor loses fields','fx7-unsupported-view'),143:('snapshot reference payload','fx7-snapshot'),144:('maybe number/boolean union reads','fx7-scalar-union-view'),145:('non-closure producer registry','fx7-callable-producers'),146:('checked cast/callable metadata','fx7-wrong-aborts'),147:('undefined writes to same-named fields','fx6-valid-programs; fx7-name-family coverage'),148:('plain undefined reads','fx6-valid-programs; fx7-name-family coverage'),149:('nullable array/Map writes','fx6-valid-programs; fx7-name-family coverage'),150:('runtime-created entry tuples','none in named branches'),151:('recursive membership depth','fx7-depth-texts'),152:('failure runtime-kind text','fx7-depth-texts'),153:('remaining cleanup','fx6-cleanup'),154:('callable union result representation','fx7-callable-union-result')}
table=['| Item | Current main versus source Node | Historical fixing branch (`compiler/`) |','|---|---|---|']
for i,(label,branch) in fixes.items():
 group=[x for x in rows if x['item']==i]
 different=sum(not x['agrees'] for x in group)
 observation=f'Yes, {different}/{len(group)} probes differ; {label}' if group else 'N/A, cleanup with no executable probe'
 if i==139: observation+=' (located refusal required, rather than acceptance)'
 table.append(f'| {i} | {observation} | {branch} |')
(r/'TABLE.md').write_text('\n'.join(table)+'\n')
leaf=re.findall(r'--- PASS: (TestLoomView\S+) \(([^)]+)\)',(r/'final-focused.log').read_text())
(r/'leaf-seconds.json').write_text(json.dumps(dict(leaf),indent=2)+'\n')
