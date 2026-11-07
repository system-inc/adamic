"""Compare normal source, all emitted JS, public declarations, and full censuses."""
from pathlib import Path
import sys,json,hashlib,difflib,collections
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/debug-messages');raw=s/'raw';pins=json.loads((repo/'stage3/adapt/99-correctness-core-debug-messages/pins.json').read_text())['pins']
hash=lambda b:hashlib.sha256(b).hexdigest()
mode=sys.argv[1]
if mode=='builds':
 before=s/'before';after=s/'after';manifests={side:{str(p.relative_to(s/side)):hash(p.read_bytes()) for p in sorted((s/side/'src').rglob('*')) if p.is_file()} for side in ['before','after']}
 differences=[p for p in manifests['before'].keys()|manifests['after'].keys() if manifests['before'].get(p)!=manifests['after'].get(p)];assert differences==['src/compiler/debug.ts'],differences
 old=(before/'src/compiler/debug.ts').read_bytes().decode();new=(after/'src/compiler/debug.ts').read_bytes().decode();expected=old
 for pin in pins:assert expected.count(pin['before'])==1;expected=expected.replace(pin['before'],pin['after'])
 assert expected==new;assert old.count('getFunctionName(test!)')-new.count('getFunctionName(test!)')==4
 (raw/'source-hunks.patch').write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='before/debug.ts',tofile='after/debug.ts')))
 # Stock emission drops non-null assertions. These are the only permitted JS substitutions.
 changes={p['oldMessage'].replace('node!.kind','node.kind').replace('test!','test'):p['newMessage'].replace('node!.kind','node.kind') for p in pins}
 jsfiles=sorted(p.relative_to(before/'built/local') for p in (before/'built/local').rglob('*.js'))
 assert jsfiles==sorted(p.relative_to(after/'built/local') for p in (after/'built/local').rglob('*.js'))
 rows=[];hunks=[]
 for rel in jsfiles:
  a=(before/'built/local'/rel).read_bytes();b=(after/'built/local'/rel).read_bytes();record={'file':str(rel),'before':hash(a),'after':hash(b),'changed':a!=b}
  if a!=b:
   text=a.decode();namePhrase=" '${getFunctionName(test)}'";assert text.count(namePhrase)==4,(rel,text.count(namePhrase))
   text=text.replace(namePhrase,'')
   assert text.encode()==b,rel
   hunks.extend(difflib.unified_diff(a.decode().splitlines(True),b.decode().splitlines(True),fromfile='before/'+str(rel),tofile='after/'+str(rel)));record['messageExpressionsChanged']=4
  rows.append(record)
 (raw/'all-emitted-javascript-hunks.patch').write_text(''.join(hunks));public=[]
 for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
  a=(before/'built/local'/name).read_bytes();b=(after/'built/local'/name).read_bytes();assert a==b;public.append({'file':name,'sha256':hash(a),'bytes':len(a),'byteIdentical':True})
 (raw/'build-comparison.json').write_text(json.dumps({'sourceDifferences':differences,'sourceManifests':manifests,'emittedJavaScript':rows,'publicDeclarations':public,'onlyFourMessageExpressionsChanged':True},indent=2)+'\n');print('source and every JS file checked; public declarations identical; changed JS',len([r for r in rows if r['changed']]))
elif mode=='census':
 records={}
 for side in ['before','after','mutant']:
  stream=[json.loads(l) for l in (raw/(side+'-census.jsonl')).read_text().splitlines()];assert len(stream)==2;header,record=stream;assert not header['diagnostics'] and header['latent_mode']=='full';assert len(record['units'])==105;records[side]=record
 def key(f):return (f['where'].replace(str(s/'before'),'<tree>').replace(str(s/'after'),'<tree>').replace(str(s/'mutant'),'<tree>'),f['kind'],f['reason'],f['phase'])
 sets={side:{key(f) for f in record['findings'] if f['kind']!='Boundary'} for side,record in records.items()}
 removed=sets['before']-sets['after'];added=sets['after']-sets['before'];reason='a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read';targets=[(267,72),(280,98),(293,108),(307,98)]
 for line,col in targets:assert any(k[0].endswith(f'debug.ts:{line}:{col}') and k[2]==reason and k[3]=='refusal_scan' for k in sets['before']);assert not any(k[0].endswith(f'debug.ts:{line}:{col}') and k[2]==reason for k in sets['after'])
 assert not added,sorted(added)
 restored=[k for k in sets['mutant']-sets['after'] if k[0].endswith('debug.ts:267:72') and k[2]==reason and k[3]=='refusal_scan'];assert len(restored)==1
 def unitname(record,where):
  units=[u for u in record['units'] if u['where']==where];return units[0].get('name',units[0]['kind']) if units else '<file refusal scan>'
 text=(s/'after/src/compiler/debug.ts').read_text().splitlines();getFunc=next(i+1 for i,l in enumerate(text) if 'export function getFunctionName(' in l);end=next(i+1 for i in range(getFunc,len(text)) if text[i]=='    }')
 reaches=[]
 for f in records['after']['findings']:
  if '/debug.ts:' not in f['where']:continue
  line=int(f['where'].split(':')[-2])
  if getFunc<=line<=end:reaches.append({'recordingUnit':unitname(records['after'],f['unit']),'where':f['where'].replace(str(s/'after')+'/', ''),'kind':f['kind'],'reason':f['reason'],'phase':f['phase']})
 (raw/'census-comparison.json').write_text(json.dumps({'unitsPerRun':105,'removed':[list(k) for k in sorted(removed)],'added':[list(k) for k in sorted(added)],'restoringMutantSite':[list(k) for k in restored],'getFunctionNameStillPresent':True,'getFunctionNameObservations':reaches,'beforeDistinct':len(sets['before']),'afterDistinct':len(sets['after'])},indent=2)+'\n');print('census four targets gone, no added sites; getFunctionName recording units',sorted({r['recordingUnit'] for r in reaches}))
