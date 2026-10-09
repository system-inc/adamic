import pathlib,subprocess,json,time,os,difflib
root=pathlib.Path('/workspace/adamic'); ev=pathlib.Path('/tmp/u124/evidence'); package='./stage1/cohere/lint/rules/typescript-no-this-alias/'; file='stage1/cohere/lint/rules/typescript-no-this-alias/rule.a'; entry='stage1/cohere/lint/rules/typescript-no-this-alias/profile.a'; source=root/file; base=source.read_text(); runs=[]
def run(id,command,env={}):
 start=time.monotonic()
 with (ev/(id+'.log')).open('w') as f: p=subprocess.run(['timeout','120',*command],cwd=root,env={**os.environ,**env},stdout=f,stderr=subprocess.STDOUT)
 r={'id':id,'command':['timeout','120',*command],'environment':env,'wall':time.monotonic()-start,'exit':p.returncode}; runs.append(r); (ev/'runs.json').write_text(json.dumps(runs,indent=2)); print(id,p.returncode,round(r['wall'],3),flush=True); return p.returncode
for i in range(1,4): assert run('timing-'+str(i),['go','test','-json','-count=1','-timeout','90s',package,'-run','^TestCompileProfiles$'])==0
assert run('compiler-build',['go','build','-o','/tmp/u124/adamic','./cmd/adamic'])==0
(ev/'input.ts').write_text('const self = this;\n(self) = this;\n'); (ev/'manifest.txt').write_text(str(ev/'input.ts')+'\n')
def build_and_observe(id):
 target='/tmp/u124/'+id+'-native'; cache={'ADAMIC_BUILD_CACHE_DIR':'/tmp/u124/cache/'+id}
 assert run(id+'-build',['/tmp/u124/adamic','build',entry,'-o',target],cache)==0
 assert run(id+'-output',[target,str(ev/'manifest.txt')])==0
 assert run(id+'-count',[target,str(ev/'manifest.txt'),'--count'])==0
build_and_observe('clean')
mutants=[('M1',"if(context.node(right).kind !== 'ThisKeyword')", "if(context.node(right).kind === 'ThisKeyword')",'flip condition'),('M2','while(wrappers.includes(context.node(target).kind))','while(false)','change condition constant, disables loop'),('M3',"'thisAssignment', assignment","'thisAssignmentAudit', assignment",'change constant'),('P1','visit(node: ParseNode, index: number): void {','visit(node: ParseNode, index: number): void { return;','empty entry probe')]
catalog=[]
for id,old,new,kind in mutants:
 assert base.count(old)==1
 changed=base.replace(old,new,1); (ev/(id+'.diff')).write_text(''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 catalog.append({'id':id,'file':file,'line':base[:base.index(old)].count('\n')+1,'before':old,'after':new,'kind':kind})
(ev/'catalog.json').write_text(json.dumps(catalog,indent=2))
(ev/'scope.md').write_text('Starting commit ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2. Live package row: TestCompileProfiles, no family. CODE UNDER TEST: owned rule.a (Rule constructor, Rule.visit, create) reached through profile.a (ancestry, visit, run and top-level manifest driver). It also reaches imported Parser, Scanner, written, RuleContext and Settings implementations, along with their dependency graph. Those shared compiler/parser/context functions are not mutated in this owned-port audit. The test loads and lowers the whole profile, then renders native C twice and JavaScript once; none of the resulting products is executed by the test. ORACLE: Adamic own successful compile/build and JavaScript rendering, self. No Go cohere, Node or ESLint comparison is invoked. Fixed menu selected from port source before mutant results: M1 flip this-input gate; M2 replace unwrapping condition with false; M3 change assignment diagnostic ID constant. P1 Rule.visit returns its empty void answer at entry. No tests, harness, oracle adapters or upstream Go rules are changed. Three production mutants use fresh native rebuilds rather than a runtime switch; standalone native builds and test-native builds are recorded separately.\n')
try:
 for id,old,new,kind in mutants:
  source.write_text(base.replace(old,new,1))
  code=run(id,['go','test','-json','-count=1','-timeout','90s',package,'-run','.'],{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u124/cache/'+id})
  if code==0: build_and_observe(id)
  else: print(id+' matrix failed, inspect before witness',flush=True)
finally: source.write_text(base)
assert run('package-vet',['go','vet',package])==0
