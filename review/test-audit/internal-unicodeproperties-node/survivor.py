import pathlib,json,subprocess
p=pathlib.Path('review/test-audit/internal-unicodeproperties-node');base=json.load((p/'base.json').open());plan=json.load((p/'plan.json').open())
assert 'all sources restored' in (p/'runner.log').read_text()
f=pathlib.Path('internal/unicodeproperties/audit_witness_test.go');f.write_text((p/'survivor-witness.go.fixture').read_text());m=plan[1];source=pathlib.Path(m['file'])
try:
 for mid in ['clean','M02']:
  source.write_text(base[m['file']] if mid=='clean' else base[m['file']].replace(m['old'],m['new']))
  with (p/('survivor-'+mid+'.log')).open('w') as log:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/unicodeproperties/','-run','^TestAuditVersionWitness$'],stdout=log,stderr=subprocess.STDOUT)
finally:source.write_text(base[m['file']]);f.unlink()
