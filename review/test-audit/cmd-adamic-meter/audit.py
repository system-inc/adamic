import pathlib,subprocess,time,json,difflib,os,re,statistics
root=pathlib.Path('/workspace/adamic'); out=pathlib.Path('/tmp/u009'); pkg='./cmd/adamic-meter/'
files=['main.go','optional.go','returns.go']; base={f:(root/'cmd/adamic-meter'/f).read_text() for f in files}
def rec(i,f,old,new,menu):
 assert base[f].count(old)==1,(i,base[f].count(old))
 return dict(id=i,file='cmd/adamic-meter/'+f,line=base[f][:base[f].index(old)].count('\n')+1,old=old,new=new,menu=menu)
ms=[
rec('M1','main.go','existing.Count++','existing.Count += 2','change constant'),
rec('M2','main.go','result.FilesExamined++','result.FilesExamined += 2','change constant'),
rec('M3','main.go','if *asJSON {','if !*asJSON {','flip condition'),
rec('M4','main.go','location.ReplaceAllString(message, "<location>")','location.ReplaceAllString(message, "<place>")','change constant'),
rec('M5','main.go','typeText.ReplaceAllString(message, "of type <type>")','typeText.ReplaceAllString(message, "of type <other>")','change constant'),
rec('M6','main.go','line[:at] + "type " + line[at:]','line[:at] + "" + line[at:]','change constant'),
rec('M7','main.go','if text == "1484" {','if text == "1485" {','change constant'),
rec('M8','main.go','diagnosticCount(err, "2379")','diagnosticCount(err, "2378")','change constant'),
rec('M9','optional.go','units := 1','units := 2','change constant'),
rec('M10','optional.go','if r > 0xffff {','if r > 0x10ffff {','change constant'),
rec('M11','optional.go','!owned[ast.GetSourceFileOfNode(declaration).FileName().AsString()]','owned[ast.GetSourceFileOfNode(declaration).FileName().AsString()]','flip condition'),
rec('M12','optional.go','symbol.Flags&ast.SymbolFlagsOptional == 0','symbol.Flags&ast.SymbolFlagsOptional != 0','flip condition'),
rec('M13','optional.go','return receiver.Flags()&(checker.TypeFlagsTypeParameter|checker.TypeFlagsIndexedAccess|checker.TypeFlagsAnyOrUnknown) != 0','return receiver.Flags()&(checker.TypeFlagsTypeParameter|checker.TypeFlagsIndexedAccess|checker.TypeFlagsAnyOrUnknown) == 0','flip condition'),
rec('M14','optional.go',"if c.GetTypeAtLocation(conditional.WhenTrue).Flags()&checker.TypeFlagsUndefined != 0 || c.GetTypeAtLocation(conditional.WhenFalse).Flags()&checker.TypeFlagsUndefined != 0 {\n\t\t\treturn false","if c.GetTypeAtLocation(conditional.WhenTrue).Flags()&checker.TypeFlagsUndefined != 0 || c.GetTypeAtLocation(conditional.WhenFalse).Flags()&checker.TypeFlagsUndefined != 0 {\n\t\t\treturn true",'change constant'),
rec('M15','optional.go','if node.Kind == ast.KindElementAccessExpression {','if node.Kind == ast.KindPropertyAccessExpression {','change constant'),
rec('M16','optional.go','") | undefined", ownerStart','") | null", ownerStart','change constant'),
rec('M17','optional.go','return list[i].start > list[j].start','return list[i].start < list[j].start','flip condition'),
rec('M18','returns.go','diagnosticCount(before, "7030") == 0','diagnosticCount(before, "7030") != 0','flip condition'),
rec('M19','main.go','return overlay, rewrites, nil\n}','return overlay, nil, nil\n}','drop returned rewrites'),
rec('M20','optional.go','NoUncheckedIndexedAccess: core.TSTrue','NoUncheckedIndexedAccess: core.TSFalse','change option')]
ps=[]
for i,f,start,ret in [('E_MEASURE','main.go','func measure(root string, adapt bool) (*report, error) {','return nil, nil'),('E_RUN','main.go','func run(arguments []string, stdout, stderr io.Writer) int {','return 0'),('E_NORMALIZE','main.go','func normalize(message string) string {','return ""'),('E_ADAPT','main.go','func adaptations(paths []string, baselineErr error) (map[string]string, []adaptation, error) {','return nil, nil, nil'),('E_RETURN','returns.go','func returnAdaptations(paths []string, overlay map[string]string, before error) (map[string]string, int, error) {','return nil, 0, nil')]:
 ps.append(rec(i,f,start,start+'\n if true { '+ret+' }','empty-answer probe'))
