#!/usr/bin/env python3
import gzip,hashlib,json,re,sys
from pathlib import Path
here=Path(sys.argv[1]) if len(sys.argv)>1 else Path(__file__).resolve().parent
d=json.loads((here/'result.json').read_text());reasons=json.loads((here/'ranking-reasons.json').read_text())
assert d['candidate']=='79dd1abae85972b8123fb813d2d98e49ef612730'
assert d['stricter']=='8f32e51e8fc41b8f1177453213ca5453ce764486'
assert d['driver_bytes_unchanged'] and not d['native_comparator_run']
assert 'vcs.revision='+d['scratch_merge'] in (here/'binary.txt').read_text()
assert 'vcs.modified=false' in (here/'binary.txt').read_text()
assert len(d['profiles'])==2
for row in d['profiles']:
    profile=row['profile'];a,b=row['builds']
    assert a['exit']==b['exit']==1 and not row['binary_exists']
    assert a['stderr']==b['stderr'] and a['stdout']==b['stdout']
    for split in (0,1):
        assert (here/'evidence'/f'{profile}-{split}.stderr').read_text()==row['builds'][split]['stderr']
    assert row['first_stop']==a['stderr'].splitlines()[0]
    assert row['ranking_matches']==[r for r in reasons if row['reason']==r['kind']+': '+r['reason']]
assert 'error TS2591:' in d['profiles'][0]['first_stop']
assert d['profiles'][1]['reason']=='NotYet: node:module.require'
assert d['profiles'][0]['owner']=='compiler' and d['profiles'][1]['owner']=='library'
for name in ('probe','typed_probe'):
    p=d[name];assert p['build']['exit']==1
    assert p['node']['exit']==p['mutant']['exit']==0
    assert p['node']['stderr']==p['mutant']['stderr']==''
    assert p['node']['stdout']=='leaf\n' and p['mutant']['stdout']=='changed\n'
assert d['node_manifest']['selected'] and len(d['node_manifest']['projects'])==301
assert len(d['node_manifest']['selected'])==301
assert hashlib.sha256(gzip.decompress((here/'evidence/node-golden.stdout.gz').read_bytes())).hexdigest()==d['node_manifest']['sha256']
assert hashlib.sha256(gzip.decompress((here/'evidence/node-request.json.gz').read_bytes())).hexdigest()==d['node_manifest']['requestSha256']
assert d['node_manifest']['sha256']=='f5b3ccb280ac67ed516240b52b476a9748dd8a60a516cec86f76be371265828a'
assert (here/'probes/commonjs.a').read_text().startswith('// a-check: type error TS2591\n')
assert not (here/'probes/node-require.a').read_text().startswith('// a-check:')
print('Two split profiles, two Node witnesses, two computed-input mutants and 301 Node observations verified; native comparator remains unrun')
