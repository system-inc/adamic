"""Recount the second census and compare it with the immutable previous cumulative run."""
import collections, gzip, hashlib, json, pathlib, shutil, sys
ROOT=pathlib.Path(__file__).resolve().parent
LABEL='measured on a checker-rejected program'
KINDS=['NotYet','Refused','SkippedDependency','error','panic']
VARIANCE_MARKERS={'adamic/invariant-mutable':'mutable-invariance','method-signature-style':'method-parameter-bivariance','adamic/contravariant-override':'contravariant-override'}
VARIANCE_REASONS={'a readonly inherited field redeclared mutable':'readonly-to-mutable-override','an override that widens its return type':'covariant-return-override'}
def site(f):return (f['kind'],f['where'],f['reason'],f['text'])
def counts(sites):
 c=collections.Counter(s[0] for s in sites);return {k:c[k] for k in KINDS}
def reasons(sites):return dict(sorted(collections.Counter(s[0]+': '+s[2] for s in sites).items()))
def variance_family(f):
 if f['kind']!='Refused':return None
 for marker,family in VARIANCE_MARKERS.items():
  if marker in f['text']:return family
 return VARIANCE_REASONS.get(f['reason'])
def normalize(v,prefix):
 if isinstance(v,str):return v.replace(prefix+'/', '')
 if isinstance(v,list):return [normalize(x,prefix) for x in v]
 if isinstance(v,dict):return {k:normalize(x,prefix) for k,x in v.items()}
 return v

def byte_offset(text,line,column):
 # Where columns are UTF-16; Go AST positions are UTF-8 byte offsets. Preserve CRLF.
 lines=text.splitlines(keepends=True);prefix=''.join(lines[:line-1]);units=0;chars=[]
 for ch in lines[line-1]:
  if units==column-1:break
  units+=len(ch.encode('utf-16-le'))//2;chars.append(ch)
 assert units==column-1,(line,column)
 return len((prefix+''.join(chars)).encode('utf-8'))

def owner_at(where,records,texts):
 file,line,col=where.rsplit(':',2);position=byte_offset(texts[file],int(line),int(col))
 declarations=records[file]['declarations']
 candidates=[d for d in declarations if d['start']<=position<d['end']]
 if not candidates:return dict(measurement=LABEL,where=file+':1:1',name='<file scope>',kind='KindSourceFile',start=0,end=len(texts[file].encode()),qualified_name='<file scope>')
 chosen=min(candidates,key=lambda d:(d['end']-d['start'],-d['start']))
 parents=sorted((d for d in candidates if d['start']<=chosen['start'] and chosen['end']<=d['end']),key=lambda d:-(d['end']-d['start']))
 result=dict(chosen);result['name']=result['name'].strip();result['qualified_name']=' / '.join(d['name'].strip() for d in parents)
 return result

def variance_breakdown(findings,raw,texts):
 records={r['file']:r for r in raw[1:]};groups={};sites=[]
 for f in findings:
  family=variance_family(f)
  if not family:continue
  owner=owner_at(f['where'],records,texts)
  k=owner['where'];entry=groups.setdefault(k,dict(measurement=LABEL,declaration=owner,count=0,families=collections.Counter(),findings=[]))
  entry['count']+=1;entry['families'][family]+=1
  record=dict(f,variance_family=family,owning_declaration=owner)
  record['attempting_units']=sorted({event['unit'] for r in raw[1:] for event in r['findings'] if site(event)==site(f)})
  entry['findings'].append(record);sites.append(record)
 return dict(measurement=LABEL,total=len(sites),owning_declaration_count=len(groups),families=dict(sorted(collections.Counter(s['variance_family'] for s in sites).items())),per_declaration=dict(sorted(groups.items())))

def summarize(raw,manifest):
 header=raw[0];assert header['checker_rejected'] and header['status']=='measurement'
 assert {r['file'] for r in raw[1:]}=={f['file'] for f in manifest} and len(raw[1:])==78
 keys={site(f) for r in raw[1:] for f in r['findings']}
 units=[u for r in raw[1:] for u in r['units']]
 per_file={}
 for row in raw[1:]:
  file=row['file'];own={s for s in keys if s[1].rsplit(':',2)[0]==file}
  per_file[file]=dict(measurement=LABEL,checker=sum(d['file']==file for d in header['diagnostic_sites']),**counts(own),per_reason=reasons(own),functions_attempted=sum(u['kind']=='KindFunctionDeclaration' and u['status']!='skipped_checker_body' for u in row['units']),bodies_skipped=sum(u['status']=='skipped_checker_body' for u in row['units']))
 clean=sorted(f for f,v in per_file.items() if v['checker']==0)
 return dict(measurement=LABEL,checker_rejected=True,checker_total=len(header['diagnostics']),checker_per_reason=dict(sorted(collections.Counter(d.split('error ')[1].split(':')[0] for d in header['diagnostics']).items())),checker_clean_files=clean,checker_clean_ratio=dict(measurement=LABEL,numerator=len(clean),denominator=78,fraction=f'{len(clean)}/78',percent=100*len(clean)/78),lowering_counts=counts(keys),per_reason=reasons(keys),per_file=per_file,top_level_units=len(units),functions_attempted=sum(u['kind']=='KindFunctionDeclaration' and u['status']!='skipped_checker_body' for u in units),bodies_skipped=sum(u['status']=='skipped_checker_body' for u in units),recorded_events=sum(len(r['findings']) for r in raw[1:]),findings=[dict(measurement=LABEL,kind=k[0],where=k[1],reason=k[2],text=k[3]) for k in sorted(keys)])

