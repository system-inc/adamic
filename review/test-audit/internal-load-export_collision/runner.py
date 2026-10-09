import pathlib, json, subprocess, difflib, time, os
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-audit/internal-load-export_collision'
menu=[]
def add(id,file,old,new,kind='expression',operator='change constant or option'):
 p='internal/load/'+file; text=(root/p).read_text(); assert text.count(old)==1,(id,text.count(old)); menu.append(dict(id=id,file=p,line=text[:text.index(old)].count('\n')+1,old=old,new=new,kind=kind,operator=operator))
add('M01','load.go','NoUncheckedIndexedAccess:   core.TSTrue','NoUncheckedIndexedAccess:   core.TSFalse')
add('M02','load.go','ExactOptionalPropertyTypes: core.TSTrue','ExactOptionalPropertyTypes: core.TSFalse')
add('M03','load.go','Strict:                     core.TSTrue','Strict:                     core.TSFalse')
add('M04','load.go','core.ModuleDetectionKindForce','core.ModuleDetectionKindAuto')
add('M05','load.go','[]string{"lib.es2024.d.ts"}','[]string{"lib.es2024.d.ts", "lib.dom.d.ts"}')
add('M06','load.go','NoEmit:                     core.TSTrue,','NoEmit:                     core.TSTrue,\n\t\tNoImplicitReturns: core.TSTrue,',kind='option')
add('M07','load.go','len(diagnostics) > 0','len(diagnostics) > 1',operator='off-by-one bound')
add('M08','load.go','\t\tall = append(all, p.compiler.GetSemanticDiagnostics(ctx, nil)...)','',kind='drop',operator='drop statement')
add('M09','load.go','return line + 1, len(utf16.Encode([]rune(prefix))) + 1','return line + 2, len(utf16.Encode([]rune(prefix))) + 1',kind='statement',operator='off-by-one')
add('M10','export_collision.go','diagnostic.Code() != 2308','diagnostic.Code() != 2309')
add('M11','export_collision.go','arguments[1], strings.Trim(arguments[0], "\'\\\""), declaration.ModuleSpecifier.Text()','arguments[1], declaration.ModuleSpecifier.Text(), strings.Trim(arguments[0], "\'\\\"")',operator='swap two arguments')
add('M12','source_fs.go','strings.HasSuffix(path.AsString(), ".a.ts")','strings.HasSuffix(path.AsString(), ".b.ts")')
add('M13','load.go','\tfor name, source := range overlay {\n\t\tnormalizedOverlay[currentDirectory.ResolveFile(name)] = source\n\t}','',kind='drop',operator='drop whole loop to avoid unused variables')
add('M14','node_library.go','pin.Version != NodeTypesVersion','pin.Version == NodeTypesVersion',operator='flip condition')
add('M15','load.go','usesNodeModules(program)','!usesNodeModules(program)',operator='flip condition')
add('M16','node_library.go','func nodePrelude() string {','func nodePrelude() string {\n\treturn prelude\n',kind='entry',operator='return early')
add('M17','declarations.go','\t\t\tnode.ForEachChild(visit)','',kind='drop',operator='drop statement')
add('M18','declarations.go','node.Kind == ast.KindTypeAliasDeclaration','node.Kind != ast.KindTypeAliasDeclaration',operator='flip condition')
add('M19','load.go','fs.FS.FileExists(fileName.AppendSuffix(".ts"))','!fs.FS.FileExists(fileName.AppendSuffix(".ts"))',operator='flip condition')
add('M20','node_library.go','strings.HasSuffix(file.FileName().AsString(), ".d.ts")','strings.HasSuffix(file.FileName().AsString(), ".ts")')
add('P01','load.go','func Load(paths []string) (*Program, error) {','func Load(paths []string) (*Program, error) {\n\treturn nil, nil\n',kind='entry',operator='empty-answer probe')
add('P02','load.go','func LoadOverlay(paths []string, overlay map[string]string) (*Program, error) {','func LoadOverlay(paths []string, overlay map[string]string) (*Program, error) {\n\treturn nil, nil\n',kind='entry',operator='empty-answer probe')
add('P03','node_library.go','func nodeTypesIndex(directory string) (string, error) {','func nodeTypesIndex(directory string) (string, error) {\n\treturn "", nil\n',kind='entry',operator='empty-answer probe')
base={m['file']:(root/m['file']).read_text() for m in menu}
(out/'menu.json').write_text(json.dumps(menu,indent=2))
functions={f:[s for s in text.splitlines() if s.startswith('func ')] for f,text in base.items()}
for f in ['internal/load/regexp_library.go']:
 functions[f]=[s for s in (root/f).read_text().splitlines() if s.startswith('func ')]
