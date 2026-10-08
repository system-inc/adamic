#!/usr/bin/env python3
"""Prepare shared hook hunks without editing another lane's files."""
import difflib,json,subprocess,sys
from pathlib import Path
root=Path(__file__).resolve().parents[3]
out=Path(sys.argv[1]);out.mkdir(parents=True,exist_ok=True)
hooks=[
 ('internal/lower/expression.go','\t\treturn l.objectIntersection(proven)','\t\tif l.structuralViewIntersection(proven) { return ir.Object, true }\n\t\treturn l.objectIntersection(proven)'),
 ('internal/lower/view_lazy.go','\tcase flags&checker.TypeFlagsIntersection != 0:\n\t\treturn "intersection"','\tcase flags&checker.TypeFlagsIntersection != 0:\n\t\tif l.structuralViewIntersection(target) { return "" }\n\t\treturn "intersection"'),
 ('internal/lower/view_contracts.go','\tbuild := func(child *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, child) }','\tif l.structuralViewIntersection(target) { return l.internStructuralViewIntersection(node, target) }\n\tbuild := func(child *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, child) }'),
 ('internal/lower/view_objects.go','func (l *lowering) viewDataType(target *checker.Type) bool {','func (l *lowering) viewDataType(target *checker.Type) bool {\n\tif l.structuralViewIntersection(target) { return true }'),
]
replace={};patch=[]
for filename,old,new in hooks:
 original=(root/filename).read_text();assert original.count(old)==1,filename
 modified=original.replace(old,new)
 destination=out/Path(filename).name;destination.write_text(modified)
 subprocess.run(['gofmt','-w',str(destination)],check=True)
 modified=destination.read_text()
 replace[str(root/filename)]=str(destination)
 patch.extend(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/'+filename,tofile='b/'+filename))
(out/'overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
(root/'stage3/interface-downcasts/lane7/shared-hooks.patch').write_text('INCOMPLETE ADMISSION HANDOFF: do not apply until all-member runtime dispatch is wired.\nRoot-only conjunction probe currently fails; flattened metadata is not certification.\n\n'+''.join(patch))
print(out/'overlay.json')
