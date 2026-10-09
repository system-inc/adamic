import pathlib,subprocess,os,time,json,re,difflib
root=pathlib.Path('.'); out=root/'review/test-audit/stage1-cohere-lint-yepesta_independent_shards';plan=json.loads((out/'frozen-plan.json').read_text()); records=json.loads((out/"runs.json").read_text()) if (out/"runs.json").exists() else []
pattern='^(TestMutantsReactJsxNoCommentTextnodes.*|TestProduct_YepestaOracle)$'
groups={'comparison':'^TestMutantsReactJsxNoCommentTextnodes(Union|_[0-9]{3})$','killed':'^TestMutantsReactJsxNoCommentTextnodesMutantKilled$','planted':'^TestMutantsReactJsxNoCommentTextnodesPlantedFailure$','setup':'^TestMutantsReactJsxNoCommentTextnodes(OracleSetup|LoweredSetup|NativeSetup|_Setup)$','product':'^TestProduct_YepestaOracle$'}
def run(name,pat=pattern,vet=False):
 for prior in records:
  if prior['name']==name:return prior
 cmd=['timeout','120','go','vet','./stage1/cohere/lint/'] if vet else ['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',pat]
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for line in (out/(name+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 rec={'name':name,'command':' '.join(cmd),'wall':time.monotonic()-start,'exit':r.returncode,'fails':[e['Test'] for e in events if e.get('Action')=='fail' and 'Test'in e],'passes':[e['Test'] for e in events if e.get('Action')=='pass' and 'Test'in e],'elapsed':[e.get('Elapsed') for e in events if e.get('Action')in ['pass','fail'] and 'Test'not in e],'cooked':'test timed out' in (out/(name+'.log')).read_text() or r.returncode==124}
 records.append(rec);(out/'runs.json').write_text(json.dumps(records,indent=2)+'\n');print(name,rec['exit'],rec['wall'],rec['fails'],flush=True);return rec
for group,pat in groups.items():
 for i in range(3):
  r=run('timing-'+group+'-'+str(i+1),pat)
  if r['exit']:raise SystemExit('red timing baseline')
plan=json.loads((out/'frozen-plan.json').read_text())
changes=plan['mutants']+[{'id':'P1','file':'stage1/cohere/lint/main.ts','from':'function run(row: string, countOnly: boolean): number {','to':'function run(row: string, countOnly: boolean): number {\n    return 0;'}]
for m in changes:
 f=root/m['file'];original=f.read_text();assert original.count(m['from'])==1
 modified=original.replace(m['from'],m['to']);(out/(m['id']+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
 try:
  f.write_text(modified)
  r=run(m['id'])
  if r['cooked']:r=run(m['id']+'-warm-lowered')
  if r['cooked']:print('COOKED unable to settle '+m['id'],flush=True)
 finally:f.write_text(original)
for ident,file,old,new,pat in [
 ('W1','stage1/cohere/lint/lint_test.go','func difference(got, want []byte) string {','func difference(got, want []byte) string {\n\treturn "";',groups['planted']),
 ('W2','stage1/cohere/lint/yepesta_independent_shards_test.go','if bytes.Equal(side.output, want) {','if true {',groups['killed']),
 ('S1','stage1/cohere/lint/yepesta_independent_shards_test.go','_, err := goOracleIn(".", out)\n\t\treturn err','return nil',groups['product']+'|'+groups['setup']),
 ('S2','stage1/cohere/lint/yepesta_independent_shards_test.go','data, err = json.Marshal(bundle)','data, err = json.Marshal(yepestaBundle{})',groups['setup'])]:
 f=root/file;original=f.read_text();assert original.count(old)==1
 if ident == "W1":
  old=original[original.index("func difference(got, want []byte) string {"):original.index("\nfunc manifest(")]
  new="func difference(got, want []byte) string { return \"\" }\n"
 modified=original.replace(old,new);(out/(ident+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 try:
  f.write_text(modified);run(ident+'-vet',vet=True);run(ident,pat)
 finally:f.write_text(original)
