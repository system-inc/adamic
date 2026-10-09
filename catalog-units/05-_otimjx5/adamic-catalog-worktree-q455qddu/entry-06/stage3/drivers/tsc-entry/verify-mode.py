#!/usr/bin/env python3
"""Check mode evidence against raw loader output, split builds and Node witnesses."""
import argparse,json,re
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('evidence',nargs='?',type=Path,default=Path(__file__).resolve().parent/'evidence/mode-c09c22b8');a=p.parse_args();e=a.evidence
rows=json.loads((e/'admission.json').read_text());assert [r['ordinal'] for r in rows]==list(range(1,16))
expected=['not-diagnosed','not-diagnosed','scheduled','not-diagnosed','not-diagnosed','not-diagnosed','not-diagnosed','not-diagnosed','ordinary-error','not-diagnosed','not-diagnosed','scheduled','scheduled','ordinary-error','not-diagnosed']
for r,want in zip(rows,expected):
 raw=json.loads((e/f"{r['ordinal']:02d}-load.json").read_text())
 matches=[d for d in raw['dispositions'] or [] if d['site']['file'].endswith('/'+r['file']) and d['site']['line']==r['line'] and d['site']['column']==r['column'] and d['site']['code']==int(re.search(r'TS(\d+)',r['message'])[1])]
 ordinary=[d for d in raw['ordinary_diagnostics'] or [] if f"/{r['file']}:{r['line']}:{r['column']}: " in d]
 status='scheduled' if matches and all(d['state']=='scheduled-check' for d in matches) else 'ordinary-error' if ordinary else 'not-diagnosed' if raw['loaded'] else 'blocked'
 assert status==r['mode_status']==want,(r['ordinal'],status)
 assert r['dispositions']==matches and r['ordinary_diagnostics']==ordinary
 assert r['project_loaded']==raw['loaded']
for suffix in ['stdout','stderr','exit']:assert (e/f'entry-0.{suffix}').read_bytes()==(e/f'entry-1.{suffix}').read_bytes()
assert (e/'entry-0.exit').read_text().strip()=='1'
assert re.match(r'.*/src/tsc/_namespaces/ts.ts:3:15: error TS6305:',(e/'entry-0.stderr').read_text())
i=json.loads((e/'identity.json').read_text());assert i['before']==i['after'] and len(i['before'])>700 and i['patch_set_equal'] and not i['main_control_diff']
roots=json.loads((e/'ledger-roots.json').read_text());assert len(roots)==79 and 'src/compiler/builder.ts' in roots
w=e/'witnesses'
for folder,output in [('app','7\n'),('outside','2\n'),('ordinary-09','real\n'),('ordinary-14','ok\n')]:
 assert (w/folder/'node.exit').read_text().strip()=='0' and (w/folder/'node.stdout').read_text()==output
 assert (w/folder/'node.stderr').read_bytes()==b''
assert 'error TS6305:' in (w/'app/build.stderr').read_text()
out=json.loads((w/'outside/load.json').read_text());assert out['loaded'] and not out['ordinary_diagnostics']
assert any(d['state']=='scheduled-check' and d['kind']=='indexed-presence' for d in out['dispositions'])
assert all(not d['site']['file'].endswith('/'+root) for d in out['dispositions'] for root in roots)
assert 'stage 0 can\'t lower the library method log yet' in (w/'outside/build.stderr').read_text()
for folder,code in [('ordinary-09','TS2345'),('ordinary-14','TS2322')]:assert 'error '+code+':' in (w/folder/'build.stderr').read_text()
print('PASS: fifteen classifications, raw loader sites, both split builds, source/control identity, project scope and four Node witnesses')
