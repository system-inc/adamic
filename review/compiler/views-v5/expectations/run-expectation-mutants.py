import difflib,json,os,pathlib,subprocess,time
root=pathlib.Path.cwd(); evidence=root/'review/compiler/views-v5/expectations'; results=[]
line='return &Refused{Where: l.program.Where(node), What: "checked intersection view with unsupported member " + path + " (" + family + ")", Fix: "prove or implement the complete member contract before creating this view"}'
cases=[
 ('whole-view-guard','internal/lower/interface_cast.go','if err := l.checkCompleteIntersectionView(node, id, l.checker.TypeToString(target), map[ir.ViewContractID]bool{}); err != nil {','if err := l.checkCompleteIntersectionView(node, id, l.checker.TypeToString(target), map[ir.ViewContractID]bool{}); false && err != nil {','^TestCheckedViewIntersectionCompoundDemand$'),
 ('refusal-kind','internal/lower/view_intersections.go',line,line.replace('&Refused{','&NotYet{'),'^TestCheckedViewIntersectionCompoundDemand$'),
 ('refusal-location','internal/lower/view_intersections.go',line,line.replace('l.program.Where(node)','l.program.Where(node) + ":mutant"'),'^TestCheckedViewIntersectionCompoundDemand$'),
 ('refusal-member','internal/lower/view_intersections.go',line,line.replace('+ path +','+ path + ".mutant" +'),'^TestCheckedViewIntersectionCompoundDemand$'),
 ('refusal-fix','internal/lower/view_intersections.go',line,line.replace('prove or implement the complete member contract before creating this view','mutant fix'),'^TestCheckedViewIntersectionCompoundDemand$'),
 ('notyet-location','internal/lower/diagnostics.go','return &NotYet{Where: l.program.Where(node), What: what}','return &NotYet{Where: l.program.Where(node) + ":mutant", What: what}','^TestCheckedViewIntersectionDeferredMember$/^read$'),
 ('notyet-alias-kind','internal/lower/diagnostics.go','return &NotYet{Where: l.program.Where(node), What: what}','return &Refused{Where: l.program.Where(node), What: what}','^TestCheckedViewIntersectionSource$/^emit-root-wrong$'),
]
for name,filename,old,new,pattern in cases:
 p=root/filename; original=p.read_bytes(); text=original.decode(); assert text.count(old)==1,name; mutated=text.replace(old,new,1)
 (evidence/(name+'.patch')).write_text(''.join(difflib.unified_diff(text.splitlines(True),mutated.splitlines(True),fromfile='a/'+filename,tofile='b/'+filename)))
 start=time.monotonic()
 try:
  p.write_text(mutated)
  command=['go','test','./internal/oracle','-run',pattern,'-count=1','-timeout=80s','-v']
  with (evidence/(name+'.log')).open('w') as out:
   code=subprocess.run(command,env=dict(os.environ,ADAMIC_INTERSECTION_ORIGINAL_DECLS='/tmp/views-v5-intersection-declarations'),stdout=out,stderr=subprocess.STDOUT,timeout=88).returncode
  output=(evidence/(name+'.log')).read_text()
  caught=code!=0 and 'want ' in output and 'got ' in output and 'build failed' not in output
  results.append(dict(mutant=name,command=command,exit=code,caught=caught,seconds=round(time.monotonic()-start,3)))
  print(json.dumps(results[-1]),flush=True)
  assert caught,name
 finally: p.write_bytes(original)
(evidence/'expectation-mutant-results.json').write_text(json.dumps(results,indent=2)+'\n')
