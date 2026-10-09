import pathlib,subprocess,time,json,os
root=pathlib.Path('/tmp/u059'); out=pathlib.Path('review/test-audit/internal-oracle-conditions')
while not (root/'timings-done').exists(): time.sleep(1)
menu=json.loads((out/'menu.json').read_text()); base=(out/'base.txt').read_text().strip()
files={m['file'] for m in menu}|{'internal/native/emit.go','internal/oracle/oracle_test.go','internal/oracle/conditions_test.go'}
original={f:subprocess.check_output(['git','show',base+':'+f],text=True) for f in files}
for m in menu:
 subprocess.run(['git','apply',str(out/(m['id']+'.diff'))],check=True)
 start=time.monotonic()
 with (root/(m['id']+'-vet.log')).open('w') as log:
  result=subprocess.run(['timeout','90','go','vet','./'+str(pathlib.Path(m['file']).parent)+'/'],stdout=log,stderr=subprocess.STDOUT)
 (root/(m['id']+'-vet.meta')).write_text(json.dumps({'exit':result.returncode,'wall':time.monotonic()-start}))
 pathlib.Path(m['file']).write_text(original[m['file']])
 if result.returncode: raise SystemExit(m['id']+' failed vet')
scratch=original.copy()
repls={
'M1':'\tif !auditMutant("M1") { counters(lowering.result) }',
'M2':'\t\tif !auditMutant("M2") { e.line("debugger;") }',
'M3':'case ir.ObjectCall:\n\t\tif expression.Checked && !auditMutant("M3") {',
'M4':'if auditMutant("M4") { return "!Boolean(" + e.values(expression.Arguments) + ")" }; return "Boolean(" + e.values(expression.Arguments) + ")"',
'M5':'func (l *lowering) namespaceInitialization(modules []*ast.SourceFile) error {\n\tif auditMutant("M5") { return nil }',
'M6':'if expression.Checked && !auditMutant("M6") {\n\t\t\treturn fmt.Sprintf("(%s ? %s : adamicUnready(%s))"',
'M7':'if auditMutant("M7") { return "a " + name }; return "an " + name',
'M8':'\tif !auditMutant("M8") { borrow(lowering.result) }'}
for m in menu: scratch[m['file']]=scratch[m['file']].replace(m['old'],repls[m['id']],1)
entries=[('internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','P1','nil, nil'),('internal/native/emit.go','func C(program *ir.Program) string {','P2','""'),('internal/javascript/javascript.go','func JavaScript(program *ir.Program) string {','P3','""'),('internal/oracle/oracle_test.go','func disagreement(oracle run, native run) string {','W1','""'),('internal/oracle/conditions_test.go','func conditionLedgerDifference(source, stdout []byte) error {','W2','nil')]
for file,entry,mid,value in entries: scratch[file]=scratch[file].replace(entry,entry+'\n\tif auditMutant("'+mid+'") { return '+value+' }',1)
for f,s in scratch.items(): pathlib.Path(f).write_text(s)
for pkg in ['lower','native','javascript','oracle']:
 f=pathlib.Path('internal')/pkg/'audit_u059_test_switch.go';f.write_text('package '+pkg+'\nimport "os"\nfunc auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\n')
subprocess.run(['gofmt','-w',*scratch.keys(),*[str(pathlib.Path('internal')/p/'audit_u059_test_switch.go') for p in ['lower','native','javascript','oracle']]],check=True)
with (out/'scratch-switch.diff').open('w') as log: subprocess.run(['git','diff'],stdout=log,check=True)
rows=json.loads((root/'rows.json').read_text()); selected=[x for x in rows if x!='TestCountsAreRecorded']; pattern='^('+'|'.join(selected)+')$'
def run(mid,pat,suffix=''):
 env=os.environ.copy();env['ADAMIC_MUTANT']=mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u059/cache/'+mid
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pat]
 start=time.monotonic()
 with (root/(mid+suffix+'.log')).open('w') as log: p=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 (root/(mid+suffix+'.meta')).write_text(json.dumps({'command':' '.join(cmd),'exit':p.returncode,'wall':time.monotonic()-start,'rows':selected if pat==pattern else [pat]}))
 return p.returncode
if run('clean-switch',pattern) != 0: raise SystemExit('clean switch baseline failed')
for m in menu:
 run(m['id'],pattern)
 run(m['id'],'^TestCountsAreRecorded$','-counts')
for row in selected:
 if row in ['TestConditionNaNMutant','TestConditionLedgerMissingSiteMutant','TestConditionLedgerNullMutant','TestEntriesAcceptance']: continue
 run('P1','^'+row+'$','-'+row)
for mid in ['P2','P3']: run(mid,pattern)
run('W1','^(TestConditionNaNMutant|TestConditionLedgerNullMutant)$')
run('W2','^TestConditionLedgerMissingSiteMutant$')
for f,s in original.items(): pathlib.Path(f).write_text(s)
for pkg in ['lower','native','javascript','oracle']: (pathlib.Path('internal')/pkg/'audit_u059_test_switch.go').unlink()
(root/'matrix-done').write_text('done')
