import collections,json,pathlib,re,sys
before,after,mutant,out=map(pathlib.Path,sys.argv[1:])
def read(directory):
 rows=[json.loads(l) for l in (directory/'raw.jsonl').read_text().splitlines()]
 assert len(rows)==2 and rows[0]['checker_rejected'] is False and rows[0]['diagnostics']==[]
 assert all(not u['checker_diagnostics'] for u in rows[1]['units'])
 return rows[1]
b,a,m=map(read,[before,after,mutant])
pins=json.loads((pathlib.Path(__file__).with_name('owners.json')).read_text());pin=next(p for p in pins if p['owner']=='createQueue')
start=next(int(u['where'].rsplit(':',2)[1]) for u in b['units'] if u.get('name')=='createQueue')
oldEnd=start+pin['previousAfter'].count('\n');newEnd=start+pin['after'].count('\n');delta=newEnd-oldEnd
pattern=re.compile(r'.*/src/compiler/core.ts:(\d+):(\d+)$')
def key(row,shift=False):
 match=pattern.fullmatch(row['where']);assert match,row['where'];line,column=map(int,match.groups())
 if shift and line>newEnd:line-=delta
 return (line,column,row['kind'],row['reason'],row['phase'])
def sites(row,shift=False):return collections.Counter(key(r,shift) for r in row['findings'] if r['kind'] in ['Refused','NotYet'])
bs,as_,ms=sites(b),sites(a,True),sites(m)
removed=bs-as_;added=as_-bs
assert not added,added
assert sum(removed.values())==3,removed
assert all(start<=k[0]<=oldEnd for k in removed),removed
assert ms==bs,ms-bs
assert not any(start<=k[0]<=newEnd for k in sites(a)),sites(a)
result={'checkerDiagnostics':0,'beforeFindingEvents':sum(bs.values()),'afterFindingEvents':sum(as_.values()),'removed':[{**dict(zip(['line','column','kind','reason','phase'],k)),'events':v} for k,v in sorted(removed.items())],'added':[],'otherCoreFindingsIdenticalAfterLineMapping':True,'mutantRestoresAllBeforeSites':True,'sourceLineDelta':delta,'genericCoverage':{label:[u for u in row['units'] if u.get('name')=='createQueue'] for label,row in [('before',b),('after',a),('mutant',m)]}}
out.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
