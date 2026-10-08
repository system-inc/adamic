"""Independently recount artifacts and prove their assertions reject corruption."""
import collections
import copy
import csv
import gzip
import hashlib
import json
from pathlib import Path
import sys

out=Path(sys.argv[1]); adapted=Path(sys.argv[2]).resolve()
with gzip.open(out/'full.jsonl.gz','rt') as stream: records=[json.loads(line) for line in stream]
raw=list(csv.DictReader((out/'raw.csv').open()))
summary=json.loads((out/'summary.json').read_text())
manifest=json.loads((out/'source-manifest.json').read_text())
old=json.loads(Path('stage3/meter/runs/20261008T035244Z.latent-full/tsc/source-manifest.json').read_text())
assert sorted(manifest,key=lambda r:r['file'])==sorted(old,key=lambda r:r['file']), 'prior source manifest'
for item in manifest:
    data=(adapted/item['file']).read_bytes()
    assert len(data)==item['bytes'] and hashlib.sha256(data).hexdigest()==item['sha256'], 'source hash'
header=records[0]
assert header['latent_mode']=='full' and header['checker_rejected'] and len(header['sources'])==81, 'scope'
assert set(header['sources'])=={str(adapted/item['file']) for item in manifest}, 'reach'
# The source bytes are identical, so the saved stock TypeScript AST ledger is
# an independent reference for declaration coverage and direct diagnostic spans.
import importlib.util
module_path=Path('stage3/census/latent/full_report.py').resolve()
spec=importlib.util.spec_from_file_location('notyet_full_report',module_path)
ledger=importlib.util.module_from_spec(spec); spec.loader.exec_module(ledger)
stock=json.loads(Path('stage3/meter/runs/20261008T035244Z.latent-full/tsc/stock-units.json').read_text())
ledger.audit_units(records,stock,adapted)
for name in ('missing-nested-unit','body-span','diagnostic-ownership'):
    changed=copy.deepcopy(records)
    if name=='missing-nested-unit':
        record=next(r for r in changed[1:] if any(u.get('depth',0)>0 for u in r['units']))
        record['units'].remove(next(u for u in record['units'] if u.get('depth',0)>0))
    elif name=='body-span':
        unit=next(u for r in changed[1:] for u in r['units'] if u['kind']=='KindFunctionDeclaration' and u.get('body_end'))
        unit['body_start']+=1
    else:
        unit=next(u for r in changed[1:] for u in r['units'] if u['kind']=='KindFunctionDeclaration' and u.get('body_end'))
        unit['checker_diagnostics']=['planted checker diagnostic']
    try:ledger.audit_units(changed,stock,adapted)
    except AssertionError as failure:print(name+' mutant caught: '+str(failure))
    else:raise AssertionError(name+' mutant survived')
print('PASS: stock TypeScript declaration coverage, spans, parents, direct checker diagnostic ownership and body eligibility')
expected=[]
for record in records[1:]:
    for f in record['findings']:
        if (f['kind'],f['phase'])!=('NotYet','lowering'): continue
        def relative(where):
            path,line,column=where.rsplit(':',2)
            return str(Path(path).relative_to(adapted))+':'+line+':'+column
        expected.append((f['kind'],f['phase'],relative(f['where']),f['reason'],f['text'],
                         str(Path(record['file']).relative_to(adapted)),relative(f['unit']),
                         str(f['reason'].startswith('reading ')),f['measurement']))
fields=['kind','phase','where','reason','text','attempt_file','unit','context_sensitive','measurement']
site_set={(row[0],row[2],row[3],row[4]) for row in expected}
reason_sites=collections.defaultdict(set)
for site in site_set:reason_sites[site[2]].add(site)
ranked=sorted(reason_sites,key=lambda reason:(-len(reason_sites[reason]),reason))

