"""Check raw/derived counts, source identity, reproducer outcomes and host evidence.

The --mutants mode corrupts one artifact at a time and requires its specific
consistency check to reject it. No compiler implementation is mutated.
"""
import collections
import copy
import gzip
import hashlib
import json
import pathlib
import re
import sys

HERE=pathlib.Path(__file__).resolve().parent
DATA=HERE/'data'
CACHE=pathlib.Path(sys.argv[1])

def load(name): return json.loads((DATA/name).read_text())
def raw(name):
    with gzip.open(DATA/(name+'.jsonl.gz'),'rt') as stream: return [json.loads(line) for line in stream]

def verify(data, source_override=None):
    texts={}
    files=data['files']; originals={f['file']:f for f in files if not f['generated']}
    assert len(originals)==77, 'original source coverage'
    for item in files:
        source=(CACHE/'prepared'/item['file']).read_bytes()
        if source_override and item['file']==source_override[0]: source=source_override[1]
        assert hashlib.sha256(source).hexdigest()==item['sha256'], 'source hash'
        assert len(source)==item['bytes'], 'source byte count'
        assert len(source.decode().split('\n'))==item['lines'], 'source line count'
        texts[item['file']]=source.decode().splitlines()
    for profile,expected in [('stock',77),('upstream',78)]:
        observations=raw(profile)
        entries=[r for r in observations[:-1] if len(r['roots'])==1 and r['roots'][0].endswith(('.ts','.a'))]
        assert len(entries)==expected, 'run coverage'
        names={r['roots'][0].split('/src/compiler/')[-1] for r in entries}
        expected_names={f['file'].removeprefix('src/compiler/') for f in (files if profile=='upstream' else originals.values())}
        assert names==expected_names, 'duplicate or missing entry'
        assert len(observations[-1]['roots'])==expected, 'whole-program roots'
        assert all(r['kind']=='checker' for r in entries+[observations[-1]]), 'unreported lower result'
    assert data['upstream_diagnostics']==[], 'outside oracle does not accept upstream source'
    count=collections.Counter(s['reason'] for s in data['sites'] if s['file'] in originals)
    assert dict(count)==data['summary']['source_sites'], 'source site count'
    for s in data['sites']:
        text=texts[s['file']]
        assert text[s['line']-1]==s['source'], 'source site line'
        assert 1 <= s['column'] <= len(text[s['line']-1].encode('utf-16-le'))//2+1, 'source column'
    for ranking in data['rankings']:
        reason=ranking['reason']
        population=[d for d in data['diagnostics'] if d.get('file') in originals and d['profile']=='stock' and 'TS'+str(d.get('code'))==reason] if reason.startswith('TS') else [s for s in data['sites'] if s['file'] in originals and s['reason']==reason]
        files_seen={s['file'] for s in population}
        lines_seen={(s['file'],s['line']) for s in population}
        assert ranking['file_count']==len(files_seen), 'rank file count'
        assert ranking['start_line_count']==len(lines_seen), 'rank line count'
        assert ranking['occurrences']==len(population), 'rank occurrence count'
        assert ranking['affected_file_lines']==sum(originals[f]['lines'] for f in files_seen), 'rank affected file lines'
        observed=ranking['observed']
        if reason.startswith('TS'): assert any('error '+reason+':' in d for d in observed.get('diagnostics',[])), 'minimal checker reproducer'
        else:
            assert observed['kind'] in ['NotYet','Refused','checker'], 'minimal source reproducer'
            if observed['kind']=='checker':
                assert ranking['policy_observed']['kind']=='Refused', 'policy refusal remains masked'
        assert (HERE/ranking['repro']).exists(), 'missing reproducer file'
    ordered=sorted(data['rankings'],key=lambda r:(-r['file_count'],-r['start_line_count'],r['reason']))
    assert [r['reason'] for r in ordered[:30]]==[r['reason'] for r in data['top30']], 'top30 ordering'
    host_counts=collections.Counter(s['family'] for s in data['host_sites'])
    assert dict(host_counts)==data['summary']['host_families'], 'host site count'
    assert len(data['system_contract'])==44, 'System contract coverage'
    for record in data['host_runtime']['results']:
        expected=[] if record['name']=='good' else [
            {'code':2322,'message':"Type 'string' is not assignable to type 'number'.",'file':'bad.ts','line':1,'column':7},
            {'code':2345,'message':"Argument of type 'string' is not assignable to parameter of type 'number'.",'file':'bad.ts','line':1,'column':106}]
        assert record['diagnostics']==expected, 'host exact diagnostics'
        assert record['calls']['ts.sys.readFile']==record['calls']['fs.readFileSync'], 'host read accounting'

data={name:load(name+'.json') for name in ['files','sites','summary','diagnostics','rankings','top30','upstream_diagnostics','host_sites','system_contract','host_runtime']}
verify(data)
print('PASS: source hashes, every entry and whole program, source lines, all ranking counts, 30 reproducers, System contract and exact host diagnostics')
if '--mutants' in sys.argv:
    mutants=[]
    changed=copy.deepcopy(data);changed['rankings'][0]['file_count']+=1
    mutants.append(('inflate highest file rank','rank file count',changed,None))
    changed=copy.deepcopy(data);index=next(i for i,s in enumerate(changed['sites']) if s['reason']=='an index signature');changed['sites'].pop(index)
    mutants.append(('omit one index-signature site','source site count',changed,None))
    item=data['files'][0];contents=(CACHE/'prepared'/item['file']).read_bytes()
    mutants.append(('alter one pinned source byte','source hash',data,(item['file'],b'X'+contents[1:])))
    changed=copy.deepcopy(data);changed['host_runtime']['results'][1]['diagnostics'].pop()
    mutants.append(('drop the incompatible-argument diagnostic','host exact diagnostics',changed,None))
    for name,expected,changed,override in mutants:
        try: verify(changed,override)
        except AssertionError as error:
            assert str(error)==expected,(name,str(error),expected)
            print('CAUGHT: '+name+' by '+expected)
        else: raise AssertionError('mutant survived: '+name)
