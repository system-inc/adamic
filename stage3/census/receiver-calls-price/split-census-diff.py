"""Compare complete selected-file records, preserving every changed finding."""
import json,pathlib,re,difflib,collections,hashlib
scratch=pathlib.Path('/workspace/scratch');base=scratch/'receiver-baseline';split=scratch/'receiver-split';out=scratch/'receiver-split-results';maps={}
for file in (base/'src/compiler').rglob('*.ts'):
 rel=str(file.relative_to(base));a=file.read_text().splitlines();b=(split/rel).read_text().splitlines();mapping={}
 for block in difflib.SequenceMatcher(a=a,b=b,autojunk=False).get_matching_blocks():
  for i in range(block.size):mapping[block.b+i+1]=block.a+i+1
 maps[rel]=mapping
location=re.compile(r'(src/compiler/[A-Za-z0-9_./-]+\.ts):(\d+):(\d+)')
def normalize(value,road):
 if isinstance(value,str):
  value=value.replace(str(base)+'/', '').replace(str(split)+'/', '')
  if road=='split':value=location.sub(lambda m:m[1]+':'+str(maps.get(m[1],{}).get(int(m[2]),int(m[2])))+':'+m[3],value)
  return value
 return value
def key(finding,road):return tuple(normalize(finding.get(k,''),road) for k in ['unit','phase','kind','where','reason','text'])
results=[];manifest=[]
for file in [r['file'] for r in json.loads((scratch/'receiver-split-rewrite.json').read_text())['files']]:
 stem=file.removeprefix('src/compiler/').removesuffix('.ts').replace('/','__');roads={}
 for road in ['baseline','split']:
  path=out/'census'/(road+'__'+stem+'.jsonl');rows=[json.loads(line) for line in path.read_text().splitlines()];assert len(rows)==3,(file,road,len(rows));assert rows[0]['module_format']=='CommonJSBundle' and rows[0]['latent_mode']=='full' and not rows[0]['diagnostics'];assert rows[-1]['status']=='relation_report';assert rows[1]['file'].endswith(file)
  meta=json.loads(path.with_suffix('.json').read_text());assert meta['exit']==0 and not meta['timeout'];roads[road]={'findings':rows[1]['findings'],'units':rows[1]['units'],'measurement':meta}
  for p in [path,path.with_suffix('.log')]:manifest.append({'path':str(p),'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()})
 old=collections.defaultdict(list);new=collections.defaultdict(list)
 for f in roads['baseline']['findings']:old[key(f,'baseline')].append(f)
 for f in roads['split']['findings']:new[key(f,'split')].append(f)
 added=[];removed=[]
 for k in old.keys()|new.keys():
  common=min(len(old[k]),len(new[k]));removed+=old[k][common:];added+=new[k][common:]
 results.append({'file':file,'baselineFindings':len(roads['baseline']['findings']),'prototypeFindings':len(roads['split']['findings']),'added':added,'removed':removed,'baselineMeasurement':roads['baseline']['measurement'],'prototypeMeasurement':roads['split']['measurement'],'baselineMethodDestructures':[f for f in roads['baseline']['findings'] if f['reason']=='a method in object destructuring'],'prototypeMethodDestructures':[f for f in roads['split']['findings'] if f['reason']=='a method in object destructuring']})
(out/'census-diff-full.json').write_text(json.dumps({'comparison':'multiset, unchanged source lines remapped through difflib matching blocks; full added/removed raw findings retained','files':results},indent=2)+'\n');(out/'raw-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(json.dumps([{'file':f['file'],'new':len(f['added']),'removed':len(f['removed'])} for f in results],indent=2))
