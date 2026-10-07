#!/usr/bin/env python3
"""Actual private Go helper, consumer suites, byte controls and clean mutants."""
import json, os, random, shutil, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[5]
owned=Path(__file__).resolve().parent
cohere=(root/'cohere').resolve()
scratch=Path(tempfile.mkdtemp(prefix='slot14-unicode-escape-'))
evidence=owned/'evidence';evidence.mkdir(exist_ok=True)
log=(evidence/'unicode.log').open('w',buffering=1)
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
 source=original.read_text();assert source.count('func decodeUnicodeEscape(')==1
 changed.write_text(source.replace('func decodeUnicodeEscape(','func slot14OriginalUnicodeEscape('))
 capture_overlay=scratch/'capture-overlay.json'
 capture_overlay.write_text(json.dumps({'Replace':{str(original):str(changed),str(original.parent/'slot14_fixed_unicode_capture.go'):str(owned/'unicode_capture.go.txt')}}))
 consumers=[
  ('@next/next/no-html-link-for-pages','next','TestNoHtmlLinkForPages'),
  ('@typescript-eslint/no-empty-object-type','typescript','TestNoEmptyObjectType'),
  ('no-restricted-exports','core','TestNoRestrictedExports'),
  ('no-restricted-imports','core','TestNoRestrictedImports')]
 rows=[];coverage={}
 for name,package,pattern in consumers:
  path=scratch/(package+'-'+pattern+'.jsonl')
  env=dict(os.environ,SLOT14_UNICODE_CAPTURE=str(path))
  run(['go','test','-overlay='+str(capture_overlay),'./internal/lint/rules/'+package,'-run','^'+pattern,'-count=1','-v','-timeout=10m'],cwd=cohere,env=env,output=evidence/('unicode-'+package+'-'+pattern+'.log'))
  captured=[json.loads(line) for line in path.read_text().splitlines()] if path.exists() else []
  rows.extend(captured);coverage[name]={'original_tests':'PASS','actual_private_calls':len(captured)}
  note(name+': original Go assertions PASS; actual private calls='+str(len(captured)))
 # Target real rule decoders with unicode-escape patterns absent from upstream rows.
 replacements=json.loads(capture_overlay.read_text())['Replace']
 replacements[str(cohere/'internal/lint/rules/core/slot14_unicode_test.go')]=str(owned/'unicode_core_test.go.txt')
 replacements[str(cohere/'internal/lint/rules/typescript/slot14_unicode_test.go')]=str(owned/'unicode_typescript_test.go.txt')
 replacements[str(cohere/'internal/lint/rules/next/slot14_unicode_test.go')]=str(owned/'unicode_next_test.go.txt')
 targeted_overlay=scratch/'targeted-overlay.json'
 targeted_overlay.write_text(json.dumps({'Replace':replacements}))
 for name,package,pattern in [
  ('@next/next/no-html-link-for-pages','next','TestSlot14UnicodeRoutes'),
  ('no-restricted-exports','core','TestSlot14UnicodeExports'),
  ('no-restricted-imports','core','TestSlot14UnicodeImports'),
  ('@typescript-eslint/no-empty-object-type','typescript','TestSlot14UnicodeEmptyObject')]:
  path=scratch/(pattern+'.jsonl')
  env=dict(os.environ,SLOT14_UNICODE_CAPTURE=str(path))
  run(['go','test','-overlay='+str(targeted_overlay),'./internal/lint/rules/'+package,'-run','^'+pattern+'$','-count=1','-v','-timeout=10m'],cwd=cohere,env=env,output=evidence/('unicode-'+pattern+'.log'))
  captured=[json.loads(line) for line in path.read_text().splitlines()] if path.exists() else []
  assert captured,name+' targeted helper invocation absent'
  rows.extend(captured)
  coverage[name]['targeted_private_calls']=len(captured)
  note(name+': targeted real-rule assertions PASS; actual private calls='+str(len(captured)))
 # Byte-preserving brace and fixed-fallback controls.
 def add(source,index=0,size=1,unicode=True):
  if isinstance(source,str):source=list(source.encode())
  assert 0<=index<=len(source) and 0<=size<=len(source)-index
  rows.append({'Source':source,'Index':index,'Size':size,'Unicode':unicode,'InClass':False,'Named':False,'Groups':0})
 for value in list(range(4096))+list(range(0x10f000,0x110100)):
  add('u{'+format(value,'x')+'}')
 for byte in range(256):
  for unicode in [False,True]:add([117,123,byte,125],unicode=unicode)
 for body in ['','0','00','00000000000000','10FFFF','110000','D800','DFFF','FFFFFFFF','100000000','+41','-41','0x41','0_41',' 41','41 ','é','😀','{41','41}extra','41}}']:
  for unicode in [False,True]:
   for ending in ['', '}']:add('u{'+body+ending,unicode=unicode)
 for suffix in ['u0041','uD800','uFFFF','uG000','u123','u','u{41}','u{10ffff}','u{110000}','u{']:
  for prefix in ['', 'é', '😀', '中']:
   for unicode in [False,True]:
    raw=(prefix+suffix).encode();add(list(raw),index=len(prefix.encode()),unicode=unicode)
 for raw in [[],[123,125],[123],[0xC0,0x80],[0xED,0xA0,0x80],[0xF0,0x9F,0x98,0x80]]:
  for unicode in [False,True]:add(raw,size=0,unicode=unicode)
 rng=random.Random(14)
 for _ in range(2048):add('u{'+format(rng.randrange(1<<32),'x')+'}')
 for _ in range(2048):
  payload=[rng.randrange(256) for _ in range(rng.randrange(0,12))]
  add([117,123]+payload+([125] if rng.randrange(2) else []))
 unique={json.dumps(row,sort_keys=True):row for row in rows}
 rows=list(unique.values());inputs=scratch/'inputs.json';inputs.write_text(json.dumps(rows))
 (evidence/'unicode-coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
 note('unique controls and consumer inputs='+str(len(rows)))
 oracle_overlay=scratch/'oracle-overlay.json'
 oracle_overlay.write_text(json.dumps({'Replace':{str(original.parent/'slot14_unicode_oracle_test.go'):str(owned/'unicode_oracle_test.go.txt')}}))
 expected=scratch/'expected'
 env=dict(os.environ,SLOT14_UNICODE_INPUTS=str(inputs),SLOT14_UNICODE_OUTPUT=str(expected))
 run(['go','test','-overlay='+str(oracle_overlay),'./internal/lint/ecmascript/regexp','-run','^TestSlot14UnicodeEscape$','-count=1','-v','-timeout=10m'],cwd=cohere,env=env,output=evidence/'unicode-oracle.log')
 want=expected.read_bytes()
 build_overlay=scratch/'build-overlay.json'
 build_overlay.write_text(json.dumps({'Replace':{str(root/'slot14_unicode_escape_build.go'):str(owned/'build.go.txt')}}))
 builder=scratch/'builder'
 run(['go','build','-overlay='+str(build_overlay),'-o',builder,root/'slot14_unicode_escape_build.go'])
 binary=scratch/'native';emitted=scratch/'emitted.mjs'
 run([builder,owned/'unicode_runner.a',binary,emitted,'sanitized'])
 node=['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs']
 for side,args in [('Node',node+[owned/'unicode_runner.a',inputs]),('native',[binary,inputs]),('emitted JavaScript',node+[emitted,inputs])]:
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
 shutil.copyfile(owned/'unicode_runner.a',variant/'unicode_runner.a')
 shutil.copyfile(owned/'regexp_decode_fixed_hex.a',variant/'regexp_decode_fixed_hex.a')
 # Keep the reader as a test dependency in scratch, not a shared repository edit.
 shutil.copyfile(owned.parent/'options_json.ts',scratch/'options_json.ts')
 spec=json.loads((owned/'unicode_mutant.json').read_text())
 text=(owned/spec['file']).read_text();assert text.count(spec['from'])==1
 (variant/spec['file']).write_text(text.replace(spec['from'],spec['to']))
 mn=scratch/'mutant-native';mj=scratch/'mutant.mjs'
 run([builder,variant/'unicode_runner.a',mn,mj,'sanitized'])
 for side,args in [('Node',node+[variant/'unicode_runner.a',inputs]),('native',[mn,inputs]),('emitted JavaScript',node+[mj,inputs])]:
  assert run(args)!=want,side+' mutant survived'
  note(side+': compiling exit-zero clean-stderr closing-brace width mutant caught only by Go byte comparison')
 note('PASS actual private Go unicode-escape behavior, all four original consumer suites, byte controls and semantic mutant')
except Exception as error:
 note('FAIL '+repr(error));raise
finally:log.close()
