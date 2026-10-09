import pathlib,json,os,subprocess,time,difflib,shlex
D=pathlib.Path('/tmp/u103/evidence');plan=json.loads((D/'plan.json').read_text());original={m['file']:pathlib.Path(m['file']).read_text() for m in plan};commands=[];selector=pathlib.Path('/tmp/u103/selector');selector.write_text('');baseenv=os.environ.copy();baseenv['ADAMIC_GATE_SAMPLE']=(D/'origin.txt').read_text().strip();baseenv['ADAMIC_JSON_PRETTIER']='/tmp/u086/library/node_modules/prettier'
def run(id,args,env=None):
 e=baseenv.copy();e.update(env or {});t=time.monotonic()
 with (D/(id+'.log')).open('w') as out:rc=subprocess.call(args,stdout=out,stderr=subprocess.STDOUT,env=e)
 commands.append(dict(id=id,command=shlex.join(args),env=env or {},sample=baseenv['ADAMIC_GATE_SAMPLE'],selector=selector.read_text(),wall_seconds=round(time.monotonic()-t,3),exit=rc));(D/'mutant-commands.json').write_text(json.dumps(commands,indent=2));print(id,rc,flush=True);return rc
def diff(id,changes):
 text=''
 for f,n in changes.items():text+=''.join(difflib.unified_diff(original[f].splitlines(True),n.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
 (D/(id+'.diff')).write_text(text)
def regex(rows):
 return '^('+'|'.join('TestPortMatchesGoCohere_[0-9]{3}' if r=='TestPortMatchesGoCohere family' else 'TestProduct_JSON(GoOracle|LoweredPort|NativeRelease|NativeSanitized)' if r=='TestProduct_JSON family' else r for r in rows)+')$'
def matrix(id,rows,env=None):return run(id,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run',regex(rows)],env)
# Wait for clean measurements to finish before any source edit.
while not (D/'timings.json').exists() or len(json.loads((D/'timings.json').read_text()))<11:
 if 'clean timing failed' in (D/'measure.log').read_text():raise RuntimeError('measurements red')
 time.sleep(1)
for m in plan:
 f=m['file'];s=original[f];n=s.replace(m['old'],m['new']);diff(m['id'],{f:n})
 if m['kind']!='production':continue
 pathlib.Path(f).write_text(n)
 try:
  args=['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/json/main.ts','-o','/tmp/u103/validate-'+m['id'],'--sanitize'] if m['id']!='M05' else ['timeout','90','go','vet','./internal/childguard/']
  assert run('validate-'+m['id'],args,{'ADAMIC_NATIVE_SPLIT':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/u103/cache/validate-'+m['id']})==0
 finally:pathlib.Path(f).write_text(s)
# All switches select alternatives at existing sites. Helper is carried in an existing copied port module.
for f,s in original.items():
 if f.endswith('/width.ts'):
  s=s.replace("import { panic } from 'adamic';","import { panic, readTextFile } from 'adamic';\nconst auditFile = readTextFile('/tmp/u103/selector');\nconst auditChoice = auditFile.kind === 'Error' ? '' : auditFile.text.trim();\nexport function auditSelected(id: string): boolean { return auditChoice === id; }")
  s=s.replace('return text.length;',"return text.length + (auditSelected('M04') ? 1 : 0);")
 if f.endswith('/formatter.ts'):
  s="import { auditSelected } from './width.ts';\n"+s.replace("base === 'package-lock.json'", "base === (auditSelected('M01') ? 'lock.json' : 'package-lock.json')")
  s=s.replace('export function format(name: string, text: string): string {',"export function format(name: string, text: string): string {\n    if(auditSelected('PPort')) { return ''; }")
 if f.endswith('/parser.ts'):
  s="import { auditSelected } from './width.ts';\n"+s.replace("this.text.slice(0, this.position).split('\\n').length", "(this.text.slice(0, this.position).split('\\n').length + (auditSelected('M02') ? 1 : 0))")
 if f.endswith('/doc.ts'):
  s=s.replace('import { stringWidth }','import { stringWidth, auditSelected }').replace('command.indent + 2',"command.indent + (auditSelected('M03') ? 1 : 2)")
 if f=='internal/childguard/childguard.go':
  s=s.replace('if len(p) > 0 {','if (os.Getenv("ADAMIC_MUTANT") != "M05" && len(p) > 0) || (os.Getenv("ADAMIC_MUTANT") == "M05" && len(p) == 0) {')
  s=s.replace('func CombinedOutput(cmd *exec.Cmd, options Options) ([]byte, error) {','func CombinedOutput(cmd *exec.Cmd, options Options) ([]byte, error) {\nif os.Getenv("ADAMIC_MUTANT") == "PGuard" {return nil,nil}')
 if f.endswith('/port_matches_prepared_split_test.go'):
  s=s.replace('func portMatchesPrepare(t *testing.T) {','func portMatchesPrepare(t *testing.T) {\nif os.Getenv("ADAMIC_MUTANT") == "PSetup" {return}')
  s=s.replace('func portMatchesProduct(t *testing.T, name string) string {','func portMatchesProduct(t *testing.T, name string) string {\nif os.Getenv("ADAMIC_MUTANT") == "PProducts" {return ""}')
 if f.endswith('/repository_test.go'):
  s=s.replace('func validateCorpus(root string, cases []textCase, expected corpusPin) (corpusPin, int, error) {','func validateCorpus(root string, cases []textCase, expected corpusPin) (corpusPin, int, error) {\nif os.Getenv("ADAMIC_MUTANT") == "PCorpus" {return corpusPin{},0,nil}')
 pathlib.Path(f).write_text(s)
diff('switch',{f:pathlib.Path(f).read_text() for f in original})
assert run('switch-vet',['timeout','90','go','vet','./internal/childguard/','./stage1/cohere/json/'])==0
ports=['TestPortMatchesGoCohere family','TestAdditionalJSONBoundaries','TestSingleFileStdoutDriver']
assert matrix('switch-clean',ports,{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u103/cache/switch'})==0
for m in plan:
 if m['kind']!='production':continue
 id=m['id'];selector.write_text(id);matrix(id,m['rows'],{'ADAMIC_MUTANT':id,'ADAMIC_BUILD_CACHE_DIR':'/tmp/u103/cache/'+('M05' if id=='M05' else 'switch')})
for id,rows in [('PPort',ports),('PGuard',['TestProgressGuard']),('PSetup',['TestPortMatchesGoCohereSplit_Setup']),('PProducts',['TestProduct_JSON family']),('PCorpus',['TestRepositoryLandingWithoutPinEdit','TestRepositoryRequiresGit'])]:
 selector.write_text(id);matrix(id,rows,{'ADAMIC_MUTANT':id,'ADAMIC_BUILD_CACHE_DIR':'/tmp/u103/cache/'+('PGuard' if id=='PGuard' else 'switch')})
selector.write_text('')
for f,s in original.items():pathlib.Path(f).write_text(s)
for m in plan:
 if m['kind']=='production':continue
 f=m['file'];s=original[f];pathlib.Path(f).write_text(s.replace(m['old'],m['new']))
 try:
  assert run('vet-'+m['id'],['timeout','90','go','vet','./stage1/cohere/json/'])==0
  matrix(m['id'],m['rows'],{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u103/cache/'+m['id']})
 finally:pathlib.Path(f).write_text(s)
print('all sources restored',flush=True)
