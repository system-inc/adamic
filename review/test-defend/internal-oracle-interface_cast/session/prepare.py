import pathlib,json,subprocess
p=pathlib.Path('review/test-defend/internal-oracle-interface_cast/session')
menu=[
('D1','TestInterfaceCastOracle','internal/lower/interface_cast.go','for _, member := range declared.Types() {','for _, member := range declared.Types()[:len(declared.Types())-1] {','off by one: omit last member of finite checked-view union'),
('D2','TestInterfaceCastScalarTags','internal/lower/cast.go','l.checker.TypeToString(literal) == "true"','l.checker.TypeToString(literal) != "true"','flip boolean literal decoding'),
('D3','TestInterfaceCastImportedConstruction','internal/lower/cast_proof.go','if err == nil {','if err != nil {','flip successful upcast proof condition'),
('D4','TestInterfaceCastImportedConstruction','internal/lower/cast_proof.go','l.checker.IsTypeAssignableTo(source, target) || l.checker.IsTypeAssignableTo(l.checker.GetWidenedType(source), target)','l.checker.IsTypeAssignableTo(target, source) || l.checker.IsTypeAssignableTo(l.checker.GetWidenedType(source), target)','swap assignability arguments for upcast proof'),
('D5','TestInterfaceCastImportedConstruction','internal/lower/cast.go','if proof.interfaceView {','if !proof.interfaceView {','flip interface-view path selection'),
('D6','TestInterfaceCastOracle','internal/native/view_fields.go','if len(property.ViewAllowed) != 0 {','if len(property.ViewAllowed) == 0 {','flip finite literal field-check condition'),
('D7','TestInterfaceCastOracle','internal/lower/interface_cast.go','allowed = append(allowed, values...)','allowed = append(allowed, values[:len(values)-1]...)','off by one finite literal list concatenation')]
for mid,target,file,old,new,change in menu:
 path=pathlib.Path(file); orig=path.read_text();assert old in orig
 line=orig[:orig.index(old)].count('\n')+1
 try:
  path.write_text(orig.replace(old,new,1)); diff=subprocess.check_output(['git','diff','--',file]);(p/(mid+'.diff')).write_bytes(diff)
 finally:path.write_text(orig)
(p/'menu.json').write_text(json.dumps([dict(zip(['mutant','target','file','old','new','change'],m)) for m in menu],indent=2))