def verify(raw_rows,report):
    assert collections.Counter(tuple(r[k] for k in fields) for r in raw_rows)==collections.Counter(expected), 'raw filter/observations'
    assert report['total']==len(site_set) and report['raw_observations']==len(expected), 'totals'
    assert report['reasons']==len(ranked) and len(report['rows'])==len(ranked), 'reason count'
    assert report['context_sensitive_sites']==sum(len(v) for k,v in reason_sites.items() if k.startswith('reading ')), 'context total'
    assert report['reasons_under_five']==sum(len(v)<5 for v in reason_sites.values()), 'tail'
    for index,reason in enumerate(ranked):
        r=report['rows'][index]; sites=reason_sites[reason]
        assert r['reason']==reason and r['count']==len(sites), 'rank/count'
        files=collections.Counter(site[1].rsplit(':',2)[0] for site in sites)
        assert r['files']==dict(files), 'file attribution'
        positions=sorted({site[1] for site in sites},key=lambda s:(s.rsplit(':',2)[0],int(s.rsplit(':',2)[1]),int(s.rsplit(':',2)[2])))[:2]
        assert r['examples']==positions, 'examples'
        assert r['context_sensitive']==reason.startswith('reading '), 'context flag'
        adaptation=('writable view' in reason or 'writable-view' in reason or 'mutable variance' in reason or reason == 'a writable-slot checked view requiring source contract certification' or reason in ('a value of type any','a parameter of type any','a function returning any','a call returning any','an array of any','a field of type any','a DebuggerStatement','a WithStatement','a VoidExpression','a VoidExpression as a statement'))
        assert r['disposition']==('adaptation' if adaptation else 'compiler lesson') and r['why'], 'disposition'
    for n in (5,10,20):
        numerator=sum(len(reason_sites[r]) for r in ranked[:n])
        assert report['coverage'][str(n)]['sites']==numerator and abs(report['coverage'][str(n)]['share']-numerator/len(site_set)*100)<1e-10, 'coverage'

verify(raw,summary)
# Check the user-facing table against the independently checked artifact.
table=(out/'roots/original-TABLE.md' if (out/'roots/original-TABLE.md').exists() else out/'TABLE.md').read_text()
esc=lambda s:s.replace('|','&#124;').replace('\n','<br>')
table_rows=[line for line in table.splitlines() if line.startswith('| ') and len(line.split(' | '))==7 and line.split(' | ')[0][2:].isdigit()]
assert len(table_rows)==len(ranked), 'table coverage'
for rendered,r in zip(table_rows,summary['rows']):
    cells=[c.strip() for c in rendered[1:-1].split('|')]
    assert cells==[str(r['count']),esc(r['reason']),'<br>'.join(f'{esc(f)}: {n}' for f,n in sorted(r['files'].items())),'<br>'.join(r['examples']),'yes' if r['context_sensitive'] else 'no',r['disposition'],r['why']], 'table row'
for n in (5,10,20):
    c=summary['coverage'][str(n)]
    assert f"| {n} | {c['sites']:,} | {c['share']:.2f}% |" in table, 'table shares'
for name in ['drop-observation','wrong-phase','total','reason-count','file-attribution','example','context','disposition','writable-variance-disposition','coverage','tail']:
    rows,report=copy.deepcopy(raw),copy.deepcopy(summary)
    if name=='drop-observation':rows.pop()
    elif name=='wrong-phase':rows[0]['phase']='refusal_scan'
    elif name=='total':report['total']+=1
    elif name=='reason-count':report['rows'][0]['count']+=1
    elif name=='file-attribution':report['rows'][0]['files']={'src/compiler/wrong.ts':report['rows'][0]['count']}
    elif name=='example':report['rows'][0]['examples'][0]='src/compiler/wrong.ts:1:1'
    elif name=='context':report['rows'][0]['context_sensitive']=not report['rows'][0]['context_sensitive']
    elif name=='disposition':report['rows'][0]['disposition']='adaptation' if report['rows'][0]['disposition']=='compiler lesson' else 'compiler lesson'
    elif name=='writable-variance-disposition':
        row=next(r for r in report['rows'] if r['reason']=='a writable-slot checked view requiring source contract certification')
        row['disposition']='compiler lesson'
    elif name=='coverage':report['coverage']['10']['share']+=1
    elif name=='tail':report['reasons_under_five']+=1
    try:verify(rows,report)
    except AssertionError as failure:print(name+' mutant caught: '+str(failure))
    else:raise AssertionError(name+' mutant survived')
print('PASS: exact filtered observations, site deduplication, source hashes/reach, reason ranking, file attribution, examples, context flags, dispositions, shares, tail and rendered table')
print('checker diagnostics:',len(header['diagnostics']))
print('unit statuses:',dict(collections.Counter(u['status'] for r in records[1:] for u in r['units'])))
print('non-NotYet lowering kinds:',dict(collections.Counter(f['kind'] for r in records[1:] for f in r['findings'] if f['phase']=='lowering' and f['kind']!='NotYet')))
