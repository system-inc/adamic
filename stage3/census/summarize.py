"""Normalize and rank a census. Counts use unique original files and diagnostic start lines."""
import collections
import gzip
import json
import pathlib
import re
import sys

HERE = pathlib.Path(__file__).resolve().parent
DATA = HERE / 'data'
CACHE = pathlib.Path(sys.argv[1]) if len(sys.argv) > 1 else pathlib.Path('/workspace/cache/tsc-census')
PREFIXES = {str(CACHE/'typescript'):'<typescript>', str(CACHE/'prepared'):'<prepared>', str(HERE.parents[1]):'<adamic>'}

def normalized(text):
    for old,new in PREFIXES.items(): text=text.replace(old,new)
    return text

def read_raw(name):
    path = CACHE / (name+'.jsonl')
    result = [json.loads(normalized(line)) for line in path.read_text().splitlines()]
    with (DATA/(name+'.jsonl.gz')).open('wb') as destination:
        with gzip.GzipFile(filename='',mode='wb',fileobj=destination,mtime=0) as stream:
            for record in result: stream.write((json.dumps(record,sort_keys=True)+'\n').encode())
    return result

def write(name, value):
    (DATA/name).write_text(json.dumps(value,indent=2,sort_keys=True)+'\n')

stock=read_raw('stock')
upstream=read_raw('upstream')
repros=read_raw('repros')
policy=read_raw('policy-repros')
non_sources=read_raw('non-sources')
read_raw('cycle-probes')
read_raw('root-probes')
read_raw('stock-library-entry')
files=json.loads((DATA/'files.json').read_text())
original_files={v['file']:v for v in files if not v['generated']}
sites=json.loads((DATA/'sites.json').read_text())
manifest=json.loads((DATA/'repro_manifest.json').read_text())
by_repro={r['reason']:r for r in manifest}
by_observation={r['roots'][0].removeprefix('stage3/census/'):r for r in repros if len(r['roots'])==1}
by_policy={r['roots'][0].removeprefix('stage3/census/'):r for r in policy if len(r['roots'])==1}
pattern=re.compile(r'^(.*?):(\d+):(\d+): error TS(\d+): (.*)',re.S)
diagnostics=[]
for profile, records in [('stock',stock),('upstream-config',upstream)]:
    for text in records[-1].get('diagnostics',[]):
        match=pattern.match(text)
        if not match: diagnostics.append({'profile':profile,'kind':'global','message':text}); continue
        file=match[1].replace('<typescript>/','').replace('<prepared>/','')
        diagnostics.append({'profile':profile,'file':file,'line':int(match[2]),'column':int(match[3]),'code':int(match[4]),'message':match[5],'diagnostic':text})
write('diagnostics.json',diagnostics)
groups=collections.defaultdict(list)
for d in diagnostics:
    if d['profile']=='stock' and d.get('file') in original_files:
        groups['TS'+str(d['code'])].append(d)
for site in sites:
    if site['file'] in original_files and site['reason'] in by_repro:
        groups[site['reason']].append(site)

rankings=[]
for reason,items in groups.items():
    seen_files=sorted({v['file'] for v in items})
    seen_lines=sorted({(v['file'],v['line']) for v in items})
    entry=by_repro[reason]
    observed=by_observation[entry['path']]
    policy_observed=by_policy[entry['path']]
    if reason.startswith('TS'):
        assert any('error '+reason+':' in d for d in observed.get('diagnostics',[])),(reason,observed)
    rankings.append({'reason':reason,'layer':'checker diagnostic' if reason.startswith('TS') else items[0]['evidence'],'occurrences':len(items),'files':seen_files,'file_count':len(seen_files),'start_line_count':len(seen_lines),'affected_file_lines':sum(original_files[f]['lines'] for f in seen_files),'locations':[[f,line] for f,line in seen_lines],'category':entry['category'],'action':entry['action'],'repro':entry['path'],'observed':observed,'policy_observed':policy_observed,'example':{k:items[0].get(k) for k in ['file','line','column']}})
rankings.sort(key=lambda r:(-r['file_count'],-r['start_line_count'],r['reason']))
for rank,item in enumerate(rankings,1): item['rank_by_files']=rank
for rank,item in enumerate(sorted(rankings,key=lambda r:(-r['start_line_count'],-r['file_count'],r['reason'])),1): item['rank_by_lines']=rank
for rank,item in enumerate(sorted(rankings,key=lambda r:(-r['affected_file_lines'],-r['file_count'],r['reason'])),1): item['rank_by_affected_file_lines']=rank
write('rankings.json',rankings)
write('top30.json',rankings[:30])

# Tarjan over actual module-symbol-resolved edges, not import-text guesses.
edges=json.loads((DATA/'module_edges.json').read_text())
adjacency=collections.defaultdict(set)
for e in edges:
    if e['target'] in original_files: adjacency[e['file']].add(e['target'])
index={}; low={}; stack=[]; active=set(); components=[]
def visit(node):
    index[node]=low[node]=len(index); stack.append(node); active.add(node)
    for target in sorted(adjacency[node]):
        if target not in index: visit(target);low[node]=min(low[node],low[target])
        elif target in active: low[node]=min(low[node],index[target])
    if low[node]==index[node]:
        component=[]
        while True:
            member=stack.pop();active.remove(member);component.append(member)
            if member==node: break
        if len(component)>1 or node in adjacency[node]: components.append(sorted(component))
for file in sorted(original_files):
    if file not in index: visit(file)
write('module_cycles.json',components)

hosts=json.loads((DATA/'host_sites.json').read_text())
host_groups=collections.defaultdict(list)
for h in hosts: host_groups[(h['family'],h['receiver']+'.'+h['member'] if h['family']!='System' else h['member'])].append(h)
host_summary=[{'family':f,'api':api,'sites':len(v),'call_sites':sum(x['call'] for x in v),'files':sorted({x['file'] for x in v})} for (f,api),v in sorted(host_groups.items())]
write('host_summary.json',host_summary)
all_files=[]
for p in sorted((CACHE/'typescript/src/compiler').rglob('*')):
    if p.is_file(): all_files.append({'file':str(p.relative_to(CACHE/'typescript')),'bytes':p.stat().st_size,'source':p.suffix=='.ts'})
write('all_files.json',all_files)
summary={'original_files':len(original_files),'original_lines':sum(v['lines'] for v in original_files.values()),'all_original_files':len(all_files),'prepared_source_files':len(files),'stock_runs':len(stock),'upstream_runs':len(upstream),'stock_diagnostics':len(stock[-1].get('diagnostics',[])),'upstream_diagnostics':len(upstream[-1].get('diagnostics',[])),'stock_kinds':dict(collections.Counter(v['kind'] for v in stock)),'upstream_kinds':dict(collections.Counter(v['kind'] for v in upstream)),'stock_seconds':sum(v['seconds'] for v in stock),'upstream_seconds':sum(v['seconds'] for v in upstream),'source_sites':dict(collections.Counter(v['reason'] for v in sites if v['file'] in original_files)),'module_cycle_sizes':[len(v) for v in components],'host_families':dict(collections.Counter(v['family'] for v in hosts)),'non_source_runs':non_sources}
write('summary.json',summary)
print(json.dumps(summary,indent=2))
