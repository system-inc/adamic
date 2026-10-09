import pathlib,subprocess,time,json,difflib,os,sys
repo=pathlib.Path('/workspace/adamic');p=repo/'review/test-audit/stage3-census-latent-statecopy';tmp=pathlib.Path('/tmp/u161');file='stage3/census/latent/statecopy/main.go';source=(repo/file).read_text()
menu=[
('M1','panic("state copy: mutable named container " + t.Name + " requires an explicit copier")','return value','return early in named-container rejection'),
('M2','panic("state copy: foreign or unrecognized pointer " + printed(t) + " requires an explicit copier")','return value','return early in foreign-pointer rejection'),
('M3','if strings.HasSuffix(path, "_test.go") {','if !strings.HasSuffix(path, "_test.go") {','flip input-file exclusion condition'),
('M4','os.WriteFile(os.Args[3], text, 0600)','os.WriteFile(os.Args[3], text, 0644)','change output permission constant'),
('M5','return b.String()','return ""','change printed type result to empty string'),
('M6','if len(os.Args) != 4 {','if len(os.Args) == 4 {','flip argument-count condition')]
manifest=[]
for id,old,new,kind in menu:
 assert source.count(old)==1,(id,source.count(old));manifest.append({'id':id,'file':file,'line':source[:source.index(old)].count('\n')+1,'old':old,'change':new,'menu':kind})
(p/'mutant-menu.json').write_text(json.dumps(manifest,indent=2)+'\n')
(p/'code-under-test.md').write_text('CODE UNDER TEST: statecopy generator, not lowering or the Go compiler.\nORACLE: self-written named-container and foreign-pointer diagnostics, nonzero subprocess result, and no output publication.\n\nComplete package function inventory reached by the rows:\n- main (entry): argument admission, input discovery, AST type inventory, struct traversal, output construction and publication.\n- generator.copy: Ident and StarExpr rejection branches reached by the fixtures.\n- printed: Go AST formatting reached by the foreign-pointer path.\nOther branches in generator.copy are not executed by the two fixtures.\n\nNo tests, fixtures or oracle expectations are mutated. Six production changes are frozen here before outcome checks, spread across all three reached functions. P1 is a separate empty main-entry probe. The tests remain independent rows because they assert different rejection guards and distinct diagnostics; there is no shared fixture-check helper.\n')
if '--plan' in sys.argv:print('six-mutant menu frozen');raise SystemExit()
builds=[];runs=[]
try:
 for id,old,new,kind in menu:
  modified=source.replace(old,new,1);diff=''.join(difflib.unified_diff(source.splitlines(True),modified.splitlines(True),fromfile='a/'+file,tofile='b/'+file));(p/(id+'.diff')).write_text(diff)
  subprocess.run(['git','apply','--check',str(p/(id+'.diff'))],cwd=repo,check=True);(repo/file).write_text(modified);cmd=['go','vet','./stage3/census/latent/statecopy/'];start=time.monotonic()
  with (p/(id+'-vet.log')).open('w') as f:r=subprocess.run(cmd,cwd=repo,stdout=f,stderr=subprocess.STDOUT)
  builds.append({'id':id,'command':cmd,'seconds':time.monotonic()-start,'exit':r.returncode});(repo/file).write_text(source)
  if r.returncode:raise RuntimeError('invalid standalone mutant '+id)
 # Single source switch; the Go run subprocess inherits the selected environment.
 switched=source
 for id,old,new,kind in menu:
  if id in ['M1','M2']:replacement='if os.Getenv("ADAMIC_MUTANT") == "'+id+'" { return value }; '+old
  elif id=='M3':replacement='if strings.HasSuffix(path, "_test.go") != (os.Getenv("ADAMIC_MUTANT") == "M3") {'
  elif id=='M4':replacement='os.WriteFile(os.Args[3], text, func() os.FileMode { if os.Getenv("ADAMIC_MUTANT") == "M4" { return 0644 }; return 0600 }())'
  elif id=='M5':replacement='if os.Getenv("ADAMIC_MUTANT") == "M5" { return "" }; '+old
  else:replacement='if (len(os.Args) != 4) != (os.Getenv("ADAMIC_MUTANT") == "M6") {'
  switched=switched.replace(old,replacement,1)
 switched=switched.replace('func main() {','func main() {\n if os.Getenv("ADAMIC_MUTANT") == "P1" { return }',1)
 probe=source.replace('func main() {','func main() {\n if len(os.Args) >= 0 { return }',1)
 (p/'P1.diff').write_text(''.join(difflib.unified_diff(source.splitlines(True),probe.splitlines(True),fromfile='a/'+file,tofile='b/'+file)));(repo/file).write_text(probe)
 with (p/'P1-vet.log').open('w') as f:subprocess.run(['go','vet','./stage3/census/latent/statecopy/'],cwd=repo,stdout=f,stderr=subprocess.STDOUT,check=True)
 (repo/file).write_text(switched)
 with (p/'scratch-switch.diff').open('w') as f:subprocess.run(['git','diff','--',file],cwd=repo,stdout=f,check=True)
 cmd=['go','test','-c','-o',str(tmp/'statecopy.test'),'./stage3/census/latent/statecopy/'];start=time.monotonic()
 with (p/'switch-build.log').open('w') as f:subprocess.run(cmd,cwd=repo,stdout=f,stderr=subprocess.STDOUT,check=True)
 builds.append({'id':'switch-test-binary','command':cmd,'seconds':time.monotonic()-start,'exit':0})
 for id in [m[0] for m in menu]+['P1']:
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage3/census/latent/statecopy/','-run','.'];start=time.monotonic()
  with (p/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=repo,env=dict(os.environ,ADAMIC_MUTANT=id),stdout=f,stderr=subprocess.STDOUT)
  runs.append({'id':id,'command':cmd,'environment':{'ADAMIC_MUTANT':id},'seconds':time.monotonic()-start,'exit':r.returncode});(p/'runs.json').write_text(json.dumps(runs,indent=2)+'\n')
 # Real changed-behavior witness for publication-mode survivors.
 witness=tmp/'witness';(witness/'internal/lower').mkdir(parents=True,exist_ok=True);fixture=witness/'internal/lower/lower.go';fixture.write_text('package lower\ntype lowering struct { state int }\n');(p/'survivor-input.go.txt').write_text(fixture.read_text());observations=[]
 for id in ['', 'M4']:
  output=tmp/('witness-'+(id or 'baseline')+'.go');cmd=['go','run','main.go',str(witness),str(fixture),str(output)]
  with (p/('survivor-'+(id or 'baseline')+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=repo/'stage3/census/latent/statecopy',env=dict(os.environ,ADAMIC_MUTANT=id,GOWORK='off'),stdout=f,stderr=subprocess.STDOUT)
  observations.append({'id':id or 'baseline','command':cmd,'exit':r.returncode,'mode':oct(output.stat().st_mode & 0o777),'output':output.read_text()});assert r.returncode==0
 (p/'survivor-witness.json').write_text(json.dumps(observations,indent=2)+'\n')
 # Both generated positive products compile alongside their declared input.
 for id in ['baseline','M4']:
  with (p/('survivor-'+id+'-compile.log')).open('w') as f:subprocess.run(['go','tool','compile','-o',str(tmp/('witness-'+id+'.o')),str(fixture),str(tmp/('witness-'+id+'.go'))],cwd=repo,stdout=f,stderr=subprocess.STDOUT,check=True)
finally:
 (repo/file).write_text(source);(p/'build-times.json').write_text(json.dumps(builds,indent=2)+'\n')
print('matrix, probe and survivor witness complete')
