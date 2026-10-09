#!/usr/bin/env python3
"""Independent Go overlays; each mutant must fail with its intended diagnostic."""
import json, os, pathlib, subprocess
root=pathlib.Path(__file__).resolve().parents[3]
source=(root/'internal/native/outline.go').read_text()
out=pathlib.Path(os.environ.get('OUTLINE_MUTANTS','/tmp/adamic-constant-outline-mutants'));out.mkdir(parents=True,exist_ok=True)
mutants=[
 ('module-order','for partIndex, indexes := range modules {','for order := range modules {\npartIndex := len(modules)-1-order\nindexes := modules[partIndex]', './internal/oracle','^TestNativeAgreesWithNode/internal/oracle/testdata/import_cycles/order/main', 'stdout differs'),
 ('unbounded-chunk','const initializerChunkStatements = 128','const initializerChunkStatements = 100000','./internal/native','^TestInitializerChunksKeepStorageAndOrder$', 'large initialization wrapper or chunk'),
 ('inline-again','const initializerAttribute = " __attribute__((noinline))"','const initializerAttribute = ""','./internal/native','^TestInitializerChunksKeepStorageAndOrder$', 'large initializer was not outlined'),
 ('omit-root-cleanup','cleanup := e.out.String()','cleanup := ""','./internal/native','^TestInitializerChunksKeepStorageAndOrder$', 'LeakSanitizer: detected memory leaks'),
]
env=os.environ.copy();env.update(ADAMIC_GATE_UNCACHED='1',ADAMIC_NATIVE_SPLIT='1',ADAMIC_NATIVE_JOBS='5')
for name,old,new,package,test,want in mutants:
 assert source.count(old)==1,(name,source.count(old))
 changed=out/(name+'.go');changed.write_text(source.replace(old,new))
 overlay=out/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(root/'internal/native/outline.go'):str(changed)}}))
 command=['go','test','-overlay='+str(overlay),package,'-run',test,'-v','-count=1','-timeout=10m']
 log=out/(name+'.log')
 with log.open('wb') as handle:result=subprocess.run(command,cwd=root,env=env,stdout=handle,stderr=subprocess.STDOUT)
 text=log.read_text();assert result.returncode!=0 and want in text,(name,result.returncode,text[-4000:])
 assert 'clang failed' not in text and 'compiling unit' not in text,(name,'compiler failure masked intended check')
 print(name,'caught:',want,flush=True)

# Prove the two narrow splitter integration hooks are needed independently.
selected=root/'internal/native/units_stable.go'
source=selected.read_text()
for name,old,new,want in [
 ('outlined-chunk-placement', 'if strings.HasPrefix(d.name, "adamic_initialize_chunk_") {', 'if false && strings.HasPrefix(d.name, "adamic_initialize_chunk_") {', 'large initializer stayed in its module unit'),
 ('outlined-storage-owner', 'if helper.initializationOwner != "" {', 'if false && helper.initializationOwner != "" {', 'initializer global table owned by module_'),
 ('outlined-global-owner', 'if item.source && strings.HasPrefix(name, "adamic_initialize_") {', 'if false && item.source && strings.HasPrefix(name, "adamic_initialize_") {', 'initializer global table owned by main.c'),
 ('outlined-noinline-prototype', 'if !initializer || unannotated != declaration {', 'if initializer { old.declaration = declaration }\nif !initializer || unannotated != declaration {', 'module initializer lost noinline prototype'),
]:
 assert source.count(old)==1,(name,source.count(old))
 changed=out/(name+'.go');changed.write_text(source.replace(old,new))
 overlay=out/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(selected):str(changed)}}))
 command=['go','test','-overlay='+str(overlay),'./internal/native','-run','^TestInitializerUnitOwnership$','-v','-count=1','-timeout=10m']
 log=out/(name+'.log')
 with log.open('wb') as handle:result=subprocess.run(command,cwd=root,env=env,stdout=handle,stderr=subprocess.STDOUT)
 text=log.read_text();assert result.returncode!=0 and want in text,(name,result.returncode,text[-4000:])
 assert 'clang failed' not in text,(name,'compiler failure masked intended check')
 print(name,'caught:',want,flush=True)
