import pathlib,subprocess,time,json,difflib,os
repo=pathlib.Path('/workspace/adamic');r=repo/'review/test-audit/stage3-fixtures-fixtures';p=repo/'stage3/fixtures/fixtures_test.go';results=[]
def experiment(mid,file,a,b,pattern,args=[]):
 q=repo/file;s=q.read_text();assert s.count(a)==1;v=s.replace(a,b);line=s[:s.index(a)].count('\n')+1;(r/(mid+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),v.splitlines(True),fromfile='a/'+file,tofile='b/'+file)));q.write_text(v)
 try:
  if file.endswith('.go'):
   with (r/'logs'/f'{mid}-vet.log').open('w') as out:rc=subprocess.run(['go','vet','./stage3/fixtures/'],stdout=out,stderr=subprocess.STDOUT).returncode
   if rc:raise RuntimeError(mid+' vet')
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage3/fixtures/','-run',pattern]+args;t=time.monotonic()
  with (r/'logs'/f'{mid}.log').open('w') as out:rc=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT).returncode
  results.append(dict(id=mid,file=file,line=line,old=a,new=b,command=cmd,wall=time.monotonic()-t,rc=rc));(r/'setup-runs.json').write_text(json.dumps(results,indent=2))
 finally:q.write_text(s)
file='stage3/fixtures/fixtures_test.go'
experiment('Kpaths',file,'filepath.Ext(name) == ".a"','filepath.Ext(name) == ".ts"','^TestFixturePaths$')
# Change constructed manifest identities, preserving every oracle assertion.
experiment('Kmanifest',file,'name := filepath.Base(filepath.Dir(status)) + "/" + entry.File\n\t\t\tif seen[name]','name := "" + ""\n\t\t\tif seen[name]','^TestFixtureShardManifest$')
# A wrong assignment is not asserted by this row. Record this construction survivor separately.
experiment('Kassignment',file,'return int(binary.BigEndian.Uint64(digest[:8]) % uint64(count))','_ = digest; _ = binary.BigEndian; return 0','^TestFixtureShardManifest$', ['-args','-test.v'])
# Flip the preparation product integrity condition; no Node or behavior oracle is edited.
experiment('Kprepare','stage3/fixtures/build-hook.py','if hashlib.sha256(product.read_bytes()).hexdigest() != product.name:','if hashlib.sha256(product.read_bytes()).hexdigest() == product.name:','^TestPrepareFixtureOracleHook$')
# Empty helper probes assess construction, not lowering.
experiment('Ppaths',file,'return fs.ValidPath(name) && !strings.ContainsAny(name, "\\\\:") && filepath.Ext(name) == ".a"','_ = fs.ValidPath; return false','^TestFixturePaths$')
original=p.read_text();a=original[original.index('func fixtureOracleHook('):original.index('// Not parallel: prepare')].rstrip();experiment('Pprepare',file,a,'func fixtureOracleHook(t *testing.T, repository string) string { return "" }','^TestPrepareFixtureOracleHook$')
