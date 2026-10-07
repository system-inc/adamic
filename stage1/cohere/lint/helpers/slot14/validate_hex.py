#!/usr/bin/env python3
"""Actual private Go helper, consumer suites, byte controls and clean mutants."""
import json, os, random, shutil, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[5]
owned=Path(__file__).resolve().parent
cohere=(root/'cohere').resolve()
scratch=Path(tempfile.mkdtemp(prefix='slot14-fixed-hex-'))
evidence=owned/'evidence';evidence.mkdir(exist_ok=True)
log=(evidence/'hex.log').open('w',buffering=1)
def note(text):print(text,file=log,flush=True)
def run(args,cwd=root,env=None,output=None):
 out=output or scratch/'out';err=scratch/'err'
 with out.open('wb') as stdout,err.open('wb') as stderr:
  p=subprocess.run(list(map(str,args)),cwd=cwd,env=env,stdout=stdout,stderr=stderr)
 if p.returncode or err.stat().st_size:
  note('FAILED '+repr(list(map(str,args)))+' exit='+str(p.returncode))
  note(err.read_text());note(out.read_text()[:6000]);raise RuntimeError('command failed')
 return out.read_bytes()
try:
 note('scratch='+str(scratch))
 original=cohere/'internal/lint/ecmascript/regexp/escape.go'
 changed=scratch/'escape.go'
 source=original.read_text();assert source.count('func decodeFixedHex(')==1
 changed.write_text(source.replace('func decodeFixedHex(','func slot14OriginalFixedHex('))
 capture_overlay=scratch/'capture-overlay.json'
 capture_overlay.write_text(json.dumps({'Replace':{str(original):str(changed),str(original.parent/'slot14_fixed_hex_capture.go'):str(owned/'hex_capture.go.txt')}}))
 consumers=[
  ('@next/next/no-html-link-for-pages','next','TestNoHtmlLinkForPages'),
  ('@typescript-eslint/no-empty-object-type','typescript','TestNoEmptyObjectType'),
  ('no-restricted-exports','core','TestNoRestrictedExports'),
  ('no-restricted-imports','core','TestNoRestrictedImports')]
 rows=[];coverage={}
 for name,package,pattern in consumers:
  path=scratch/(package+'-'+pattern+'.jsonl')
  env=dict(os.environ,SLOT14_HEX_CAPTURE=str(path))
  run(['go','test','-overlay='+str(capture_overlay),'./internal/lint/rules/'+package,'-run','^'+pattern,'-count=1','-v','-timeout=10m'],cwd=cohere,env=env,output=evidence/('hex-'+package+'-'+pattern+'.log'))
  captured=[json.loads(line) for line in path.read_text().splitlines()] if path.exists() else []
  rows.extend(captured);coverage[name]={'original_tests':'PASS','actual_private_calls':len(captured)}
  note(name+': original Go assertions PASS; actual private calls='+str(len(captured)))
 # Target real rule decoders with fixed-hex patterns absent from upstream rows.
 replacements=json.loads(capture_overlay.read_text())['Replace']
 replacements[str(cohere/'internal/lint/rules/core/slot14_hex_test.go')]=str(owned/'hex_core_test.go.txt')
 replacements[str(cohere/'internal/lint/rules/typescript/slot14_hex_test.go')]=str(owned/'hex_typescript_test.go.txt')
 replacements[str(cohere/'internal/lint/rules/next/slot14_hex_test.go')]=str(owned/'hex_next_test.go.txt')
 targeted_overlay=scratch/'targeted-overlay.json'
 targeted_overlay.write_text(json.dumps({'Replace':replacements}))
 for name,package,pattern in [
  ('@next/next/no-html-link-for-pages','next','TestSlot14HexRoutes'),
  ('no-restricted-exports','core','TestSlot14HexExports'),
  ('no-restricted-imports','core','TestSlot14HexImports'),
  ('@typescript-eslint/no-empty-object-type','typescript','TestSlot14HexEmptyObject')]:
  path=scratch/(pattern+'.jsonl')
  env=dict(os.environ,SLOT14_HEX_CAPTURE=str(path))
  run(['go','test','-overlay='+str(targeted_overlay),'./internal/lint/rules/'+package,'-run','^'+pattern+'$','-count=1','-v','-timeout=10m'],cwd=cohere,env=env,output=evidence/('hex-'+pattern+'.log'))
  captured=[json.loads(line) for line in path.read_text().splitlines()] if path.exists() else []
  assert captured,name+' targeted helper invocation absent'
  rows.extend(captured)
  coverage[name]['targeted_private_calls']=len(captured)
  note(name+': targeted real-rule assertions PASS; actual private calls='+str(len(captured)))
 # Caller-independent controls include byte indexes, uint32 conversion, all
 # byte pairs, every byte as a fallback introducer, and malformed UTF-8.
 def add(source,index=0,size=1,digits=2,unicode=False):
  if isinstance(source,str):source=list(source.encode())
  rows.append({'Source':source,'Index':index,'Size':size,'Digits':digits,'Unicode':unicode,'InClass':False,'Named':False,'Groups':0})
 for first in range(256):
  for second in range(256):add([120,first,second])
  for unicode in [False,True]:add([first],unicode=unicode)
 for text in ['x','x0','x00','xFF','x+1','x-1','x0x','x_0','x 0','xé','u0000','uD800','uFFFF','u+001','u000_','uffffffff','u80000000','u100000000','u000000000','u']:
  for digits in [0,1,2,4,8,9,16]:
   for unicode in [False,True]:add(text,digits=digits,unicode=unicode)
 for prefix in ['é','😀','中','\uFEFF']:
  for suffix in ['x00','u0041','xG0','u12','x']:
   data=prefix.encode()+suffix.encode()
   for unicode in [False,True]:
    add(list(data),index=len(prefix.encode()),digits=2 if suffix[0]=='x' else 4,unicode=unicode)
    add(list(data),size=len(prefix.encode()),unicode=unicode)
 for raw in [[0xC0,0x80],[0xE0,0x80,0x80],[0xED,0xA0,0x80],[0xF4,0x90,0x80,0x80],[0xF0,0x9F,0x98,0x80],[0xF4,0x8F,0xBF,0xBF],[0xC2],[0xE2,0x82],[]]:
  for unicode in [False,True]:add(raw,unicode=unicode)
 rng=random.Random(14)
 for _ in range(2500):
  raw=[rng.randrange(256) for _ in range(rng.randrange(0,20))]
  add(raw,index=rng.randrange(len(raw)+1),size=rng.randrange(0,5),digits=rng.choice([0,1,2,4,8,9,16]),unicode=bool(rng.randrange(2)))
 unique={json.dumps(row,sort_keys=True):row for row in rows}
 rows=list(unique.values());inputs=scratch/'inputs.json';inputs.write_text(json.dumps(rows))
 (evidence/'hex-coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
 note('unique controls and consumer inputs='+str(len(rows)))
 oracle_overlay=scratch/'oracle-overlay.json'
 oracle_overlay.write_text(json.dumps({'Replace':{str(original.parent/'slot14_hex_oracle_test.go'):str(owned/'hex_oracle_test.go.txt')}}))
 expected=scratch/'expected'
 env=dict(os.environ,SLOT14_HEX_INPUTS=str(inputs),SLOT14_HEX_OUTPUT=str(expected))
 run(['go','test','-overlay='+str(oracle_overlay),'./internal/lint/ecmascript/regexp','-run','^TestSlot14FixedHex$','-count=1','-v','-timeout=10m'],cwd=cohere,env=env,output=evidence/'hex-oracle.log')
 want=expected.read_bytes()
 build_overlay=scratch/'build-overlay.json'
 build_overlay.write_text(json.dumps({'Replace':{str(root/'slot14_fixed_hex_build.go'):str(owned/'build.go.txt')}}))
 builder=scratch/'builder'
 run(['go','build','-overlay='+str(build_overlay),'-o',builder,root/'slot14_fixed_hex_build.go'])
 binary=scratch/'native';emitted=scratch/'emitted.mjs'
 run([builder,owned/'hex_runner.a',binary,emitted,'sanitized'])
 node=['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs']
 for side,args in [('Node',node+[owned/'hex_runner.a',inputs]),('native',[binary,inputs]),('emitted JavaScript',node+[emitted,inputs])]:
  got=run(args)
  if got!=want:
   a=got.splitlines();b=want.splitlines()
   for index in range(max(len(a),len(b))):
    left=a[index] if index<len(a) else b'<EOF>';right=b[index] if index<len(b) else b'<EOF>'
    if left!=right:
     note('DIFFERENCE '+side+' case='+str(index)+' '+repr(rows[index])+' '+repr(left)+' != '+repr(right));break
   raise RuntimeError('byte comparison failed')
  note(side+': '+str(len(rows))+' inputs / '+str(len(want))+' bytes identical to Go')
 variant=scratch/'mutant';variant.mkdir()
 shutil.copyfile(owned/'hex_runner.a',variant/'hex_runner.a')
 # Keep the reader as a test dependency in scratch, not a shared repository edit.
 shutil.copyfile(owned.parent/'options_json.ts',scratch/'options_json.ts')
 spec=json.loads((owned/'hex_mutant.json').read_text())
 text=(owned/spec['file']).read_text();assert text.count(spec['from'])==1
 (variant/spec['file']).write_text(text.replace(spec['from'],spec['to']))
 mn=scratch/'mutant-native';mj=scratch/'mutant.mjs'
 run([builder,variant/'hex_runner.a',mn,mj,'sanitized'])
 for side,args in [('Node',node+[variant/'hex_runner.a',inputs]),('native',[mn,inputs]),('emitted JavaScript',node+[mj,inputs])]:
  assert run(args)!=want,side+' mutant survived'
  note(side+': compiling exit-zero clean-stderr width mutant caught only by Go byte comparison')
 note('PASS actual private Go fixed-hex behavior, all four original consumer suites, byte controls and semantic mutant')
except Exception as error:
 note('FAIL '+repr(error));raise
finally:log.close()
