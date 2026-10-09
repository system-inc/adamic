import pathlib,json,time,subprocess,difflib,os,shlex
D=pathlib.Path('/tmp/u103/evidence')
while 'all sources restored' not in (D/'mutation-runner.log').read_text():
 if 'Traceback' in (D/'mutation-runner.log').read_text():raise RuntimeError('runner failed')
 time.sleep(1)
env=os.environ.copy();records=[]
def run(id,args,overrides=None):
 e=env.copy();e.update(overrides or {});t=time.monotonic()
 with (D/(id+'.log')).open('w') as out:rc=subprocess.call(args,stdout=out,stderr=subprocess.STDOUT,env=e)
 records.append(dict(id=id,command=shlex.join(args),env=overrides or {},seconds=round(time.monotonic()-t,3),exit=rc));(D/'finish-commands.json').write_text(json.dumps(records,indent=2));print(id,rc,flush=True);return rc
# Survivor output is observed using a built standalone mutant, against restored source Node.
case=pathlib.Path('/tmp/u103/M01-cases.txt');case.write_text('package-lock.json\t[1,2]\n');(D/'M01-cases.txt').write_text(case.read_text())
assert run('M01-witness-before',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs','stage1/cohere/json/main.ts','--cases',str(case)])==0
assert run('M01-witness-after',['/tmp/u103/validate-M01','--cases',str(case)],{'ASAN_OPTIONS':'detect_leaks=1'})==0
assert (D/'M01-witness-before.log').read_text()!=(D/'M01-witness-after.log').read_text()
# All probes are replayable without unrelated mutation alternatives.
probes=[('PPort','stage1/cohere/json/formatter.ts','export function format(name: string, text: string): string {',"export function format(name: string, text: string): string {\n    if(auditSelected('PPort')) { return ''; }"),('PGuard','internal/childguard/childguard.go','func CombinedOutput(cmd *exec.Cmd, options Options) ([]byte, error) {','func CombinedOutput(cmd *exec.Cmd, options Options) ([]byte, error) {\nif os.Getenv("ADAMIC_MUTANT") == "PGuard" { return nil,nil }'),('PSetup','stage1/cohere/json/port_matches_prepared_split_test.go','func portMatchesPrepare(t *testing.T) {','func portMatchesPrepare(t *testing.T) {\nif os.Getenv("ADAMIC_MUTANT") == "PSetup" { return }'),('PProducts','stage1/cohere/json/port_matches_prepared_split_test.go','func portMatchesProduct(t *testing.T, name string) string {','func portMatchesProduct(t *testing.T, name string) string {\nif os.Getenv("ADAMIC_MUTANT") == "PProducts" { return "" }'),('PCorpus','stage1/cohere/json/repository_test.go','func validateCorpus(root string, cases []textCase, expected corpusPin) (corpusPin, int, error) {','func validateCorpus(root string, cases []textCase, expected corpusPin) (corpusPin, int, error) {\nif os.Getenv("ADAMIC_MUTANT") == "PCorpus" { return corpusPin{},0,nil }')]
for id,f,old,new in probes:
 originals={f:pathlib.Path(f).read_text()};changes={f:originals[f].replace(old,new)}
 if id=='PPort':
  wf='stage1/cohere/json/width.ts';originals[wf]=pathlib.Path(wf).read_text();changes[wf]=originals[wf].replace("import { panic } from 'adamic';","import { panic, readTextFile } from 'adamic';\nconst auditFile = readTextFile('/tmp/u103/selector');\nconst auditChoice = auditFile.kind === 'Error' ? '' : auditFile.text.trim();\nexport function auditSelected(id: string): boolean { return auditChoice === id; }");changes[f]="import { auditSelected } from './width.ts';\n"+changes[f]
 patch=''
 for cf,n in changes.items():patch+=''.join(difflib.unified_diff(originals[cf].splitlines(True),n.splitlines(True),fromfile='a/'+cf,tofile='b/'+cf));pathlib.Path(cf).write_text(n)
 (D/(id+'.diff')).write_text(patch)
 args=['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/json/main.ts','-o','/tmp/u103/validate-'+id,'--sanitize'] if id=='PPort' else ['timeout','90','go','vet','./internal/childguard/','./stage1/cohere/json/']
 try:assert run('validate-'+id,args,{'ADAMIC_NATIVE_SPLIT':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/u103/cache/validate-'+id})==0
 finally:
  for cf,s in originals.items():pathlib.Path(cf).write_text(s)
for p in D.glob('*.diff'):
 if p.name=='switch.diff':continue
 subprocess.run(['git','apply','--check',str(p)],check=True)
assert subprocess.call(['git','diff','--exit-code'])==0
# Restore the same fixed sampled scope, with no instrumentation or harness weakening.
regex='^(TestPortMatchesGoCohereSplit_Setup|TestPortMatchesGoCohereSplitUnion|TestPortMatchesGoCohere_[0-9]{3}|TestProduct_JSON(GoOracle|LoweredPort|NativeRelease|NativeSanitized)|TestThreePortMutantsAreCaught|TestAdditionalJSONBoundaries|TestSingleFileStdoutDriver|TestProgressGuard|TestRepositoryCorpusMutants|TestRepositoryLandingWithoutPinEdit|TestRepositoryRequiresGit)$'
assert run('restored-baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run',regex],{'ADAMIC_GATE_SAMPLE':(D/'origin.txt').read_text().strip(),'ADAMIC_JSON_PRETTIER':'/tmp/u086/library/node_modules/prettier'})==0
print('finished',flush=True)
