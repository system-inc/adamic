"""Run independent runtime mutants; restore sources even when a mutant survives."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[4]
runtime = root / 'internal/native/runtime/typed_array.c'
heap = root / 'internal/native/runtime/heap.c'
logs = Path('/tmp/adamic-typed-array-mutants')
logs.mkdir(exist_ok=True)
originals = {runtime: runtime.read_text(), heap: heap.read_text()}
mutants = [
    ('missing-wrap', runtime, 'fmod(trunc(value), modulus)', 'trunc(value)'),
    ('missing-write-bounds', runtime, 'if (!valid_index(array, index)) {\n\t\tchar number', 'if (false && !valid_index(array, index)) {\n\t\tchar number'),
    ('copying-subarray', runtime,
     'view->data = (unsigned char *)array->data + from * width(array->kind);\n\tview->owner = adamic_retain(array->owner != NULL ? array->owner : (adamic_typed_array *)array);',
     'view->data = calloc(view->length == 0 ? 1 : view->length, width(array->kind));\n\tif (view->length > 0) { memcpy(view->data, (unsigned char *)array->data + from * width(array->kind), view->length * width(array->kind)); }\n\tview->owner = NULL;'),
    ('missing-view-retain', runtime,
     'adamic_retain(array->owner != NULL ? array->owner : (adamic_typed_array *)array)',
     '(array->owner != NULL ? array->owner : (adamic_typed_array *)array)'),
    ('missing-iterator-retain', runtime, 'iterator->array = adamic_retain(array);', 'iterator->array = array;'),
    ('missing-buffer-free', heap, 'free(array->data);', '(void)array->data;'),
    ('overlap-memcpy', runtime, 'memmove((unsigned char *)array->data', 'memcpy((unsigned char *)array->data'),
    ('missing-kind-check', runtime, 'if (array->kind != source->kind)', 'if (false && array->kind != source->kind)'),
    ('missing-offset-check', runtime, 'if (!(integer >= 0) || integer > (double)array->length)', 'if (false && (!(integer >= 0) || integer > (double)array->length))'),
    ('missing-set-capacity-check', runtime, 'if (source->length > array->length - from)', 'if (false && source->length > array->length - from)'),
    ('missing-length-check', runtime,
     'if (!(integer >= 0) || integer > 9007199254740991.0 || integer >= (double)SIZE_MAX)',
     'if (false && (!(integer >= 0) || integer > 9007199254740991.0 || integer >= (double)SIZE_MAX))'),
    ('fractional-index-accepted', runtime, ' && index == trunc(index)', ''),
]
try:
    for name, path, before, after in mutants:
        assert originals[path].count(before) == 1, (name, 'stale mutation')
        path.write_text(originals[path].replace(before, after, 1))
        with (logs / (name + '.log')).open('w') as output:
            result = subprocess.run(['go', 'test', './internal/native', '-run', '^TestTypedArrayRuntime$', '-count=1', '-timeout', '10m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        path.write_text(originals[path])
        report = (logs / (name + '.log')).read_text()
        if result.returncode == 0 or 'clang failed' in report or '[build failed]' in report:
            raise RuntimeError(f'{name}: survived or failed to build; see {logs / (name + ".log")}')
        catcher = ('LeakSanitizer' if 'LeakSanitizer' in report else
                   'ASan' if 'AddressSanitizer' in report else
                   'UBSan' if 'runtime error:' in report else
                   'Node comparison' if 'differs from Node' in report else
                   'panic assertion' if 'wanted panic' in report else
                   'runtime assertion')
        print(f'{name}: caught by {catcher}', flush=True)
finally:
    for path, original in originals.items():
        path.write_text(original)
print(f'all {len(mutants)} caught; logs: {logs}')
