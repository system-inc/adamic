"""Real production-source mutants, restored even after a failed command.
Run with the setup environment sourced. No other tests may run concurrently:
the mutated compiler/runtime is deliberate, and each build embeds that version.
"""
import json
import os
from pathlib import Path
import re
import subprocess

REPO = Path(__file__).resolve().parents[2]
LOGS = Path(os.environ.get('MAPLIKE_MUTANT_LOGS', '/tmp/maplike-records-evidence/mutants'))
LOGS.mkdir(parents=True, exist_ok=True)


def run(label, args, environment=None):
    result = subprocess.run(args, cwd=REPO, env=environment, capture_output=True)
    (LOGS / (label + '.stdout')).write_bytes(result.stdout)
    (LOGS / (label + '.stderr')).write_bytes(result.stderr)
    return result


def replace(path, before, after):
    original = path.read_text()
    if original.count(before) != 1:
        raise RuntimeError(f'{path}: mutant anchor is not unique')
    path.write_text(original.replace(before, after, 1))


fixed_lower = '''for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
        name := property.Name()
        if name != nil && (name.Kind == ast.KindIdentifier || name.Kind == ast.KindStringLiteral || name.Kind == ast.KindNumericLiteral) {
            found := false
            for _, fixed := range result.Fixed { if fixed == name.Text() { found = true } }
            if !found { result.Fixed = append(result.Fixed, name.Text()) }
        }
    }
    for i, p := range node.AsObjectLiteralExpression().Properties.Nodes {'''
fixed_keys = '''adamic_object *shape_view = adamic_object_new(record->shape);
    adamic_array *fixed = adamic_object_keys(shape_view);
    adamic_release(shape_view);
    adamic_array *shaped = adamic_array_new(keys->length, true);
    for (size_t index = 0; index < fixed->length; index++) {
        adamic_string *key = fixed->elements[index].reference;
        if (adamic_record_has_own(record, key)) {
            adamic_array_push(shaped, (adamic_value){.reference = adamic_retain(key)});
        }
    }
    for (size_t index = 0; index < keys->length; index++) {
        adamic_string *key = keys->elements[index].reference;
        if (fixed_slot(record, key) == NULL) {
            adamic_array_push(shaped, (adamic_value){.reference = adamic_retain(key)});
        }
    }
    adamic_release(fixed);
    adamic_release(keys);
    return shaped;'''
mutants = [
    ('index-keys-in-fixed-shape', 'stage3/fixtures/records/05_integer_order.a', [
        ('internal/lower/records.go', 'for i, p := range node.AsObjectLiteralExpression().Properties.Nodes {', fixed_lower),
        ('internal/native/runtime/record.c', 'return keys;', fixed_keys),
    ]),
    ('optional-view-dictionary-miss', 'stage3/fixtures/records/07_optional_view.a', [
        ('internal/native/runtime/adamic.h', 'return adamic_object_dictionary_field(object, name);\n\t}', 'return NULL;\n\t}'),
    ]),
]
summary = []
for name, source, patches in mutants:
    originals = {REPO / path: (REPO / path).read_bytes() for path, _, _ in patches}
    try:
        for path, before, after in patches:
            replace(REPO / path, before, after)
        compiler = LOGS / (name + '.adamic')
        built = run(name + '.compiler', ['go', 'build', '-o', str(compiler), './cmd/adamic'])
        if built.returncode:
            raise RuntimeError(name + ': compiler build failed; this does not count as a kill')
        binary = LOGS / (name + '.native')
        built = run(name + '.build', [str(compiler), 'build', source, '-o', str(binary), '--sanitize', '--count'])
        if built.returncode:
            raise RuntimeError(name + ': native build failed; this does not count as a kill')
        environment = os.environ.copy()
        environment.update(ASAN_OPTIONS='detect_leaks=1:halt_on_error=1', UBSAN_OPTIONS='halt_on_error=1', LSAN_OPTIONS='exitcode=23')
        native = run(name + '.native', [str(binary)], environment)
        node = run(name + '.node', ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', source])
        counts = re.fullmatch(rb'adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n', native.stderr)
        if native.returncode or node.returncode or node.stderr or counts is None:
            raise RuntimeError(name + ': mutant must exit 0 with no sanitizer/leak finding')
        if int(counts[1]) != int(counts[2]) + int(counts[6]):
            raise RuntimeError(name + ': counted ownership must balance')
        if native.stdout == node.stdout:
            raise RuntimeError(name + ': survived Node comparison')
        if name == 'index-keys-in-fixed-shape':
            if sorted(native.stdout.strip().split(b',')) != sorted(node.stdout.strip().split(b',')):
                raise RuntimeError(name + ': must change order alone, retaining every key')
        summary.append({'mutant': name, 'caught': 'Node stdout', 'native_exit': 0,
                        'native_stdout': native.stdout.decode(), 'node_stdout': node.stdout.decode(),
                        'counts': native.stderr.decode(), 'sanitizers_and_leaks': 'clean'})
        print(name + ': caught only by Node stdout; valid compiler/C, exit 0, balanced counts, clean sanitizers/leaks', flush=True)
    finally:
        for path, original in originals.items():
            path.write_bytes(original)
(LOGS / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
