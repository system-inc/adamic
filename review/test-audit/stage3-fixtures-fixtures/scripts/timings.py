import pathlib,subprocess,time,json
r=pathlib.Path('/workspace/adamic/review/test-audit/stage3-fixtures-fixtures');results=[]
for row,pattern in [('family','^TestFixtures'),('prepare','^TestPrepareFixtureOracleHook$'),('paths','^TestFixturePaths$'),('manifest','^TestFixtureShardManifest$')]:
 for i in range(1,4):
  t=time.monotonic()
  with (r/'logs'/f'timing-{row}-{i}.log').open('w') as f:rc=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage3/fixtures/','-run',pattern],stdout=f,stderr=subprocess.STDOUT).returncode
  results.append(dict(row=row,run=i,pattern=pattern,wall=time.monotonic()-t,rc=rc));(r/'timing-runs.json').write_text(json.dumps(results,indent=2))
  if rc:raise RuntimeError('clean timing failed')
with (r/'logs/coverage.log').open('w') as f:rc=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/lower','-coverprofile='+str(r/'lower.cover'),'./stage3/fixtures/','-run','^TestFixtures'],stdout=f,stderr=subprocess.STDOUT).returncode
