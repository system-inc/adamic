"""Regroup the annotated census into root sites and symbol-linked echo edges."""
import collections
import csv
import gzip
import importlib.util
import json
from pathlib import Path
import sys

OUT = Path(__file__).resolve().parent
ROOTS = OUT/'roots'
PREFIX = '/tmp/stage3-notyet-adapted/'

def key(f):return tuple(f[k] for k in ('kind','where','reason','text'))
def local(where):return where.removeprefix(PREFIX)
def write_csv(path, rows, fields):
    with path.open('w',newline='') as stream:
        writer=csv.DictWriter(stream,fieldnames=fields,lineterminator='\n');writer.writeheader();writer.writerows(rows)

def calculate(records):
    observations=[];sites=collections.defaultdict(list);edges=collections.defaultdict(set);root_values={}
    for record in records[1:]:
        actual={tuple(f[k] for k in ('unit','phase','kind','where','reason','text')):f for f in record['findings']}
        for f in record['findings']:
            if (f['phase'],f['kind'])!=('lowering','NotYet'):continue
            row=dict(f,attempt_file=record['file']);observations.append(row);sites[key(f)].append(row)
            if cause:=f.get('blocked_by'):
                assert f['reason'].startswith('reading ') and f.get('blocked_symbol_declaration'), 'symbol-linked read only'
                assert tuple(cause[k] for k in ('unit','phase','kind','where','reason','text')) in actual, 'actual root in same attempt record'
                root=actual[tuple(cause[k] for k in ('unit','phase','kind','where','reason','text'))]
                assert not root.get('blocked_by'), 'flattened root'
                edges[key(cause)].add(key(f));root_values[key(cause)]=cause
    echo_only={site for site,rows in sites.items() if all(r.get('blocked_by') for r in rows)}
    mixed={site for site,rows in sites.items() if any(r.get('blocked_by') for r in rows) and not all(r.get('blocked_by') for r in rows)}
    primary={site for site in sites if site not in echo_only}
    for site in primary:root_values.setdefault(site,sites[site][0])
    groups=collections.defaultdict(set)
    for site in primary:groups[site[2]].add(site)
    spec=importlib.util.spec_from_file_location('original_table',OUT/'build-table.py');module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
    def grouped(reason, members):
        files=collections.Counter(local(site[1]).rsplit(':',2)[0] for site in members)
        examples=sorted({local(site[1]) for site in members},key=lambda s:(s.rsplit(':',2)[0],*map(int,s.rsplit(':',2)[1:])))[:2]
        echoes=set().union(*(edges[site] for site in members))
        disposition,why=module.disposition(reason)
        if reason.startswith('reading '):
            disposition='unattributed read'
            why='No failed-declaration provenance proves an echo in every attempt; retain conservatively until its complete binding context is settled.'
        return dict(reason=reason,root_sites=len(members),echo_sites=len(echoes),files=dict(sorted(files.items())),examples=examples,context_sensitive=reason.startswith('reading '),disposition=disposition,why=why)
    rows=sorted((grouped(reason,members) for reason,members in groups.items()),key=lambda r:(-r['root_sites'],r['reason']))
    other=collections.defaultdict(set)
    for site in edges:
        if site[0]!='NotYet':other[(site[0],site[2])].add(site)
    other_rows=[]
    for (kind,reason),members in sorted(other.items(),key=lambda item:(-len(item[1]),item[0])):
        row=grouped(reason,members);row['kind']=kind;other_rows.append(row)
    edge_rows=[]
    for root,echoes in sorted(edges.items()):
        for echo in sorted(echoes):
            matching=[r for r in sites[echo] if r.get('blocked_by') and key(r['blocked_by'])==root]
            edge_rows.append(dict(root_kind=root[0],root_where=local(root[1]),root_reason=root[2],echo_where=local(echo[1]),echo_reason=echo[2],echo_only=echo in echo_only,mixed=echo in mixed,declarations='; '.join(sorted({local(r['blocked_symbol_declaration']) for r in matching})),units='; '.join(sorted({local(r['unit']) for r in matching}))))
    summary=dict(original_notyet_sites=len(sites),raw_notyet_observations=len(observations),echo_only_sites=len(echo_only),tagged_echo_sites=len(echo_only|mixed),mixed_sites=len(mixed),notyet_root_sites=len(primary),root_reasons=len(rows),root_reasons_under_five=sum(r['root_sites']<5 for r in rows),unattributed_read_sites=sum(r['root_sites'] for r in rows if r['context_sensitive']),referenced_other_root_sites=sum(len(v) for v in other.values()),referenced_other_roots_by_kind=dict(collections.Counter(site[0] for site in edges if site[0]!='NotYet')),coverage={str(n):dict(sites=sum(r['root_sites'] for r in rows[:n]),share=100*sum(r['root_sites'] for r in rows[:n])/len(primary)) for n in (5,10,20)},rows=rows,other_roots=other_rows)
    return summary,observations,edge_rows

if __name__=='__main__':
    raw=Path(sys.argv[1]);records=[json.loads(line) for line in raw.read_text().splitlines()]
    summary,observations,edges=calculate(records)
    ROOTS.mkdir(exist_ok=True)
    (ROOTS/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
    flat=[]
    for f in observations:
        cause=f.get('blocked_by',{})
        flat.append(dict(kind=f['kind'],phase=f['phase'],where=local(f['where']),reason=f['reason'],text=f['text'],attempt_file=local(f['attempt_file']),unit=local(f['unit']),blocked_symbol_declaration=local(f.get('blocked_symbol_declaration','')),**{'blocked_by_'+k:local(cause.get(k,'')) if k in ('where','unit') else cause.get(k,'') for k in ('kind','where','reason','text','unit','phase')}))
    write_csv(ROOTS/'raw.csv',flat,list(flat[0]))
    write_csv(ROOTS/'echoes.csv',edges,['root_kind','root_where','root_reason','echo_where','echo_reason','echo_only','mixed','declarations','units'])
    with gzip.open(ROOTS/'full.jsonl.gz','wb') as stream:stream.write(raw.read_bytes())
    print(json.dumps({k:v for k,v in summary.items() if k not in ('rows','other_roots')},indent=2))