(out/'code-and-oracle.md').write_text('CODE UNDER TEST: Adamic internal/load wrappers, configuration, declarations, diagnostics, source filesystem and library selection. Upstream checker and installed Node declarations are not mutated.\nORACLE: hand-written acceptance, diagnostic and declaration assertions; TypeScript diagnostic authority, official Node type signatures, and Adamic docs/0.1.md contract. No scoped row runs an external comparator.\nFunctions available on the reached paths (read-only FS methods are interface capabilities; not every invocation was dynamically traced):\n'+json.dumps(functions,indent=2)+'\nFixed menu was declared before inspecting mutant failures. M06 changes the default-false NoImplicitReturns option to true. M13 drops the whole loop. M20 may be equivalent because every .d.ts also ends in .ts.\n')
# Every standalone diff is validated from the same base, then restored.
validation=[]
for m in menu:
 text=base[m['file']].replace(m['old'],m['new'])
 if m['kind']=='entry':
  startpos=base[m['file']].index(m['old']); endpos=base[m['file']].index('\n}',startpos)+2
  text=base[m['file']][:startpos]+m['new'].rstrip()+'\n}'+base[m['file']][endpos:]
 if m['id']=='P03':
  for imp in ['encoding/json','fmt','os','path/filepath']: text=text.replace('\t"'+imp+'"\n','')
 (root/m['file']).write_text(text)
 diff=''.join(difflib.unified_diff(base[m['file']].splitlines(True),text.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file']))
 (out/(m['id']+'.diff')).write_text(diff)
 start=time.monotonic()
 with (out/(m['id']+'-vet.log')).open('w') as log:
  rc=subprocess.run(['go','vet','./internal/load/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=90).returncode
 validation.append(dict(id=m['id'],exit=rc,seconds=time.monotonic()-start))
 (root/m['file']).write_text(base[m['file']])
 if rc: raise RuntimeError(m['id']+' vet failed')
(out/'validation.json').write_text(json.dumps(validation,indent=2))
# Switch only instruments the already specified one-change mutants.
for file,text in base.items():
 for m in [m for m in menu if m['file']==file]:
  old,new=m['old'],m['new']; id=m['id']
  if m['kind']=='entry':
   ret=new[len(old):].strip(); replacement=old+'\n\tif auditMutant("'+id+'") { '+ret+' }\n'
  elif m['kind']=='drop': replacement='if !auditMutant("'+id+'") {\n'+old+'\n}'
  elif m['kind']=='statement': replacement='if auditMutant("'+id+'") { '+new+' }; '+old
  elif m['kind']=='option': replacement=old+'\n\t\tNoImplicitReturns: auditChoose("'+id+'", core.TSFalse, core.TSTrue),'
  elif id in ['M01','M02','M03']:
   prefix,orig=old.split(':',1); mutant=new.split(':',1)[1]; replacement=prefix+': auditChoose("'+id+'", '+orig.strip()+', '+mutant.strip()+')'
  elif id=='M11':
   # Swap argument expressions without touching the formatting oracle.
   replacement='arguments[1], auditChoose("M11", strings.Trim(arguments[0], "\'\\\""), declaration.ModuleSpecifier.Text()), auditChoose("M11", declaration.ModuleSpecifier.Text(), strings.Trim(arguments[0], "\'\\\""))'
  else: replacement='auditChoose("'+id+'", '+old+', '+new+')'
  text=text.replace(old,replacement)
 (root/file).write_text(text)
(root/'internal/load/u024_switch.go').write_text('package load\nimport "os"\nfunc auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\nfunc auditChoose[T any](id string, normal, mutant T) T { if auditMutant(id) { return mutant }; return normal }\n')
subprocess.run(['gofmt','-w',*[str(root/f) for f in base],str(root/'internal/load/u024_switch.go')],check=True)
start=time.monotonic()
with (out/'switch-build.log').open('w') as log:
 subprocess.run(['go','test','-c','-o','/tmp/u024-load.test','./internal/load/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=90,check=True)
(out/'switch-build-seconds.json').write_text(json.dumps(time.monotonic()-start))
for m in menu:
 env=os.environ.copy(); env['ADAMIC_MUTANT']=m['id']; start=time.monotonic()
 with (out/(m['id']+'.log')).open('w') as log:
  rc=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/load/','-run','.'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
 print(m['id'],rc,round(time.monotonic()-start,3),flush=True)
 if 'panic:' in (out/(m['id']+'.log')).read_text():
  rows=[l for l in pathlib.Path('/tmp/u024-list-retry.log').read_text().splitlines() if l.startswith('Test')]
  for row in rows:
   with (out/(m['id']+'-'+row+'.log')).open('w') as log:
    subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/load/','-run','^'+row+'$'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
# Restore production source after all runs.
for file,text in base.items(): (root/file).write_text(text)
(root/'internal/load/u024_switch.go').unlink()