(out/'mutants.json').write_text(json.dumps(ms+ps,indent=2)); (out/'base-source.json').write_text(json.dumps(base))
def run(cmd,name,env=None):
 t=time.monotonic()
 with open(out/(name+'.log'),'w') as log:r=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT,env=env)
 result=dict(name=name,command=cmd,seconds=time.monotonic()-t,exit=r.returncode)
 with open(out/'commands.jsonl','a') as log:log.write(json.dumps(result)+'\n')
 return result
names=[l.strip() for l in (out/'list.log').read_text().splitlines() if l.startswith('Test')]
(out/'names.json').write_text(json.dumps(names))
# Complete baseline timings before installing instrumentation.
for name in names:
 for n in range(3):run(['timeout','120','go','test','-count=1','-timeout','90s',pkg,'-run','^'+name+'$'],name+'-'+str(n))
# Compile-check every independent diff, against original source.
for m in ms+ps:
 f=m['file'].split('/')[-1]; changed=base[f].replace(m['old'],m['new']);
 diff=''.join(difflib.unified_diff(base[f].splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file']))
 (out/(m['id']+'.diff')).write_text(diff)
 (root/m['file']).write_text(changed)
 r=run(['go','vet',pkg],m['id']+'-vet');(root/m['file']).write_text(base[f]);assert r['exit']==0,m['id']
# Switch expressions/statements preserving each independent mutation.
sw=base.copy()
for m in ms:
 f=m['file'].split('/')[-1];old,new=m['old'],m['new'];sel='os.Getenv("ADAMIC_MUTANT") == "'+m['id']+'"'
 if old.startswith('if '):
  expr=old[3:old.index(' {')];nexpr=new[3:new.index(' {')]
  # M14 is a multi-line constant return, handle separately.
  if m['id']=='M14':replacement=old.replace('return false','return os.Getenv("ADAMIC_MUTANT") == "M14"')
  else:replacement='if (('+sel+') && ('+nexpr+')) || (!( '+sel+') && ('+expr+')) {'
 elif old=='existing.Count++' or old=='result.FilesExamined++':replacement='if '+sel+' { '+new+' } else { '+old+' }'
 elif old.startswith('return receiver.'):replacement='return ('+sel+' && ('+new[7:]+')) || (!('+sel+') && ('+old[7:]+'))'
 elif m['id']=='M19':replacement='if '+sel+' { return overlay, nil, nil };\n'+old
 elif m['id'] in ['M11','M12']:replacement='(('+sel+') != ('+old+'))'
 else:
  # Generic value selector closure, special literal snippets require replacing the value.
  if m['id']=='M9':replacement='units := 1; if '+sel+' { units = 2 }'
  elif m['id']=='M16':replacement='func() string { if '+sel+' { return ") | null" }; return ") | undefined" }(), ownerStart'
  elif m['id']=='M8':replacement='func() int { if '+sel+' { return '+new+' }; return '+old+' }()'
  elif m['id']=='M18':replacement='(('+sel+' && ('+new+')) || (!('+sel+') && ('+old+')))'
  elif m['id']=='M20':replacement='NoUncheckedIndexedAccess: func() core.Tristate { if '+sel+' { return core.TSFalse }; return core.TSTrue }()'
  elif m['id']=='M17':replacement='return func() bool { if '+sel+' { return list[i].start < list[j].start }; return list[i].start > list[j].start }()'
  else:replacement='func() string { if '+sel+' { return '+new+' }; return '+old+' }()'
 sw[f]=sw[f].replace(old,replacement,1)
for p in ps:
 f=p['file'].split('/')[-1];sw[f]=sw[f].replace(p['old'],p['new'].replace('if true','if os.Getenv("ADAMIC_MUTANT") == "'+p['id']+'"'),1)
sw['returns.go']=sw['returns.go'].replace('"context"','"context"\n "os"')
for f,text in sw.items():(root/'cmd/adamic-meter'/f).write_text(text)
run(['gofmt','-w']+[str(root/'cmd/adamic-meter'/f) for f in files],'switch-gofmt')
r=run(['go','test','-c','-o',str(out/'meter.test'),pkg],'switch-build');assert r['exit']==0
(out/'switch.diff').write_text(subprocess.check_output(['git','diff','--','cmd/adamic-meter'],cwd=root,text=True))
for m in ms+ps:
 env=os.environ.copy();env['ADAMIC_MUTANT']=m['id']
 run(['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'],m['id']+'-matrix',env)
print('done')
