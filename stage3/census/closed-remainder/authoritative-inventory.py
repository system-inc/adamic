#!/usr/bin/env python3
"""Rebuild the user-specified shard inventory with pair's exact distinct key."""
import collections,gzip,hashlib,importlib.util,json,os,pathlib,signal,subprocess,sys
sys.dont_write_bytecode=True
source,tree,typescript,dest=map(pathlib.Path,sys.argv[1:]);tree=tree.resolve()
spec=importlib.util.spec_from_file_location('archived_merge',source/'merge.py');merge=importlib.util.module_from_spec(spec);spec.loader.exec_module(merge)
shards=[source/f'shard-{i}.jsonl.gz' for i in range(8)];rows=merge.merge([[json.loads(s) for s in gzip.open(p,'rt') if s.strip()] for p in shards])
reason='checked predicate membership over an open tag domain';raw=[f for r in rows for f in r.get('findings',[]) if f.get('reason')==reason]
def norm(w):return 'src/'+w.split('/src/',1)[1] if '/src/' in w else w
def normalized(f):return {k:norm(v) if isinstance(v,str) else v for k,v in f.items()}
files=collections.defaultdict(list);seen=set()
for f in raw:
 for field in ['where','unit']:
  w=f[field]
  if w in seen:continue
  seen.add(w);file,line,col=w.rsplit(':',2);files[norm(file)].append([w,int(line),int(col)])
argv=['node',str(source/'normalize.cjs')];child=subprocess.run(argv,input=json.dumps({'tree':str(tree),'typescript':str(typescript.resolve()),'files':files}),text=True,capture_output=True)
print(json.dumps({'argv':argv,'child_exit_code':child.returncode}),flush=True)
if child.returncode:
 sys.stderr.write(child.stderr)
 if child.returncode<0:
  if -child.returncode not in (signal.SIGKILL,signal.SIGSTOP):signal.signal(-child.returncode,signal.SIG_DFL)
  os.kill(os.getpid(),-child.returncode)
 sys.exit(child.returncode)
ids=json.loads(child.stdout);hashes=json.loads((source/'adapted-source-hashes.json').read_text())
assert all(ids[w]['sha256']==hashes[ids[w]['file']] for w in ids),'archived/current source bytes differ'
units={};first={}
for r in rows[1:]:
 for u in r.get('units',[]):units[u['where']]=u
 for f in r.get('findings',[]):
  if f['phase'] in ('lowering','deferred') and f['unit'] in units:first.setdefault(f['unit'],f)
# This is pair.py:view(..., distinct=True), including first representative wins.
distinct={};occurrences=collections.defaultdict(list)
for f in raw:
 i=ids[f['where']];key=((i['file'],i['sha256'],i['read_offset']),f['reason']);distinct.setdefault(key,f);occurrences[key].append(f)
sites=[]
for key,f in distinct.items():
 i=ids[f['unit']]
 sites.append({'key':[list(key[0]),key[1]],'finding':normalized(f),'unit_identity':[i['file'],i['sha256'],i['byte_offset']],'historical_first_stop':normalized(first[f['unit']]),'is_historical_first_stop':first[f['unit']]==f,'raw_occurrences':[normalized(g) for g in occurrences[key]]})
result={'source_branch':'typescript/overload-proof-census','source_commit':'481552841c9b5e1496fc6237da10501b707178d0','source_files':[{'path':str(p.name),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in shards],'key_definition':'((source path, source SHA-256, outermost contiguous access-chain UTF-8 byte offset), exact reason); first representative in archived merge order wins','first_stop_definition':'pair: first lowering/deferred finding whose unit exists; ordinary scan findings excluded','raw_occurrences':len(raw),'raw_owner_units':len({f['unit'] for f in raw}),'distinct_findings':len(sites),'retained_units':len({tuple(s['unit_identity']) for s in sites}),'retained_first_stops':sum(s['is_historical_first_stop'] for s in sites),'dedup_difference_from_pair':'none','relevant_source_hashes_match':True,'sites':sites}
# Cross-check counts by executing the repository's unmodified pair comparator.
pair_path=pathlib.Path(__file__).resolve().parent.parent/'pair/pair.py'
spec=importlib.util.spec_from_file_location('reference_pair',pair_path);pair=importlib.util.module_from_spec(spec);spec.loader.exec_module(pair)
owner_locations={f['unit'] for f in raw};filtered=[rows[0]]
for r in rows[1:]:
 filtered.append(dict(findings=[f for f in r.get('findings',[]) if f.get('reason')==reason],units=[u for u in r.get('units',[]) if u['where'] in owner_locations]))
reference=pair.compare(filtered,filtered,tree,typescript)
result['reference_pair_raw_count']=reference['raw_findings']['per_reason'][reason]['before']
result['reference_pair_distinct_count']=reference['distinct_reads']['per_reason'][reason]['before']
expected_raw={(ids[f['where']]['file'],ids[f['where']]['sha256'],ids[f['where']]['byte_offset'],f['kind'],f['reason'],f['text']) for f in raw}
assert result['reference_pair_raw_count']==len(expected_raw) and result['reference_pair_distinct_count']==len(sites)
result['raw_occurrences']=len(raw)
result['raw_findings']=result['reference_pair_raw_count']
print(json.dumps({'reference_pair_normalizer_child_exit_code':0,'distinct_count':result['reference_pair_distinct_count']}),flush=True)
dest.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:result[k] for k in ['raw_occurrences','raw_findings','raw_owner_units','distinct_findings','retained_units','retained_first_stops','dedup_difference_from_pair']}))
