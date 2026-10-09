import pathlib,json,subprocess,time,os,difflib
root=pathlib.Path('.');p=root/'review/test-audit/internal-native-arguments_length';menu=json.loads((p/'menu.json').read_text());names=json.loads((p/'names.json').read_text());pattern='^('+ '|'.join(names)+')$'
base={x['file']:pathlib.Path(x['file']).read_text() for x in menu};base['internal/native/native.go']=pathlib.Path('internal/native/native.go').read_text()
(p/'diffs').mkdir(exist_ok=True)
stats=[]
def run(id,cmd,env=None):
 start=time.monotonic()
 with open(p/(id+'.log'),'w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT,env=env)
 stats.append(dict(id=id,seconds=time.monotonic()-start,exit=r.returncode,command=' '.join(cmd)));(p/'commands.json').write_text(json.dumps(stats,indent=2));print(id,r.returncode,round(stats[-1]['seconds'],2),flush=True)
 return r.returncode
def diff(id,file,text):
 (p/'diffs'/f'{id}.diff').write_text(''.join(difflib.unified_diff(base[file].splitlines(True),text.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
def restore():
 for f,t in base.items():pathlib.Path(f).write_text(t)
# Standalone Go diffs validated independently; C diffs are compiled by their native matrix builds.
for m in menu:
 f=m['file'];t=base[f].replace(m['old'],m['new']);diff(m['id'],f,t)
 if f.endswith('.go'):
  pathlib.Path(f).write_text(t);assert run(m['id']+'-vet',['timeout','120','go','vet','./internal/native/'])==0;pathlib.Path(f).write_text(base[f])
# One switched Go build.
expressions={
'M01':('return false\n\t}\n\treturn pure(expression)','return auditChoice("M01", false, true)\n\t}\n\treturn pure(expression)'),
'M02':('true','auditChoice("M02", true, false)'),
'M03':('names[write.Name]','auditChoice("M03", names[write.Name], !names[write.Name])'),
'M04':('targets.Unknown','auditChoice("M04", targets.Unknown, !targets.Unknown)'),
'M05':('element || chain','auditChoice("M05", element || chain, element && chain)'),
'M06':(menu[5]['old'],'auditChoice("M06", '+menu[5]['old']+', '+menu[5]['new']+')'),
'M07':('append([]int{signature}, targets...)','auditChoice("M07", append([]int{signature}, targets...), []int{})'),
'M08':('plan.feedsRegion(statement)','auditChoice("M08", plan.feedsRegion(statement), !plan.feedsRegion(statement))'),
'M09':('return true','return auditChoice("M09", true, false)'),
'M10':('program.ClosureConventionNeeded()','auditChoice("M10", program.ClosureConventionNeeded(), !program.ClosureConventionNeeded())'),
'M11':('closure, packed, count)','closure, packed, auditChoice("M11", count, count+" + 1"))'),
'M12':('e.program.StructuralMethodThunks[function]','auditChoice("M12", e.program.StructuralMethodThunks[function], false)'),
'M13':('"-ffp-contract=off"','auditChoice("M13", "-ffp-contract=off", "-ffp-contract=fast")'),
'M17':('strings.Contains(source, "#define "+feature+" 1\\n")','auditChoice("M17", strings.Contains(source, "#define "+feature+" 1\\n"), !strings.Contains(source, "#define "+feature+" 1\\n"))'),
}
for m in menu:
 if m['id'] not in expressions:continue
 a,b=expressions[m['id']];replacement=m['old'].replace(a,b);assert replacement!=m['old'],m['id'];f=m['file'];t=pathlib.Path(f).read_text();assert t.count(m['old'])==1,m['id'];pathlib.Path(f).write_text(t.replace(m['old'],replacement))
helper=pathlib.Path('internal/native/audit_switch.go');helper.write_text('package native\nimport "os"\nfunc auditChoice[T any](id string, original, changed T) T { if os.Getenv("ADAMIC_MUTANT") == id { return changed }; return original }\n')
# Entry probes do not count as mutants.
probes=[('P_C','internal/native/emit.go','func C(program *ir.Program) string {','return ""'),('P_CHAIN','internal/native/borrow.go','func chainUnchanged(program *ir.Program, body []ir.Statement, names map[string]bool) bool {','return false'),('P_CONSUMES','internal/native/borrow.go','func consumes(expression ir.Expression) bool {','return false'),('P_ELEMENTS','internal/native/element_borrow.go','func planElementBorrows(program *ir.Program) (map[*ir.Statement]bool, map[int]bool) {','return nil, nil'),('P_REGIONS','internal/native/region.go','func planRegions(program *ir.Program) *regionPlan {','return &regionPlan{fresh: map[int]bool{}, escapes: map[int]map[int]bool{}, statements: map[*ir.Statement]bool{}, program: program, classObjects: map[int]int{}}'),('P_BUILD','internal/native/native.go','func Build(source string, output string, options Options) error {','return nil'),('P_LIBRARY','internal/native/library.go','func RuntimeLibraryForSource(directory string, source string, options Options) (string, error) {','return "", nil')]
for id,f,entry,ret in probes:
 standalone=base[f].replace(entry,entry+'\n\t'+ret);diff(id,f,standalone)
 t=pathlib.Path(f).read_text();pathlib.Path(f).write_text(t.replace(entry,entry+'\n\tif auditChoice("'+id+'", false, true) { '+ret+' }'))
# Witness check weakening: discard clang rejection in Build, while still invoking clang.
f='internal/native/native.go';old='return fmt.Errorf("native: clang failed: %w\\n%s", err, combined)';diff('W_BUILD',f,base[f].replace(old,'_ = combined\n\t\treturn nil'));t=pathlib.Path(f).read_text();pathlib.Path(f).write_text(t.replace(old,'if auditChoice("W_BUILD", false, true) { return nil }; '+old))
(p/'probes.json').write_text(json.dumps([dict(id=i,file=f,entry=e,change=r) for i,f,e,r in probes],indent=2));(p/'switch.diff').write_text(subprocess.check_output(['git','diff']).decode()+helper.read_text())
run('switched-vet',['timeout','120','go','vet','./internal/native/'])
for id in [x['id'] for x in menu]+[x[0] for x in probes]+['W_BUILD']:
 env=os.environ.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u045/cache/'+id
 cm=next((m for m in menu if m['id']==id and not m['file'].endswith('.go')),None)
 if cm:pathlib.Path(cm['file']).write_text(base[cm['file']].replace(cm['old'],cm['new']))
 run(id,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern],env)
 if cm:pathlib.Path(cm['file']).write_text(base[cm['file']])
restore();helper.unlink()
