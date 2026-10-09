import pathlib,subprocess,json,time,difflib,os,re,shutil
ROOT=pathlib.Path.cwd();OUT=ROOT/'review/test-audit/stage1-cohere-markdownblocks-structure_layout_shards';BASE=ROOT/'stage1/cohere/markdownblocks';PKG='./stage1/cohere/markdownblocks/'
selector=pathlib.Path('/tmp/adamic-u130-mutant');meta=[]
rows=dict(json.loads((OUT/'row-patterns.json').read_text()))
original={str(p.relative_to(ROOT)):p.read_text() for p in BASE.glob('*.ts')}
original.update({str(p.relative_to(ROOT)):p.read_text()for p in BASE.glob('*_test.go')})
def diff(id,file,new):
 old=original[file];(OUT/(id+'.diff')).write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
def run(id,pattern,env=None):
 s=time.monotonic();cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run',pattern]
 with(OUT/(id+'.log')).open('w')as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=dict(os.environ,**(env or {})))
 events=[]
 for l in(OUT/(id+'.log')).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 fails=sorted({e['Test']for e in events if e['Action']=='fail'and'Test'in e and'/'not in e['Test']})
 cooked=any('test timed out' in e.get('Output','')or 'cooked:'in e.get('Output','')for e in events)
 m=dict(id=id,command=cmd,wall=time.monotonic()-s,exit=r.returncode,fails=fails,cooked=cooked)
 meta.append(m);(OUT/'matrix-meta.json').write_text(json.dumps(meta,indent=2));print(m,flush=True);return m
# Mutants and probes were selected and written to planning.md before this script.
planned=json.loads((OUT/'planned-mutants.json').read_text())
for m in planned:diff(m['id'],m['file'],original[m['file']].replace(m['original'],m['replacement'],1))
probes=[('P1','width.ts','export function stringWidth(text: string): number {','return 0;'),('P2','splitText.ts','export function splitText(text: string): readonly TextTokenInterface[] {','return [];'),('P3','document.ts','): string {\n    propagate(arena, root);',"return '';"),('P4','whitespace.ts','export function printWhitespace(arena: DocumentArena, frame: WhitespaceFrameInterface): number {',"return arena.text('');")]
# Standalone probes replace the function body, avoiding unreachable-code diagnostics.
def body(s,start):
 a=s.index('{',start);level=1;i=a+1
 while level:
  if s[i]=='{':level+=1
  elif s[i]=='}':level-=1
  i+=1
 return a,i
for id,name,anchor,value in probes:
 file=str((BASE/name).relative_to(ROOT));s=original[file];idx=s.index(anchor);a,b=body(s,idx)
 diff(id,file,s[:a]+'{\n    '+value+'\n}'+s[b:])
# Runtime selection is in port sources, never in the Go oracle or fixture drivers.
instrument={}
for name in ['width.ts','structure.ts','whitespace.ts','splitText.ts','document.ts']:
 file=str((BASE/name).relative_to(ROOT));s=original[file]
 s="import { readTextFile } from 'adamic';\nconst auditSelection = readTextFile('/tmp/adamic-u130-mutant');\nconst auditMode = auditSelection.kind === 'Error' ? '' : auditSelection.text;\n"+s
 instrument[file]=s
for m in planned:
 s=instrument[m['file']]
 if m['id']=='M1':s=s.replace(m['original'],"if(ascii) return auditMode === 'M1' ? text.length + 1 : text.length;",1)
 elif m['id']=='M2':s=s.replace('if(!setext) return arena.concat',"if(auditMode === 'M2' ? setext : !setext) return arena.concat",1)
 elif m['id']=='M3':s=s.replace(m['original'],"if(proseWrap !== (auditMode === 'M3' ? 'never' : 'always')) return false;",1)
 else:s=s.replace(m['original'],"if(unit === (auditMode === 'M4' ? 9 : 10)) newline = true;",1)
 instrument[m['file']]=s
for id,name,anchor,value in probes:
 file=str((BASE/name).relative_to(ROOT));s=instrument[file]
 if id=='P3':s=s.replace(anchor,"): string {\n    if(auditMode === 'P3') "+value+'\n    propagate(arena, root);',1)
 else:s=s.replace(anchor,anchor+"\n    if(auditMode === '"+id+"') "+value,1)
 instrument[file]=s
