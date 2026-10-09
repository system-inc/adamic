import pathlib,subprocess,time,json,difflib,statistics,os
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/internal-load-export_collision'
# Probe remaining direct public answer entries, apart from Load preparation.
probes=[('P04','internal/load/declarations.go','func (p *Program) Declarations(ctx context.Context) []Declaration {','return nil'),('P05','internal/load/node_library.go','func IsNodeLibrary(file *ast.SourceFile) bool {','return false')]
base={f:(root/f).read_text() for _,f,_,_ in probes}
for id,f,header,ret in probes:
 s=base[f];start=s.index(header);end=s.index('\n}',start)+2;new=s[:start]+header+'\n\t'+ret+'\n}'+s[end:]
 (out/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 (root/f).write_text(new)
 with (out/(id+'-vet.log')).open('w') as log: subprocess.run(['go','vet','./internal/load/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=90)
 (root/f).write_text(s)
for id,f,header,ret in probes:
 s=(root/f).read_text();s=s.replace(header,header+'\n\tif os.Getenv("ADAMIC_MUTANT") == "'+id+'" { '+ret+' }')
 if f.endswith('declarations.go'):s=s.replace('"context"','"context"\n"os"')
 (root/f).write_text(s)
start=time.monotonic()
for id,_,_,_ in probes:
 env=os.environ.copy();env['ADAMIC_MUTANT']=id
 with (out/(id+'.log')).open('w') as log:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/load/','-run','.'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
for f,s in base.items():(root/f).write_text(s)
(out/'additional-probes-seconds.json').write_text(json.dumps(time.monotonic()-start))
rows=[l for l in pathlib.Path('/tmp/u024-list-retry.log').read_text().splitlines() if l.startswith('Test')][:15]
timings={}
for row in rows:
 vals=[]
 for n in range(3):
  logpath=out/(row+'-timing-'+str(n+1)+'.log')
  with logpath.open('w') as log: rc=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/load/','-run','^'+row+'$'],cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
  assert rc==0,(row,n)
  for line in logpath.read_text().splitlines():
   try:e=json.loads(line)
   except:continue
   if e.get('Action')=='pass' and not e.get('Test'):vals.append(e['Elapsed'])
 timings[row]={'runs':vals,'median':statistics.median(vals),'source':'test binary ok line, package Elapsed'}
 print(row,timings[row],flush=True)
(out/'timings.json').write_text(json.dumps(timings,indent=2))
