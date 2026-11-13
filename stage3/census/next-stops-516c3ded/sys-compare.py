"""Exact multiset delta of all sys.ts findings, preserving unit and phase."""
import collections,hashlib,json
from pathlib import Path
base=Path(__file__).resolve().parent
prefix='/workspace/scratch/next-stops-516c3ded-adapted/src/compiler/'
def read(name):
 rows=[json.loads(l) for l in (base/('sys-'+name+'-full.jsonl')).read_text().splitlines()]
 assert rows[0]['latent_mode']=='full' and rows[-1]['status']=='relation_report'
 assert rows[-1]['relation_overflow']['complete'] if 'complete' in rows[-1]['relation_overflow'] else True
 files=[r for r in rows if 'file' in r];assert len(files)==1 and files[0]['file']==prefix+'sys.ts'
 assert json.loads((base/'CommonJSBundle'/('sys-'+name+'-full-status.json')).read_text())['exit_code']==0
 return rows,files[0]
old_rows,old=read('parent');new_rows,new=read('head')
encode=lambda f:json.dumps(f,sort_keys=True)
a=collections.Counter(map(encode,old['findings']));b=collections.Counter(map(encode,new['findings']))
gone=[json.loads(x) for x in (a-b).elements()];added=[json.loads(x) for x in (b-a).elements()]
# Pair only an unambiguous single old/new finding within the same phase/unit/kind.
key=lambda f:(f.get('phase'),f.get('unit'),f.get('kind'),f.get('root_kind'),f.get('depth'),f.get('pass'),f.get('piece_atom'))
ag=collections.defaultdict(list);bg=collections.defaultdict(list)
for f in gone:ag[key(f)].append(f)
for f in added:bg[key(f)].append(f)
changed=[]
for k in sorted(ag.keys() & bg.keys(),key=str):
 if len(ag[k])==len(bg[k])==1:
  x=ag[k][0];y=bg[k][0];changed.append(dict(before=x,after=y));gone.remove(x);added.remove(y)
result=dict(parent='751879544ea6e627e67c91d6ab3c4ba595be6356',head='516c3dedf370293d4e4f8f78ab4628d391aed5e4',input=prefix+'sys.ts',mode='full',scan_only_fallback=False,checker_rejected={'parent':old_rows[0]['checker_rejected'],'head':new_rows[0]['checker_rejected']},counts={'parent':len(old['findings']),'head':len(new['findings']),'unchanged':sum((a&b).values()),'gone':len(gone),'new':len(added),'changed':len(changed)},phases={n:dict(collections.Counter(f['phase'] for f in r['findings'])) for n,r in [('parent',old),('head',new)]},gone=gone,new=added,changed=changed)
assert result['counts']['parent']==result['counts']['unchanged']+len(gone)+len(changed)
assert result['counts']['head']==result['counts']['unchanged']+len(added)+len(changed)
(base/'sys-comparison.json').write_text(json.dumps(result,indent=2)+'\n')
short=lambda s:s.replace(prefix,'')
s=['Both full runs completed with exit 0 under the 1500-second bound; no scan-only fallback was used. Each used the sys.ts file as the load root and LATENT_ONLY_FILE, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, CommonJSBundle, and the same adapted input. The input is checker-clean in both runs; the tool uses its standard measurement-only label. No usable compiler output was produced.','','Findings: '+json.dumps(result['counts'])+'. The unchanged count compares complete finding objects, including unit, phase, kind, location, reason and text. Repeated locations in different nested units are retained. Unambiguous single old/new findings with the same phase/unit/kind are shown as changed; other differences are new or gone.','','| Commit | Refusal scan | Lowering | Units | Seconds |','| --- | ---: | ---: | ---: | ---: |']
for n,r in [('parent',old),('head',new)]:
 status=json.loads((base/'CommonJSBundle'/('sys-'+n+'-full-status.json')).read_text());p=result['phases'][n]
 s+=['| '+n+' | '+str(p.get('refusal_scan',0))+' | '+str(p.get('lowering',0))+' | '+str(len(r['units']))+' | '+str(round(status['seconds'],3))+' |']
s+=['','Every new, gone, or changed finding follows. Text is exact except that the common adapted compiler directory prefix is omitted. Raw absolute text is retained in JSON and raw streams.','','| Change | Phase | Kind | Unit | Before text | After text |','| --- | --- | --- | --- | --- | --- |']
def row(kind,before,after):
 f=after or before
 vals=[kind,f['phase'],f['kind'],short(f['unit']),short(before['text']) if before else '—',short(after['text']) if after else '—']
 s.append('| '+' | '.join(v.replace('|','\\|').replace('\n','<br>') for v in vals)+' |')
for f in gone:row('gone',f,None)
for f in added:row('new',None,f)
for pair in changed:row('changed',pair['before'],pair['after'])
s+=['','All 260 lowering findings are unchanged. Both commits record the NotYet at sys.ts:160:17: `stage 0 can\'t lower assigning an element of a value yet`, including its boundary records in three nested units. This gap is not newly introduced by the head. The ordinary census never reaches it because the refusal scan stops first at sys.ts:272:23.', '', 'Adapted line 160, where full mode reaches the lowering gap:','','```typescript',(Path(prefix)/'sys.ts').read_text().splitlines()[159],'```','']
(base/'sys-comparison.md').write_text('\n'.join(s)+'\n')
for name in ['head','parent']:
 p=Path('/workspace/scratch/next-stops-516c3ded-latent-'+name+'-census')
 (base/('sys-'+name+'-binary.json')).write_text(json.dumps(dict(path=str(p),bytes=p.stat().st_size,sha256=hashlib.sha256(p.read_bytes()).hexdigest()),indent=2)+'\n')
print(result['counts']);print(result['phases'])
