from pathlib import Path
import re,json,sys
logs=[Path(p).read_text() for p in sys.argv[1:3]]
kinds=[{int(a):int(b) for a,b in re.findall(r'^kind (\d+) (\d+)$',s,re.M)} for s in logs]
categories=[list(map(int,re.search(r'categories (\d+) (\d+) (\d+) (\d+)',s).groups())) for s in logs]
delta={k:kinds[0].get(k,0)-kinds[1].get(k,0) for k in set(kinds[0])|set(kinds[1])}
cats={n:a-b for n,a,b in zip(['slices','concatenation','number_formatting','growing_append'],*categories)}
cats.update({n:delta.get(k,0) for k,n in enumerate(['objects','arrays','maps','cells','closures','map_iterators'],2)})
cats['other_strings']=delta[1]-sum(cats[n] for n in ['slices','concatenation','number_formatting','growing_append'])
assert sum(cats.values())==sum(delta.values())
print(json.dumps(dict(requests=100000,total=sum(delta.values()),by_category=cats,run_kinds=kinds[0],control_kinds=kinds[1]),indent=2))