def delta(current,previous):
 files=current['per_file'];pr=previous['per_reason'];cr=current['per_reason']
 return dict(measurement=LABEL,checker=current['checker_total']-previous['checker_total'],checker_clean_files=current['checker_clean_ratio']['numerator']-previous['checker_clean_ratio']['numerator'],counts={k:current['lowering_counts'][k]-previous['lowering_counts'][k] for k in KINDS},per_reason={k:cr.get(k,0)-pr.get(k,0) for k in sorted(cr.keys()|pr.keys())},per_file={f:dict(measurement=LABEL,checker=files[f]['checker']-previous['per_file'][f]['checker'],**{k:files[f][k]-previous['per_file'][f][k] for k in KINDS}) for f in files},new_checker_clean_files=sorted(set(current['checker_clean_files'])-set(previous['checker_clean_files'])),no_longer_checker_clean_files=sorted(set(previous['checker_clean_files'])-set(current['checker_clean_files'])))

def main():
 scratch=pathlib.Path(sys.argv[1]).resolve();adapted=pathlib.Path(sys.argv[2]).resolve();data=ROOT/'data/rerun2';data.mkdir(exist_ok=True)
 with gzip.open(ROOT/'data/previous-REPORT.json.gz','rt') as f:previous=json.load(f)
 prior_run=next(r for r in previous['runs'] if r['name']=='cumulative')
 with gzip.open(ROOT/'data/cumulative.jsonl.gz','rt') as f:prior_raw=[json.loads(l) for l in f]
 prior_manifest=previous['source']['files'];baseline=dict(name='cumulative',head=prior_run['head'],delivery_commit='70456b766243ee155104c4c9c6928cc53b7b5640',**summarize(prior_raw,prior_manifest))
 baseline['variance']=dict(measurement=LABEL,total=sum(variance_family(f) is not None for f in baseline['findings']),families=dict(sorted(collections.Counter(variance_family(f) for f in baseline['findings'] if variance_family(f)).items())))
 manifest=[dict(file=str(p.relative_to(adapted)),sha256=hashlib.sha256(p.read_bytes()).hexdigest(),generated='.generated.' in p.name) for p in sorted((adapted/'src/compiler').rglob('*.ts'))]
 assert {f['file'] for f in manifest}=={f['file'] for f in prior_manifest}
 texts={f['file']:(adapted/f['file']).read_bytes().decode('utf-8') for f in manifest}
 runs=[]
 for spec in json.loads((scratch/'runs.json').read_text()):
  assert spec['status']=='complete',spec
  name=spec['name'];raw=[normalize(json.loads(l),str(adapted)) for l in (scratch/(name+'.jsonl')).read_text().splitlines()]
  with gzip.open(data/(name+'.jsonl.gz'),'wt') as f:
   for r in raw:f.write(json.dumps(r)+'\n')
  row=dict(spec,**summarize(raw,manifest));row['variance']=variance_breakdown(row['findings'],raw,texts)
  row['delta_vs_previous_cumulative']=delta(row,baseline)
  row['delta_vs_previous_cumulative']['variance']=row['variance']['total']-baseline['variance']['total']
  row['overlay_sha256']={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((scratch/(name+'-overlay')).glob('*.go'))}
  row['binary_sha256']=hashlib.sha256((scratch/(name+'-census')).read_bytes()).hexdigest()
  for phase in ['overlay','build','run']:shutil.copyfile(scratch/(name+'-'+phase+'.log'),data/(name+'-'+phase+'.log'))
  runs.append(row)
 assert [r['name'] for r in runs]==['cumulative-2','cumulative-2-nested']
 runs[1]['delta_vs_cumulative_2']=delta(runs[1],runs[0])
 result=dict(complete=True,measurement=LABEL,count_definition=previous['count_definition'],method='same eligibility, independent unit lowering and unique finding deduplication; added declaration-span metadata only',comparison_note='Delta against previous cumulative combines changed source adaptations and compiler commits. Exact reason strings and locations can change with types and source edits.',source=dict(typescript='6.0.3',commit=previous['source']['commit'],files=manifest,adaptations=['10-type-imports','20-optional-declarations','30-indexed-reads','45-regex-captures','46-fix-pragma-empty-argument']),baseline=baseline,runs=runs,variance_definition='Refused sites tagged adamic/invariant-mutable, method-signature-style, adamic/contravariant-override; plus readonly-to-mutable or widened-return overrides. Owning declaration is the nearest named AST declaration at the diagnostic location, not the attempting caller. Enum/nominal/cast refusals excluded.')
 (ROOT/'REPORT.json').write_text(json.dumps(result,indent=2)+'\n')
 print(json.dumps([dict(name=r['name'],checker=r['checker_total'],ratio=r['checker_clean_ratio'],counts=r['lowering_counts'],variance=r['variance']['total'],delta=r['delta_vs_previous_cumulative']['counts']) for r in runs],indent=2))
if __name__=='__main__':main()
