import pathlib,json,subprocess,hashlib,re
ROOT=pathlib.Path.cwd();OUT=ROOT/'review/compiler/loom-r-checked-writes-ac';rows=json.loads((OUT/'status.json').read_text());changes=json.loads((OUT/'headers.json').read_text())
refs={}; candidates=[p for p in ROOT.rglob('*') if p.is_file() and (p.name.endswith('_test.go') or 'testdata' in p.parts or p.suffix=='.md') and not any(x in p.parts for x in ('cohere','.git','node_modules'))]
for row in changes:
 p=row['path'];matches=[]
 for q in candidates:
  if q.stat().st_size>2000000:continue
  for i,line in enumerate(q.read_text(errors='replace').splitlines(),1):
   if p in line or (pathlib.Path(p).name in line and len(pathlib.Path(p).name)>8):matches.append(dict(file=str(q.relative_to(ROOT)),line=i,text=line))
 refs[p]=matches
 old=subprocess.check_output(['git','show','5aa205ec:'+p]);new=(ROOT/p).read_bytes()
 def body(data):return data.split(b'\n',1)[1] if data.startswith(b'// a-check:') else data
 assert body(old)==body(new),p
(OUT/'path-references.json').write_text(json.dumps(refs,indent=2)+'\n')
# Compare the actual emitted C for all successful .a programs and all 11 TypeScript adapters.
checks=[]
for i,row in enumerate(rows):
 if row['status']=='checked':
  a=(OUT/f'compile-{i:02d}.stdout.log').read_bytes();b=(OUT/f'final-{i:02d}.stdout.log').read_bytes();assert a==b,row['path'];checks.append(dict(path=row['path'],kind='direct .a',c_sha256=hashlib.sha256(b).hexdigest(),identical=True))
 if '/checked-runtime-fields_refused/' in row['path'] or '/checked-write-messages_refused/' in row['path']:
  before=subprocess.check_output(['git','show','5aa205ec:'+row['path']]);after=(ROOT/row['path']).read_bytes();scratch=OUT/'adapter-source';scratch.mkdir(exist_ok=True);p=scratch/(pathlib.Path(row['path']).stem+'.ts');emitted=[]
  for label,data in [('before',before),('after',after)]:
   p.write_bytes(data);r=subprocess.run(['timeout','30',str(OUT/'adamic-acheck'),'c',str(p)],stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=35);assert r.returncode==0,(p,r.stderr.decode());(OUT/(p.stem+'-'+label+'.c.log')).write_bytes(r.stdout);emitted.append(r.stdout)
  assert emitted[0]==emitted[1],row['path'];checks.append(dict(path=row['path'],kind='TypeScript adapter',c_sha256=hashlib.sha256(emitted[1]).hexdigest(),identical=True));p.unlink()
(OUT/'emitted-c-identity.json').write_text(json.dumps(checks,indent=2)+'\n')
assert subprocess.check_output(['git','show','5aa205ec:internal/oracle/counts.md'])==(ROOT/'internal/oracle/counts.md').read_bytes()
print('PASS 23 bodies byte-identical; 5 direct .a outputs and 11 TypeScript adapter C outputs byte-identical; counts.md unchanged')
print('Only line shift:',[(x['path'],x['line_shift']) for x in changes if x['line_shift']])
print('Line-shift path references:',refs['stage3/checked-writes/proven-generic.a'])
