"""Produce reviewable owner-hook hunks without editing shared tracked files."""
import json
import sys
from pathlib import Path
root = Path(__file__).resolve().parents[4]
out = Path(sys.argv[1]).resolve()
out.mkdir(parents=True, exist_ok=True)
replacements = {
 'internal/lower/object.go': [('if censusFieldSlotless(of) && !(of == ir.MaybeBoolean && l.result.CheckedFields[name]) {', 'if censusFieldSlotless(of) && !(of == ir.MaybeBoolean && l.result.CheckedFields[name]) && !l.objectPrimitiveViewType(l.checker.GetTypeAtLocation(node)) {')],
 'internal/lower/interface_cast.go': [('if !l.viewDataType(l.checker.GetTypeOfSymbol(field)) || !known || (of < ir.Number || of > ir.Array) && of != ir.MaybeNumber && of != ir.MaybeBoolean || accessorSymbol(field) {', 'if (!l.viewDataType(l.checker.GetTypeOfSymbol(field)) || !known || (of < ir.Number || of > ir.Array) && of != ir.MaybeNumber && of != ir.MaybeBoolean) && !l.objectPrimitiveViewType(l.checker.GetTypeOfSymbol(field)) || accessorSymbol(field) {')],
 'internal/native/view_fields.go': [('func (e *emitter) viewField(property ir.Property) string {\n\tof := property.Type()', 'func (e *emitter) viewField(property ir.Property) string {\n\tof := property.Type()\n if of == ir.Union { return e.viewObjectPrimitive(property) }')],
 'internal/javascript/javascript.go': [('case ir.Property:\n\t\tif expression.View != "" {', 'case ir.Property:\n\t\tif expression.View != "" && expression.Of == ir.Union { return e.viewObjectPrimitive(expression) }\n\t\tif expression.View != "" {')],
}
result = {}
import difflib
patch = []
for relative, changes in replacements.items():
 original = (root / relative).read_text()
 modified = original
 for old, new in changes:
  assert modified.count(old) == 1, (relative, old)
  modified = modified.replace(old, new)
 target = out / relative
 target.parent.mkdir(parents=True, exist_ok=True)
 target.write_text(modified)
 result[str(root / relative)] = str(target)
 patch.extend(difflib.unified_diff(original.splitlines(keepends=True),modified.splitlines(keepends=True),fromfile='a/'+relative,tofile='b/'+relative))
(out / 'overlay.json').write_text(json.dumps({'Replace': result}, indent=2)+'\n')
Path(__file__).with_name('shared-hooks.patch').write_text(''.join(patch))
print(out / 'overlay.json')
