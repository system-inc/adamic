"""Independently verify unchanged observations, root edges and root-only counts."""
import collections
import copy
import csv
import gzip
import json
from pathlib import Path

OUT=Path(__file__).resolve().parent
ROOTS=OUT/'roots'
FIELDS=('unit','phase','kind','where','reason','text')
SITE=('kind','where','reason','text')
PREFIX='/tmp/stage3-notyet-adapted/'
local=lambda p:p.removeprefix(PREFIX)
with gzip.open(OUT/'full.jsonl.gz','rt') as stream:original=[json.loads(line) for line in stream]
with gzip.open(ROOTS/'full.jsonl.gz','rt') as stream:current=[json.loads(line) for line in stream]
assert current[0]==original[0], 'identical source reach and checker diagnostics'
assert len(current)==len(original)==82, 'complete entry census'
old_by_file={r['file']:r for r in original[1:]}
all_sites=collections.defaultdict(list);expected_edges=set()
for record in current[1:]:
    old=old_by_file[record['file']]
    assert record['units']==old['units'] and record['declarations']==old['declarations'], 'same attempt ledger'
    strip=lambda f:{k:v for k,v in f.items() if k not in ('blocked_by','blocked_symbol_declaration')}
    # Different provenance can retain multiple observations of a previously
    # identical signature. The underlying observation set must not change.
    encoded=lambda fs:{json.dumps(strip(f),sort_keys=True) for f in fs}
    assert encoded(record['findings'])==encoded(old['findings']), 'provenance changes no underlying observation'
    actual={tuple(f[k] for k in FIELDS):f for f in record['findings']}
    for f in record['findings']:
        if (f['phase'],f['kind'])!=('lowering','NotYet'):continue
        site=tuple(f[k] for k in SITE);all_sites[site].append(f)
        if cause:=f.get('blocked_by'):
            assert f['reason'].startswith('reading ') and f.get('blocked_symbol_declaration'), 'only symbol-linked missing reads'
            root=actual[tuple(cause[k] for k in FIELDS)]
            assert not root.get('blocked_by'), 'direct root, no echo chains'
            assert root['phase']=='lowering', 'actual lowerer root'
            root_site=tuple(root[k] for k in SITE)
            expected_edges.add((root_site,site))
assert len(all_sites)==10426, 'same unique NotYet observations'
root_sites={site for site,fs in all_sites.items() if any(not f.get('blocked_by') for f in fs)}
echo_sites=set(all_sites)-root_sites
mixed={site for site,fs in all_sites.items() if any(f.get('blocked_by') for f in fs) and any(not f.get('blocked_by') for f in fs)}
groups=collections.defaultdict(set)
for site in root_sites:groups[site[2]].add(site)
ranked=sorted(groups,key=lambda reason:(-len(groups[reason]),reason))
other_roots={root for root,echo in expected_edges if root[0]!='NotYet'}

def verify(report):
    assert report['original_notyet_sites']==len(all_sites), 'original count'
    assert report['notyet_root_sites']==len(root_sites) and report['echo_only_sites']==len(echo_sites), 'root count and subtraction'
    assert report['notyet_root_sites']+report['echo_only_sites']==report['original_notyet_sites'], 'partition'
    assert report['mixed_sites']==len(mixed), 'conservative mixed-site treatment'
    assert report['referenced_other_root_sites']==len(other_roots), 'separate Refused or measurement roots'
    assert report['root_reasons']==len(groups) and len(report['rows'])==len(groups), 'reason count'
    assert report['root_reasons_under_five']==sum(len(v)<5 for v in groups.values()), 'root tail'
    assert report['unattributed_read_sites']==sum(len(v) for k,v in groups.items() if k.startswith('reading ')), 'unattributed reads'
    for index,reason in enumerate(ranked):
        row=report['rows'][index];sites=groups[reason]
        assert row['reason']==reason and row['root_sites']==len(sites), 'root rank and count'
        assert row['files']==dict(collections.Counter(local(s[1]).rsplit(':',2)[0] for s in sites)), 'root diagnostic files'
        echoes={echo for root,echo in expected_edges if root in sites}
        assert row['echo_sites']==len(echoes), 'echoes under their root'
    for n in (5,10,20):
        count=sum(len(groups[r]) for r in ranked[:n]);coverage=report['coverage'][str(n)]
        assert coverage['sites']==count and abs(coverage['share']-100*count/len(root_sites))<1e-10, 'root coverage'

report=json.loads((ROOTS/'summary.json').read_text());verify(report)
with (ROOTS/'echoes.csv').open() as stream:edge_csv=list(csv.DictReader(stream))
expected_short={(r[0],local(r[1]),r[2],local(e[1]),e[2]) for r,e in expected_edges}
assert {(r['root_kind'],r['root_where'],r['root_reason'],r['echo_where'],r['echo_reason']) for r in edge_csv}==expected_short, 'complete root-to-echo CSV'
for row in edge_csv:
    echo=next(s for s in all_sites if local(s[1])==row['echo_where'] and s[2]==row['echo_reason'])
    assert (row['echo_only']=='True')==(echo in echo_sites) and (row['mixed']=='True')==(echo in mixed), 'edge disposition'
