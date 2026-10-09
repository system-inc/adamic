from pathlib import Path
import json,subprocess,time
root=Path(__file__).resolve().parents[3]
evidence=Path(__file__).resolve().parent
selector='internal/lower/program_region_selection.go'
output='internal/lower/program_region.go'
cases=[
 ('drop-Node',selector,'\treturn selected\n}', '\tfor node := range selected { if node.proven != nil && f.l.checker.TypeToStringEx(node.proven, nil, checker.TypeFormatFlagsNoTruncation, nil) == "Node" { delete(selected,node) } }\n\treturn selected\n}',True,'TestProgramRegionTscNodeMemberSites'),
 ('drop-NodeArray',selector,'\treturn selected\n}', '\tfor node := range selected { if node.proven != nil { label := f.l.checker.TypeToStringEx(node.proven, nil, checker.TypeFormatFlagsNoTruncation, nil); if label == "NodeArray" || (len(label)>10 && label[:10]=="NodeArray<") { delete(selected,node) } } }\n\treturn selected\n}',True,'TestProgramRegionTscNodeMemberSites'),
 ('drop-host-callback-cell',output,'plan.cells[i] = x.ProgramRegion','plan.cells[i] = x.ProgramRegion && x.Name != "callback"',True,'TestProgramRegionHost14MemberSites'),
 ('select-acyclic-type',selector,'\treturn selected\n}', '\tfor _, node := range candidates { if node.proven != nil { selected[node] = true; break } }\n\treturn selected\n}',True,'TestProgramRegionScoutParseSelectionEmpty'),
]
results=[]
for name,file,before,after,caught,test in cases:
 source=(root/file).read_text();assert source.count(before)==1,(name,source.count(before))
 changed=evidence/(name+'.go.txt');changed.write_text(source.replace(before,after))
 overlay=evidence/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(changed)}}))
 command=['timeout','90','go','test','-overlay='+str(overlay),'./internal/lower','-run','^'+test+'$','-count=1','-timeout','60s','-v']
 start=time.monotonic();run=subprocess.run(command,cwd=root,capture_output=True,text=True,timeout=95)
 log=run.stdout+run.stderr;(evidence/'logs'/(name+'.log')).write_text(log)
 failed=run.returncode!=0 and '--- FAIL: '+test in log
 assert failed==caught,(name,run.returncode,log)
 assert caught or (run.returncode==0 and '--- PASS: '+test in log),(name,log)
 results.append({'mutant':name,'caught':failed,'expected_caught':caught,'exit':run.returncode,'seconds':time.monotonic()-start,'command':command})
 print(name,'caught' if failed else 'survived',flush=True)
(evidence/'mutants-results.json').write_text(json.dumps(results,indent=2)+'\n')
