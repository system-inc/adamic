from pathlib import Path
import json,re,difflib,collections
out=Path(__file__).resolve().parent
roots={'base':'/workspace/scratch/single-watcher-before','adapted':'/workspace/scratch/single-watcher-after','mutant':'/workspace/scratch/single-watcher-mutant'}
source_roots=roots
fields=('kind','where','reason','text')
def rows(n):return [json.loads(x) for x in (out/n/'latent.jsonl').read_text().splitlines()]
def norm(n,x):
 if isinstance(x,str):return x.replace(roots[n]+'/','')
 return x

def findings(n):
 r=rows(n);assert r[-1]['status']=='relation_report' and r[0]['latent_mode']=='full' and r[0]['module_format']=='CommonJSBundle'
 assert len(r)==3 and r[1]['file']==roots[n]+'/src/compiler/sys.ts'
 return r,{tuple(norm(n,f[k]) for k in fields) for f in r[1]['findings']}
def dicts(x):return [dict(zip(fields,t)) for t in sorted(x)]
base,bs=findings('base');summary={}
for n in ['adapted','mutant']:
 rs,ns=findings(n);status=json.loads((out/n/'latent-status.json').read_text());assert status['exit_code']==0
 a=Path(roots['base']+'/src/compiler/sys.ts').read_text().splitlines();b=Path(source_roots[n]+'/src/compiler/sys.ts').read_text().splitlines();mapping={}
 for tag,i,j,k,l in difflib.SequenceMatcher(None,a,b,autojunk=False).get_opcodes():
  if tag=='equal':mapping.update({x+1:k+x-i+1 for x in range(i,j)})
 def mapped(value):
  return re.sub(r'src/compiler/sys\.ts:(\d+):(\d+)',lambda m:f'src/compiler/sys.ts:{mapping.get(int(m[1]),"unmapped-base-"+m[1])}:{m[2]}',value)
 mappedbs={tuple(mapped(v) for v in x) for x in bs}
 added=ns-bs;gone=bs-ns;alignedadded=ns-mappedbs;alignedgone=mappedbs-ns
 # Changed tuples are co-located removed/added observations; ambiguous multiple reasons stay explicit in the added/gone ledgers.
 pairs=[]
 for old in sorted(alignedgone):
  matches=[new for new in alignedadded if new[1]==old[1]]
  if len(matches)==1:pairs.append({'before':dict(zip(fields,old)),'after':dict(zip(fields,matches[0]))})
 data={'dedup_key':fields,'exact_added':dicts(added),'exact_gone':dicts(gone),'source_aligned_added':dicts(alignedadded),'source_aligned_gone':dicts(alignedgone),'source_aligned_colocated_changes':pairs,'line_mapping':mapping}
 (out/n/'findings-delta.json').write_text(json.dumps(data,indent=2)+'\n')
 text=['# '+n.upper()+': complete finding deltas','', 'Deduplicated by (kind, where, reason, text). All paths are relative to the adapted root. Raw JSONL retains every attempt and phase.','']
 for label,finds in [('Exact gone',gone),('Exact new',added),('Source-aligned gone',alignedgone),('Source-aligned new',alignedadded)]:
  text+=['## '+label,'','| Kind | Location | Reason |','|---|---|---|']
  for kind,where,reason,detail in sorted(finds):text.append('| '+kind+' | '+where+' | '+reason.replace('|','\\|').replace('\n',' ')+' |')
  text+=['']
 (out/n/'findings-delta.md').write_text('\n'.join(text).rstrip()+'\n')
 unique_real=[x for x in ns if x[0] in ('Refused','NotYet')]
 summary[n]={'units':len(rs[1]['units']),'raw_findings':len(rs[1]['findings']),'unique_all':len(ns),'unique_refused_notyet':len(unique_real),'kind_counts_unique':dict(collections.Counter(x[0] for x in ns)),'exact_added':len(added),'exact_gone':len(gone),'source_aligned_added':len(alignedadded),'source_aligned_gone':len(alignedgone),'source_aligned_changes':len(pairs),'helper_cast_survives':any('a cast the runtime can\'t check'==x[2] and bool(re.search(r'src/compiler/sys\.ts:51[45]:17$', x[1])) for x in ns),'checker_diagnostics':len(rs[0]['diagnostics']),'seconds':status['seconds'],'unit_statuses':dict(collections.Counter(x['status'] for x in rs[1]['units']))}
summary['base']={'units':len(base[1]['units']),'raw_findings':len(base[1]['findings']),'unique_all':len(bs),'kind_counts_unique':dict(collections.Counter(x[0] for x in bs)),'unique_refused_notyet':sum(x[0] in ('Refused','NotYet') for x in bs),'unit_statuses':dict(collections.Counter(x['status'] for x in base[1]['units']))}
(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))
