import json,pathlib
p=pathlib.Path(__file__).parent
def covered(name):
 lines=set()
 for row in (p/(name+'.cover')).read_text().splitlines()[1:]:
  loc,statements,count=row.split()
  if int(count)==0:continue
  file,span=loc.split(':');start,end=span.split(',')
  for line in range(int(start.split('.')[0]),int(end.split('.')[0])+1):lines.add(file+':'+str(line))
 return lines
pairs=[('TestNamespaceCallGraphCycleUnion','TestEnumNamespaceSharedCycle'),('TestNamespaceEnumInitializationIndependentOfModuleAnalysis','TestEnumInitializationReach')]
(p/'coverage-diffs.json').write_text(json.dumps({a:sorted(covered(a)-covered(b)) for a,b in pairs},indent=2)+'\n')
