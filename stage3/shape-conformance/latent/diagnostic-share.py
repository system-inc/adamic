"""Count diagnostic provenance per site; codes overlap, scope totals partition sites."""
import collections,gzip,json,pathlib,re,sys
path=pathlib.Path(sys.argv[1]);r=json.load(gzip.open(path,'rt') if path.suffix=='.gz' else path.open());label=r['measurement']
scopes=collections.Counter();codes=collections.defaultdict(collections.Counter);examples=collections.defaultdict(set);missing=[]
for row in r['sites']:
 if row['reason']!='diagnosed body':continue
 causes=row.get('diagnostic_causes',[])
 scope='own function' if any(c['scope']=='own function' for c in causes) else 'dependency'
 scopes[(scope,row['kind'])]+=1
 found=set()
 for cause in causes:
  for d in cause['diagnostics']:
   code=re.search(r'error TS(\d+):',d)
   if code:found.add(code[1]);examples[code[1]].add(d.split('error TS'+code[1]+':',1)[1].split('\n')[0].strip())
 if not found:missing.append({k:row[k] for k in ('file','line','column','detail')})
 for code in found:codes[scope][(code,row['kind'])]+=1
out={'measurement':label,'diagnosed_sites':sum(scopes.values()),'scope_totals':{scope:{kind:scopes[(scope,kind)] for kind in ('tagged','untagged')} for scope in ('own function','dependency')},'diagnostic_codes_overlap':[{ 'code':'TS'+code,'message_examples':sorted(examples[code]),**{scope:{kind:codes[scope][(code,kind)] for kind in ('tagged','untagged')} for scope in codes}} for code in sorted(examples,key=int)],'missing_provenance':missing,'note':'Each code is counted once per site/scope. A site may depend on several diagnosed bodies or codes, so diagnostic-code rows overlap.'}
pathlib.Path(sys.argv[2]).write_text(json.dumps(out,indent=2)+'\n');print(json.dumps(out,indent=2))
