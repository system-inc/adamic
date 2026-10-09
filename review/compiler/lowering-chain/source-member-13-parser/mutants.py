"""Replay parser compiler mutations independently; builds are not catchers."""
import json,os,pathlib,subprocess,time
root=pathlib.Path.cwd()
out=pathlib.Path(__file__).resolve().parent
cases=[]
for name,before,after in [
 ('branch-union','yes.written[field] && no.written[field]','yes.written[field] || no.written[field]'),
 ('omit-escape-missing','a.result.BeforeEscape[field] = false','a.result.BeforeEscape[field] = true'),
 ('erase-earlier-read','Checked: !state.written[node.Name().Text()] || state.escaped','Checked: false'),
 ('trust-assertion','return !forged && interfaceScalar(actual)','return (forged || !forged) && interfaceScalar(actual)')]:
 cases.append((name,'internal/lower/factory_completion.go',before,after,'./internal/lower','^TestParserFactory(Completion.*|ReadBeforeCompletion)$'))
for field in ['pos','end','hasTrailingComma','transformFlags']:
 cases.append(('slot-'+field,'internal/lower/node_array_layout.go','return slot, nil','if field == "'+field+'" { slot.Slot++ }; return slot, nil','./internal/lower','^TestParserNodeArrayLayout$'))
cases.extend([
 ('drop-unset-use','internal/lower/factory_unset.go','b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})','_ = matches; _ = message','./internal/oracle','^TestParserConstructionUnsetUse$'),
 ('drop-speculation-check','internal/lower/speculation.go','To: held, Checked: true, Message: message','To: held, Checked: false, Message: message','./internal/oracle','^TestParserSpeculationMisfit$'),
 ('disable-tiny-stack','internal/native/runtime/stack.c','base - size + reserved : UINTPTR_MAX','base - size + reserved : 0','./internal/oracle','^TestParserStackTinyLimit$'),
])
results=[]
for name,file,before,after,package,selector in cases:
 data=(root/file).read_text();assert data.count(before)==1,(name,data.count(before))
 extension='.go.txt' if file.endswith('.go') else '.c.txt'
 changed=out/(name+extension);changed.write_text(data.replace(before,after))
 overlay=out/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(changed)}}))
 command=['go','test','-overlay='+str(overlay),package,'-run',selector,'-count=1','-v']
 start=time.monotonic();log=out/'logs'/(name+'.log')
 with log.open('w') as stream:code=subprocess.run(command,cwd=root,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=stream,stderr=subprocess.STDOUT).returncode
 text=log.read_text();caught=code!=0 and '--- FAIL: TestParser' in text and '[build failed]' not in text and 'clang failed' not in text
 results.append(dict(mutant=name,command=command,caught=caught,exit=code,seconds=time.monotonic()-start));print(name,'caught' if caught else 'NOT CAUGHT',flush=True)
(out/'mutants-results.json').write_text(json.dumps(results,indent=2)+'\n');assert all(x['caught'] for x in results)