# Preserve a combined scratch diff for reproducibility, then restore all files in finally.
try:
 selector.write_text('')
 for file,s in instrument.items():(ROOT/file).write_text(s)
 (OUT/'switch.diff').write_text(subprocess.check_output(['git','diff','--','stage1/cohere/markdownblocks'],text=True))
 # Cold switched products are bounded separately from the kill matrix.
 products=['TestProduct_MarkdownStructure_lowered','TestProduct_MarkdownStructure_native','TestProduct_MarkdownTable_lowered','TestProduct_MarkdownTable_native_true','TestProduct_MarkdownTable_native_false','TestProduct_MarkdownWhitespace_manifest','TestProduct_MarkdownTextLowered','TestProduct_MarkdownTextNativeSanitized','TestProduct_MarkdownTextNativeRelease']
 for name in products:
  m=run('build-'+name,'^'+name+'$')
  if m['exit'] and not m['cooked']:raise RuntimeError('switched source did not build: '+name)
 control=run('instrumented-control','^(TestMarkdownStructureLayout_000|TestMarkdownTableLayout_000|TestMarkdownWhitespaceLayout_000|TestMarkdownTextSplitting_000)$')
 if control['exit']:raise RuntimeError('instrumentation control failed')
 # Only reached unit rows: other package rows are unknown after the baseline cooked.
 layouts=['structure-family','table-family','whitespace-family']
 reaches={'M1':layouts+['width'],'M2':layouts,'M3':layouts,'M4':layouts+['text-family'],'P1':layouts+['width'],'P2':layouts+['text-family'],'P3':layouts,'P4':layouts}
 for id in reaches:
  selector.write_text(id)
  pattern='|'.join('(?:'+rows[r]+')'for r in reaches[id]);m=run(id,pattern)
  if m['cooked']:
   for r in reaches[id]:run(id+'-'+r,rows[r])
 # Instrumented control ensures the selector itself does not change answers.
 selector.write_text('')
finally:
 for file in instrument:(ROOT/file).write_text(original[file])
 selector.write_text('')
# Construction mutations and witness weakening are the explicitly allowed harness exception.
changes=[('S1','support_test.go',None,'sanitize: options.Sanitize','sanitize: false','native-modes'),('S2','table_layout_shards_test.go','tableLayoutReady',None,'return;','table-setup'),('S3','text_independent_shards_test.go','textReadySetup',None,'return textSetupState{};','text-setup'),('S4','whitespace_layout_split_test.go','whitespaceLayoutReadyProducts',None,'return whitespaceLayoutProducts{}, whitespaceLayoutProducts{};','whitespace-setup'),('W1','text_independent_shards_test.go','textOutputDifference',None,'return nil;','text-witness'),('W2','whitespace_layout_split_test.go',None,'return bytes.Equal(actual, expected)','return true','whitespace-witness')]
for id,name,func,anchor,value,row in changes:
 file=str((BASE/name).relative_to(ROOT));s=original[file]
 if id in ['S2','S3','S4']:
  once={'S2':'tableLayoutOnce','S3':'textSetupOnce','S4':'whitespaceLayoutReadyOnce'}[id]
  start=s.index(once+'.Do(',s.index('func '+func+'('))
  a,b=body(s,start)
  assert s[b]==')'
  new=s[:start]+s[b+1:]
 elif func:
  a,b=body(s,s.index('func '+func+'('));new=s[:a]+'{\n    '+value+'\n}'+s[b:]
 else:
  assert s.count(anchor)==1;new=s.replace(anchor,value,1)
 diff(id,file,new)
 try:
  (ROOT/file).write_text(new)
  start=time.monotonic()
  with(OUT/(id+'-vet.log')).open('w')as f:v=subprocess.run(['timeout','90','go','vet',PKG],stdout=f,stderr=subprocess.STDOUT)
  print(id,'vet',v.returncode,time.monotonic()-start,flush=True)
  if v.returncode:raise RuntimeError('standalone vet failed '+id)
  run(id,rows[row])
 finally:(ROOT/file).write_text(s)
(OUT/'mutation-complete.json').write_text(json.dumps({'complete':True,'time':time.time()},indent=2))
