import pathlib,json,re,subprocess,time,difflib,os
root=pathlib.Path('/workspace/adamic'); ev=root/'review/test-audit/internal-lower-library_node_fs_file'; src=root/'internal/lower'; base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(); rows=json.loads((ev/'rows.json').read_text())
def run(cmd,name,env=None):
 t=time.monotonic()
 with (ev/name).open('w') as f:r=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,env=env)
 return {'exit':r.returncode,'wall':time.monotonic()-t,'command':' '.join(cmd)}
if os.sys.argv[1]=='timings':
 out={}
 for row in rows:
  out[row]=[run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],row+f'-{i}.log') for i in range(3)]
 (ev/'timings.json').write_text(json.dumps(out,indent=2));exit()
if os.sys.argv[1]=='prepare':
 specs=[
 ('library_node_fs_file.go','index := 1','index := 0','index := auditValue("M01", 1, 0)'),
 ('library_node_fs_buffer.go','len(call.Arguments.Nodes) != 5','len(call.Arguments.Nodes) != 4','auditValue("M02", len(call.Arguments.Nodes) != 5, len(call.Arguments.Nodes) != 4)'),
 ('library_node.go','!load.IsNodeLibrary(file)','load.IsNodeLibrary(file)','auditValue("M03", !load.IsNodeLibrary(file), load.IsNodeLibrary(file))'),
 ('library_node.go','owners = append([]string{parent.Name().Text()}, owners...)','','if !auditMutant("M04") { owners = append([]string{parent.Name().Text()}, owners...) }'),
 ('library_node.go','!implementedNodeMembers[name]','implementedNodeMembers[name]','auditValue("M05", !implementedNodeMembers[name], implementedNodeMembers[name])'),
 ('library_node_fs_file.go','!discarded && !returned','!discarded || !returned','auditValue("M06", !discarded && !returned, !discarded || !returned)'),
 ('library_node_fs_file.go','"fs option literals containing evaluated expressions; bind a plain options object first"','"unsupported options"','auditValue("M07", "fs option literals containing evaluated expressions; bind a plain options object first", "unsupported options")'),
 ('library_node_fs_file.go','l.result.Strings[text.Index] == "utf8"','l.result.Strings[text.Index] == "ascii"','auditValue("M08", l.result.Strings[text.Index] == "utf8", l.result.Strings[text.Index] == "ascii")'),
 ('library_node_fs_file.go','args[1].Type() != ir.String','args[1].Type() == ir.String','auditValue("M09", args[1].Type() != ir.String, args[1].Type() == ir.String)'),
 ('library_node_fs_file.go','!ok || b.Value','!ok || !b.Value','!ok || auditValue("M10", b.Value, !b.Value)'),
 ('library_node_fs_file.go','fallback = number(100)','fallback = number(101)','fallback = number(auditValue("M11", 100.0, 101.0))'),
 ('library_node_fs_file.go','operation, of = "exists", ir.Boolean','operation, of = "stat", ir.Boolean','operation, of = auditValue("M12", "exists", "stat"), ir.Boolean'),
 ('library_node_fs_file.go','throws := ir.Expression(ir.BooleanConstant{Value: true})','throws := ir.Expression(ir.BooleanConstant{Value: false})','throws := ir.Expression(ir.BooleanConstant{Value: !auditMutant("M13")})'),
 ('invariance.go','from == to || visited[[2]*checker.Type{from, to}]','from != to || visited[[2]*checker.Type{from, to}]','auditValue("M14", from == to, from != to) || visited[[2]*checker.Type{from, to}]'),
 ('optional_widening.go','"declare " + found.property + " on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"','"declare " + found.property','auditValue("M15", "declare " + found.property + " on the source type, or build a fresh object with known fields (adamic/no-optional-widening)", "declare " + found.property)'),
 ('refusals.go','"a method read as a value (" + symbol.Name','"a method read as a value (" + "method"','"a method read as a value (" + auditValue("M16", symbol.Name, "method")'),
 ('library_node_fs_file.go','len(call.Arguments.Nodes) > index &&','len(call.Arguments.Nodes) > index+1 &&','auditValue("M17", len(call.Arguments.Nodes) > index, len(call.Arguments.Nodes) > index+1) &&'),
 ('library_node_fs_file.go','len(call.Arguments.Nodes) != 5','len(call.Arguments.Nodes) != 4','auditValue("M18", len(call.Arguments.Nodes) != 5, len(call.Arguments.Nodes) != 4)'),
 ('library_node_fs_file.go','operation = "rm"','operation = "unlink"','operation = auditValue("M19", "rm", "unlink")'),
 ('library_node_fs_buffer.go','load.IsNodeLibrary(file) &&','!load.IsNodeLibrary(file) &&','auditValue("M20", load.IsNodeLibrary(file), !load.IsNodeLibrary(file)) &&')]
 originals={f:(src/f).read_text() for f,_,_,_ in specs}; switched=dict(originals); inventory=[]
 for i,(f,old,new,switch) in enumerate(specs,1):
  mid=f'M{i:02}'; original=originals[f];assert old in original,(mid,old)
  mutant=original.replace(old,new,1);line=original[:original.index(old)].count('\n')+1
  diff=''.join(difflib.unified_diff(original.splitlines(True),mutant.splitlines(True),fromfile='a/internal/lower/'+f,tofile='b/internal/lower/'+f))
  (ev/(mid+'.diff')).write_text(diff)
  inventory.append({'id':mid,'file':f,'line':line,'old':old,'new':new})
  switched[f]=switched[f].replace(old,switch,1)
 (ev/'plan.json').write_text(json.dumps(inventory,indent=2));(ev/'base.txt').write_text(base+'\n')
 for m in inventory:
  subprocess.run(['git','apply','--check',str(ev/(m['id']+'.diff'))],cwd=root,check=True)
  subprocess.run(['git','apply',str(ev/(m['id']+'.diff'))],cwd=root,check=True)
  result=run(['timeout','90','go','vet','./internal/lower/'],m['id']+'-vet.log');assert result['exit']==0,(m,result)
  subprocess.run(['git','restore','--source=HEAD','--','internal/lower/'+m['file']],cwd=root,check=True)
 for f,s in switched.items():(src/f).write_text(s)
 (src/'audit_selector.go').write_text('package lower\nimport "os"\nfunc auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\nfunc auditValue[T any](id string, normal, mutant T) T {if auditMutant(id) {return mutant};return normal}\n')
 for f,signature,value,pid in [('lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','return nil, nil','P01'),('library_node.go','func (l *lowering) nodeLibraryMember(node *ast.Node) string {','return ""','P02')]:
  p=src/f;s=p.read_text();p.write_text(s.replace(signature,signature+'\n if auditMutant("'+pid+'") { '+value+' }',1))
 subprocess.run(['gofmt','-w']+[str(src/f) for f in list(originals)+['lower.go','audit_selector.go']],check=True)
 exit()
if os.sys.argv[1]=='matrix':
 results={}
 for mid in [f'M{i:02}' for i in range(1,21)]+['P01','P02']:
  env=dict(os.environ,ADAMIC_MUTANT=mid,ADAMIC_BUILD_CACHE_DIR='/tmp/u035/cache/'+mid)
  results[mid]=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],mid+'.log',env)
  (ev/'matrix-status.json').write_text(json.dumps(results,indent=2))
