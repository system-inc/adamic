import pathlib,json,difflib,subprocess,time
p=pathlib.Path('review/test-audit/internal-oracle-private_generic_mutant');root=pathlib.Path.cwd();rows=json.loads((p/'scope.json').read_text());names=[r['test'] for r in rows if r['test'] not in ['TestRuntimeLastIndexOfMatchesNode','TestScannerNestedReferences','TestRegexCycleFixtureHasItsNativeDependency']];replacements={}
for file,a,b in [('internal/oracle/oracle_test.go','func disagreement(oracle run, native run) string {','func disagreement(oracle run, native run) string {\n return ""'),('internal/oracle/non_null_migration_test.go','string(got.stdout) == string(native.stdout)','true')]:
 original=(root/file).read_text();assert original.count(a)==1;weak=original.replace(a,b)
 if file.endswith('oracle_test.go'):
  begin=original.index(a);end=original.index('\n}\n\nfunc TestNativeAgreesWithNode',begin)+2;weak=original[:begin]+a+'\n return \"\"\n}'+original[end:]
 tmp=pathlib.Path('/tmp/u068-'+pathlib.Path(file).name);tmp.write_text(weak);replacements[str(root/file)]=str(tmp);(p/('W-'+pathlib.Path(file).name+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),weak.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
overlay=pathlib.Path('/tmp/u068-W-overlay.json');overlay.write_text(json.dumps({'Replace':replacements}));cmd=['timeout','120','go','test','-json','-overlay',str(overlay),'-count=1','-timeout','90s','./internal/oracle/','-run','^('+'|'.join(names)+')$'];start=time.monotonic()
with (p/'W.log').open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
(p/'W-run.json').write_text(json.dumps({'command':'ADAMIC_GATE_UNCACHED=1 '+' '.join(cmd),'exit':r.returncode,'wall':time.monotonic()-start},indent=2))
# Break the dependency-presence construction, leaving its assertion unchanged.
file='internal/oracle/regexp_cycle_test.go';orig=(root/file).read_text();a='return present';assert orig.count(a)==1;weak=orig.replace(a,'return present && false');tmp=pathlib.Path('/tmp/u068-cycle.go');tmp.write_text(weak);overlay=pathlib.Path('/tmp/u068-S01-overlay.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(tmp)}}));(p/'S01.diff').write_text(''.join(difflib.unified_diff(orig.splitlines(True),weak.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
cmd=['timeout','120','go','test','-json','-overlay',str(overlay),'-count=1','-timeout','90s','./internal/oracle/','-run','^TestRegexCycleFixtureHasItsNativeDependency$'];start=time.monotonic()
with (p/'S01.log').open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
(p/'S01-run.json').write_text(json.dumps({'command':'ADAMIC_GATE_UNCACHED=1 '+' '.join(cmd),'exit':r.returncode,'wall':time.monotonic()-start},indent=2))
