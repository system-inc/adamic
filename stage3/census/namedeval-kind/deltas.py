import pathlib,json,re,collections
HERE=pathlib.Path(__file__).resolve().parent
ROOT='/workspace/scratch/namedeval-kind-runtime/adapted-B/'
POS=re.compile(r'((?:[A-Za-z]:)?[A-Za-z0-9_./@+\\-]+\.(?:tsx?|jsx?|a|go)):\d+:\d+')
def stripped(s):
 s=s.replace(ROOT,'')
 s=re.sub(r'/workspace/scratch/namedeval-kind-(?:B|S1|S2)/','<compiler>/',s)
 return POS.sub(r'\1:<position>',s)
def identity(f):return (f['kind'],f['phase'],stripped(f.get('reason','')),stripped(f.get('text','')))
def location(f):return (f.get('unit',''),f.get('where',''),f['phase'])
def read(case,file):
 rows=[json.loads(s) for s in (HERE/f'{case}-latent-{file}.jsonl').read_text().splitlines()]
 assert len(rows)==3 and rows[0]['latent_mode']=='full' and rows[0]['module_format']=='CommonJSBundle' and rows[-1]['status']=='relation_report'
 return rows[1]
def compare(before,after):
 a=before['findings'];b=after['findings'];usedA=set();usedB=set();paired=[]
 ga=collections.defaultdict(list);gb=collections.defaultdict(list)
 for i,f in enumerate(a):ga[identity(f)].append(i)
 for i,f in enumerate(b):gb[identity(f)].append(i)
 for key in sorted(ga.keys()&gb.keys()):
  # Exact source/unit pairs first; identity itself never depends on positions.
  for i in ga[key]:
   j=next((j for j in gb[key] if j not in usedB and location(a[i])==location(b[j])),None)
   if j is not None:usedA.add(i);usedB.add(j);paired.append({'before':a[i],'after':b[j]})
  ai=[i for i in ga[key] if i not in usedA];bi=[i for i in gb[key] if i not in usedB]
  for i,j in zip(ai,bi):usedA.add(i);usedB.add(j);paired.append({'before':a[i],'after':b[j]})
 strictGone=[a[i] for i in range(len(a)) if i not in usedA];strictAppeared=[b[j] for j in range(len(b)) if j not in usedB]
 changed=[]
 # In-place replacements are disclosed separately from canonical pairing.
 for i in range(len(a)):
  if i in usedA:continue
  j=next((j for j in range(len(b)) if j not in usedB and location(a[i])==location(b[j])),None)
  if j is not None:usedA.add(i);usedB.add(j);changed.append({'before':a[i],'after':b[j],'grade':'changed diagnostic at the same unit/source boundary'})
 blocked={f['unit'] for f in a if f['phase'] in ['lowering','deferred','registration','prepass']}
 new=[];reached=[]
 for j,f in enumerate(b):
  if j in usedB:continue
  (reached if f['unit'] in blocked else new).append(f)
 gone=[a[i] for i in range(len(a)) if i not in usedA]
 return {'identityFields':['kind','phase','reason with positions stripped','text/message with positions stripped'],'beforeCount':len(a),'afterCount':len(b),'pairedUnchanged':paired,'strictCanonicalGone':strictGone,'strictCanonicalAppeared':strictAppeared,'new':new,'reached':reached,'gone':gone,'changed':changed}
if __name__=='__main__':
 summary={}
 for left,right in [('B','S1'),('S1','S2')]:
  for file in ['namedEvaluation','esDecorators']:
   d=compare(read(left,file),read(right,file));name=f'{left}-to-{right}-{file}'
   (HERE/(name+'-delta.json')).write_text(json.dumps(d,indent=2)+'\n')
   counts={k:len(d[k]) for k in ['new','reached','gone','changed','pairedUnchanged','strictCanonicalGone','strictCanonicalAppeared']}
   counts.update(before=d['beforeCount'],after=d['afterCount']);summary[name]=counts;print(name,json.dumps(counts))
 (HERE/'delta-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
