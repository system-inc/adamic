import json, pathlib, subprocess, time
out=pathlib.Path(__file__).resolve().parent
root=out.parents[2]
cases=[
 ('plain-placeholder-map','internal/ir/ir.go','PlaceholderSources SortedMap[string, string]','PlaceholderSources map[string]string','./internal/ir','^TestEveryMapAProgramHoldsIsSorted$','is a plain'),
 ('placeholder-physical-path','internal/lower/placeholder.go','l.program.Where(declaration) + ":" + symbol.Name','ast.GetSourceFileOfNode(declaration).FileName().AsString() + ":" + symbol.Name','./internal/oracle','^TestParserPlaceholderProductsUseDisplayedPaths$','parser product changed'),
 ('factory-physical-path','internal/lower/factory_unset.go','l.program.Where(node)','ast.GetSourceFileOfNode(node).FileName().AsString()','./internal/oracle','^TestParserFactoryProductsUseDisplayedPaths$','parser product changed'),
 ('speculation-physical-path','internal/lower/speculation.go','l.program.Where(call)','ast.GetSourceFileOfNode(call).FileName().AsString()','./internal/oracle','^TestParserSpeculationProductsUseDisplayedPaths$','parser product changed'),
]
results=[]
for name,file,before,after,pkg,test,catcher in cases:
 source=(root/file).read_text();assert source.count(before)==1
 changed=out/(name+'.go.txt');changed.write_text(source.replace(before,after))
 overlay=out/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(changed)}}))
 command=['go','test','-overlay='+str(overlay),pkg,'-run',test,'-count=1','-v','-timeout','90s']
 start=time.monotonic();log=out/'logs'/(name+'.log')
 with log.open('w') as stream:result=subprocess.run(command,cwd=root,stdout=stream,stderr=subprocess.STDOUT,timeout=120)
 text=log.read_text();caught=result.returncode==1 and catcher in text and '[build failed]' not in text
 results.append(dict(mutant=name,command=command,exit=result.returncode,caught=caught,seconds=time.monotonic()-start,catcher=catcher));print(name,caught,flush=True)
 (out/'product-mutants-results.json').write_text(json.dumps(results,indent=2)+'\n')
 if not caught:raise SystemExit(1)
