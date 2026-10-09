import pathlib,json,subprocess,os,time,re
out=pathlib.Path('review/test-audit/stage1-cohere-css-print'); originals=json.loads((out/'originals.json').read_text()); changes={m['id']:m for m in json.loads((out/'mutants.json').read_text())}; env=os.environ.copy(); env.pop('ADAMIC_MUTANT',None); env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u079/cache/baseline'
records=[]
for ident in ['M1','M2','M3','P1','P2']:
 m=changes[ident]; p=pathlib.Path(m['file']); assert p.read_text()==originals[m['file']]
 patch=(out/'diffs'/f'{ident}.diff').resolve(); check=subprocess.run(['git','apply','--check',str(patch)],capture_output=True,text=True); assert check.returncode==0,check.stderr
 subprocess.run(['git','apply',str(patch)],check=True)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run','^TestProduct_CSSPrinterSanitizedAndLowered$']; log=out/(ident+'-standalone-native-workspace.log'); start=time.time()
 try:
  with log.open('w') as f: code=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env).returncode
 finally: p.write_text(originals[m['file']])
 records.append(dict(id=ident,command=cmd,code=code,wall=time.time()-start,cache=env['ADAMIC_BUILD_CACHE_DIR'],log=str(log)))
 (out/'workspace-port-validations.json').write_text(json.dumps(records,indent=2))
 if code not in [0]: break
# Every input-only native product wrapper, including its platform-gated variant,
# belongs to the same family. Measure the complete family three times.
records=[]
for attempt in range(3):
 pattern='^TestProduct_CSSPrinter(SanitizedAndLowered|SemicolonMutant|IndentMutant|WidthMutant|DarwinLeaks)$'; cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run',pattern]; log=out/f'timing-native-family-all-{attempt+1}.log'; start=time.time()
 with log.open('w') as f: code=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env).returncode
 records.append(dict(test='TestProduct_CSSPrinterNative family',attempt=attempt+1,pattern=pattern,code=code,wall=time.time()-start,log=str(log)))
 (out/'complete-native-family-timings.json').write_text(json.dumps(records,indent=2))
