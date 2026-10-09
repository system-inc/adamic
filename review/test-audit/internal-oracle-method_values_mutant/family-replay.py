import pathlib,json,os,subprocess,time
p=pathlib.Path('review/test-audit/internal-oracle-method_values_mutant');base=json.load((p/'base.json').open())
pattern=r'^TestNativeAgreesWithNode$/^(internal|stage3)$/^(oracle|namespace-live-export)$/^(testdata|live[.]a)$/^(library_method_values.*|module_namespace_reads|namespaces.*|e4eec87_f2.*|narrowed_union_valid[.]a)$'
(p/'family-pattern.txt').write_text(pattern)
helper=pathlib.Path('internal/lower/audit_mutant.go')
for f in (p/'switched-source').glob('*.fixture'):pathlib.Path('internal/lower/'+f.name.removesuffix('.fixture')).write_text(f.read_text())
try:
 for mid in ['clean0','clean1','clean2']+[f'M{i:02}' for i in range(1,11)]:
  env=os.environ.copy();env['ADAMIC_MUTANT']='' if mid.startswith('clean') else mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u063/cache/'+mid
  s=time.monotonic();cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern]
  with (p/('family-'+mid+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  with (p/'family-runs.jsonl').open('a') as log:log.write(json.dumps({'id':mid,'cmd':cmd,'exit':r.returncode,'seconds':time.monotonic()-s})+'\n')
  if mid.startswith('clean') and r.returncode:raise SystemExit('new family baseline red; stop replay')
finally:
 for f,t in base.items():
  if f.startswith('internal/lower/'):pathlib.Path(f).write_text(t)
 if helper.exists():helper.unlink()
print('family replay complete, source restored')
