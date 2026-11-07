import json,pathlib,re,sys
raw,declarations,output=map(pathlib.Path,sys.argv[1:])
records=[json.loads(s) for s in raw.read_text().splitlines()]
record=next(r for r in records if r.get('file','').endswith('/compiler/core.ts'))
units=[u for u in record['units'] if u.get('name')=='or']
locations={u['where'] for u in units}
decls=json.loads(declarations.read_text());lo=min(d['line'] for d in decls);hi=max(d['endLine'] for d in decls)
def belongs(f):
 for key in ['where','site_where']:
  m=re.search(r'/compiler/core\.ts:(\d+):',f.get(key,''))
  if m and lo<=int(m.group(1))<=hi:return True
 return f.get('unit') in locations
findings=[f for f in record['findings'] if belongs(f)]
summary={'checkerDiagnostics':records[0]['diagnostics'],'mode':records[0]['latent_mode'],'coreUnits':len(record['units']),'coreFindingsIncludingBoundaries':len(record['findings']),'orRange':[lo,hi],'orUnits':units,'orFindings':findings,'orOverloadStopsAnywhereInCore':[f for f in record['findings'] if 'overload 1 of or' in f.get('reason','')],'requiredGatePassed':any('cannot satisfy implementation parameter' in f.get('reason','') or 'cannot be served by implementation result' in f.get('reason','') for f in findings)}
output.write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k!='orUnits'},indent=2))
