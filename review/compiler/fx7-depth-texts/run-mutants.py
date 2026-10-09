from pathlib import Path
import difflib
import subprocess

root = Path.cwd()
evidence = root / 'review/compiler/fx7-depth-texts'

def removed_block(path, start, end):
    text = Path(path).read_text()
    left = text.index(start)
    right = text.index(end, left)
    return text[left:right], ''

native = 'internal/native/runtime/view_unions_untagged.c'
js = 'internal/javascript/view_unions_untagged.go'
mixed = 'internal/native/runtime/view_unions_mixed.c'
object_path = 'internal/native/runtime/object.c'
mutants = [
 ('native-depth-128', native, '    ADAMIC_CHECK_STACK();', '    if (depth > 128) { return false; }', './internal/oracle', '^TestFX7Depth65$'),
 ('javascript-depth-128', js, '  const contract=contracts[id-1];', '  if(depth>128) return false;\n  const contract=contracts[id-1];', './internal/oracle', '^TestFX7Depth65$'),
 ('native-stack-guard', native, '    ADAMIC_CHECK_STACK();', '    (void)depth;', './internal/native', '^TestFX7ViewCycleAndStackGuard$'),
 ('union-found-kind', mixed, *removed_block(mixed, '    if (value->kind == adamic_view_union_object && value->payload.reference != NULL) {', '    size_t capacity = strlen(expression)'), './internal/oracle', '^TestFX7TextP(67|72)$'),
 ('scalar-found-kind', object_path, *removed_block(object_path, '\tif ((actual == adamic_rep_object ||', '\tsize_t capacity = strlen(expression) + strlen(type)'), './internal/oracle', '^TestFX7TextScalarTuple$'),
 ('native-declared-name', 'internal/native/emit_statements.go', 'cString(expected)', 'cString(map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string"}[statement.Value.Type()])', './internal/oracle', '^TestFX7TextP(32|33)$'),
 ('javascript-declared-name', 'internal/javascript/readiness.go', 'type, declared || adamicViewTypeNames[type] || "unknown");', 'type);', './internal/oracle', '^TestFX7TextP(32|33)$'),
 ('javascript-optional-frames', js, *removed_block(js, '  // A defined value can only use', '  const reference=value'), './internal/oracle', '^TestFX7Depth2000$'),
 ('native-cycle', native, 'if (seen->contract == id && seen->reference == value->payload.reference) { return true; }', 'if (seen->contract == id && seen->reference == value->payload.reference) { return false; }', './internal/native', '^TestFX7ViewCycleAndStackGuard$'),
 ('javascript-cycle', js, 'if(seen?.has(id)) return true;', 'if(seen?.has(id)) return false;', './internal/javascript', '^TestFX7ViewCycle$'),
]
for name, filename, before, after, package, test in mutants:
    path = Path(filename)
    original = path.read_text()
    assert original.count(before) == 1, (name, original.count(before))
    changed = original.replace(before, after)
    patch = ''.join(difflib.unified_diff(original.splitlines(True), changed.splitlines(True), fromfile='a/'+filename, tofile='b/'+filename))
    (evidence/(name+'.patch')).write_text(patch)
    command = ['go', 'test', package, '-run', test, '-v', '-count=1', '-timeout', '90s']
    try:
        path.write_text(changed)
        with (evidence/(name+'.log')).open('w') as log:
            result = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=100)
        output = (evidence/(name+'.log')).read_text()
        assert result.returncode != 0 and '--- FAIL: TestFX7' in output, (name, output)
        assert '[build failed]' not in output and 'error: ' not in output, (name, output)
        print(name+': caught by '+test, flush=True)
    finally:
        path.write_text(original)
