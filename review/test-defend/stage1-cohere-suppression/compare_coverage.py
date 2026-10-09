import pathlib,json
p=pathlib.Path(__file__).resolve().parent
def covered(name):
 result=set()
 for row in (p/(name+'.cover')).read_text().splitlines()[1:]:
  loc,stmts,count=row.split()
  if not int(count):continue
  file,span=loc.split(':');start,end=span.split(',')
  for line in range(int(start.split('.')[0]),int(end.split('.')[0])+1):result.add(file+':'+str(line))
 return result
(p/'coverage-diffs.json').write_text(json.dumps(sorted(covered('TestEachGapStandsWhereGapsMdSaysItDoes')-covered('TestThePortParsesAsGoCohereDoes')),indent=2)+'\n')