for name in ['root-count','echo-count','mixed-count','root-reason-count','echo-attribution','coverage','other-root-count']:
    changed=copy.deepcopy(report)
    if name=='root-count':changed['notyet_root_sites']+=1
    elif name=='echo-count':changed['echo_only_sites']+=1
    elif name=='mixed-count':changed['mixed_sites']+=1
    elif name=='root-reason-count':changed['rows'][0]['root_sites']+=1
    elif name=='echo-attribution':changed['rows'][0]['echo_sites']+=1
    elif name=='coverage':changed['coverage']['10']['share']+=1
    else:changed['referenced_other_root_sites']+=1
    try:verify(changed)
    except AssertionError as failure:print(name+' mutant caught:',failure)
    else:raise AssertionError(name+' mutant survived')
def verify_table(text):
    rendered=[line for line in text.splitlines() if line.startswith('| ') and len(line.split(' | '))==8 and line.split(' | ')[0][2:].isdigit()]
    assert len(rendered)==len(report['rows']), 'rendered root coverage'
    esc=lambda value:value.replace('|','&#124;').replace('\n','<br>')
    for line,row in zip(rendered,report['rows']):
        cells=[cell.strip() for cell in line[1:-1].split('|')]
        assert cells==[str(row['root_sites']),esc(row['reason']),str(row['echo_sites']),'<br>'.join(f'{esc(file)}: {count}' for file,count in row['files'].items()),'<br>'.join(row['examples']),'yes' if row['context_sensitive'] else 'no',row['disposition'],row['why']], 'rendered root and echo row'
    assert f"**{report['notyet_root_sites']:,} remaining NotYet root sites**" in text, 'rendered headline'
    for n in (5,10,20):
        c=report['coverage'][str(n)]
        assert f"| {n} | {c['sites']:,} | {c['share']:.2f}% |" in text, 'rendered shares'

table=(OUT/'TABLE.md').read_text();verify_table(table)
for name,changed in [('headline',table.replace('**6,791 remaining','**6,792 remaining',1)),('root-row',table.replace('| 1761 |','| 1762 |',1)),('echo-row',table.replace(' | 559 |',' | 560 |',1))]:
    try:verify_table(changed)
    except AssertionError as failure:print('rendered-'+name+' mutant caught:',failure)
    else:raise AssertionError('rendered-'+name+' mutant survived')

for name in ('nodeFactory','nodeFactoryBigInt'):
    replay=json.loads((ROOTS/'evidence'/f'{name}.json').read_text())
    command=json.loads((ROOTS/'evidence'/f'{name}-command.json').read_text());where=command[command.index('-where')+1]
    echo=next(f for f in replay['findings'] if (f['phase'],f['kind'],f['where'],f['reason'])==('lowering','NotYet',where,'reading node'))
    cause=echo['blocked_by'];assert tuple(cause[k] for k in FIELDS) in {tuple(f[k] for k in FIELDS) for f in replay['findings']}, 'replay root present'
    full=next(r for r in current[1:] if r['file']==replay['file'])
    assert any(all(f.get(k)==echo.get(k) for k in ('unit','phase','kind','where','reason','text','blocked_by','blocked_symbol_declaration')) for f in full['findings']), 'tagged replay agrees with full census'
# The additional checker replay reaches a different genuine initializer failure.
replay=json.loads((ROOTS/'evidence/checker.json').read_text())
echo=next(f for f in replay['findings'] if f['reason']=='reading node' and f['where'].endswith('checker.ts:2043:22'))
assert tuple(echo['blocked_by'][k] for k in FIELDS) in {tuple(f[k] for k in FIELDS) for f in replay['findings']}, 'checker replay has an actual root'
full=next(r for r in current[1:] if r['file']==replay['file'])
full_echo=next(f for f in full['findings'] if (f['unit'],f['where'],f['reason'])==(echo['unit'],echo['where'],echo['reason']))
assert echo['blocked_symbol_declaration']==full_echo['blocked_symbol_declaration'], 'checker same failed declaration'
assert echo['blocked_by']!=full_echo['blocked_by'], 'recorded checker root differs with replay context'
print('Checker root context difference preserved: full census',full_echo['blocked_by']['kind'],full_echo['blocked_by']['reason'],'; standalone replay',echo['blocked_by']['kind'],echo['blocked_by']['reason'])
print('PASS: unchanged source/checker/unit/observation sets; complete root/echo partition; conservative mixed sites; root counts/ranking/files/echo attribution/coverage; Refused roots separated; exact tagged replays')
print(json.dumps({k:v for k,v in report.items() if k not in ('rows','other_roots')},indent=2))
