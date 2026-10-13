"""List every observable site/table difference on identical historical inputs.
Usage: compare_main.py OLD_EVIDENCE NEW_BUFFERED NEW_STREAMED OUTPUT
Incomplete directory measurements are compared only on completed file records.
"""
from collections import Counter
import gzip
import json
from pathlib import Path
import sys
old, buffered, streamed, output = map(Path, sys.argv[1:])
output.mkdir(parents=True, exist_ok=True)
attribution = {'?.[] on a value': 'main cfa29460 optional indexing', 'an ElementAccessExpression': 'main 735347bb finite field indexing', 'a computed key without an own data field': 'main 735347bb computed field read guard', 'checked view union arm awaits views-v3: array element kind': 'main 4556d540 checked union dispatch', 'a union or optional field read in a program with record storage': 'main 8c962b67 contextual contracts and record aliases'}
results = []
for scope in ('_namespaces', 'factory', 'transformers/declarations', 'transformers/module'):
    before = [json.loads(x) for x in gzip.decompress((old/scope/'speculative.jsonl.gz').read_bytes()).decode().splitlines()]
    target = (buffered if scope in ('_namespaces','transformers/declarations') else streamed)/scope
    after = [json.loads(x) for x in (target/'speculative.jsonl').read_text().splitlines()]
    oldroot = str(Path(before[1]['file']).parent) if '/' not in scope else str(Path(before[1]['file']).parent)
    newroot = str(Path(after[1]['file']).parent)
    # All records in these directory projects are direct children of their root.
    complete = {Path(x['file']).name for x in after[1:]}
    def sites(rows, root):
        counts = Counter()
        for row in rows[1:]:
            if Path(row['file']).name not in complete: continue
            for f in row['findings']:
                if f['kind'] not in ('NotYet','Refused') or not f['where'].startswith(root+'/'): continue
                key = tuple(str(f.get(k,'')).replace(root,'ROOT') for k in ('kind','where','reason','text','site_where','site_kind','site_start','site_end'))
                counts[(key,f['depth'])] = 1
        return counts
    a,b = sites(before,oldroot),sites(after,newroot)
    def table(c):
        out={k:[0]*5 for k in ('NotYet','Refused')}
        for (key,depth),n in c.items(): out[key[0]][min(4,depth)]+=n
        return out
    removed=[dict(site=list(k),depth=d) for k,d in sorted(a.keys()-b.keys())]
    added=[dict(site=list(k),depth=d) for k,d in sorted(b.keys()-a.keys())]
    coverage_differences=[]
    old_rows={Path(r['file']).name:r for r in before[1:]}
    for row in after[1:]:
        filename=Path(row['file']).name
        a_cov=old_rows[filename]['speculative_coverage'];b_cov=row['speculative_coverage']
        for key in sorted(a_cov.keys()|b_cov.keys()):
            if a_cov.get(key)!=b_cov.get(key):coverage_differences.append(dict(file=filename,field=key,old=a_cov.get(key),main=b_cov.get(key)))
    old_result=json.loads((old.parent.parent/'directories'/scope/'RESULT.json').read_text()) if (old.parent.parent/'directories'/scope/'RESULT.json').exists() else None
    new_result=json.loads((target/'RESULT.json').read_text())
    results.append(dict(scope=scope, completed_files=sorted(complete), historical_files=len(before)-1, coverage_differences=coverage_differences,
                        old_common_counts=table(a),new_common_counts=table(b),removed=removed,added=added,
                        attribution={reason: dict(origin=attribution.get(reason, 'main lowering changed; individual change not isolated'), confidence='source-history inference; no individual compiler replay') for reason in sorted({s['site'][2] for s in removed+added})},
                        historical_complete_counts=old_result and old_result['counts'],new_published_counts=new_result['counts'],
                        old_top20=old_result and old_result['top20'],new_top20=new_result['top20']))
(output/'DIFFERENCES.json').write_text(json.dumps(results,indent=2)+'\n')
lines=['Compared identical f6bb0b41 adapted source bytes against main.','Only completed file records enter the common-file comparison; missing files have no measured delta.','Every added/removed unique observation and both top-20 tables are in DIFFERENCES.json.','Main changed lowering and registration since a5630a90. Port controls pass, but they do not establish the cause of each individual changed compiler site.','Individual attribution remains an inference unless backed by a compiler-change replay; these differences must not all be claimed as proven main changes.','', '| Directory | Completed / old files | Old common NY / Ref | Main common NY / Ref | Removed / added observations |','|---|---:|---:|---:|---:|']
for r in results:
    a=r['old_common_counts'];b=r['new_common_counts']
    lines.append(f"| {r['scope']} | {len(r['completed_files'])} / {r['historical_files']} | {sum(a['NotYet'])} / {sum(a['Refused'])} | {sum(b['NotYet'])} / {sum(b['Refused'])} | {len(r['removed'])} / {len(r['added'])} |")
(output/'README.md').write_text('\n'.join(lines)+'\n')
print('\n'.join(lines))
