"""Prove Uint16-specific storage, write checking and view aliasing independently."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[4]
path = root / 'internal/native/runtime/typed_array.c'
original = path.read_text()
logs = Path('/tmp/adamic-uint16-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('wrap', '((uint16_t *)array->data)[index] = (uint16_t)modulo;',
     '((uint16_t *)array->data)[index] = isfinite(value) && value >= 65536 ? UINT16_MAX : (uint16_t)modulo;'),
    ('bounds', 'if (!valid_index(array, index)) {\n\t\tchar number',
     'if (array->kind != adamic_typed_array_uint16 && !valid_index(array, index)) {\n\t\tchar number'),
    ('copying-subarray', 'adamic_typed_array *view = adamic_allocate(sizeof *view, adamic_kind_typed_array);',
     '''if (array->kind == adamic_typed_array_uint16) {
		adamic_typed_array *copy = allocate(array->kind, to > from ? to - from : 0);
		if (copy->length > 0) { memcpy(copy->data, (unsigned char *)array->data + from * width(array->kind), copy->length * width(array->kind)); }
		return copy;
	}
	adamic_typed_array *view = adamic_allocate(sizeof *view, adamic_kind_typed_array);'''),
]
checks = [
    ('C', ['./internal/native', '-run', '^TestTypedArrayRuntime$']),
    ('backends', ['./internal/oracle', '-run', 'TestTypedArrayWriteStopIsPinned/uint16_stop|TestNativeAgreesWithNode/internal/oracle/testdata/typed_arrays_uint16']),
]
environment = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
try:
    for name, before, after in mutants:
        assert original.count(before) == 1, (name, 'stale mutation')
        path.write_text(original.replace(before, after, 1))
        for check, arguments in checks:
            log = logs / f'{name}-{check}.log'
            with log.open('w') as output:
                result = subprocess.run(['go', 'test', *arguments, '-count=1', '-timeout', '10m'], cwd=root, env=environment, stdout=output, stderr=subprocess.STDOUT)
            report = log.read_text()
            if result.returncode == 0 or 'clang failed' in report or '[build failed]' in report:
                raise RuntimeError(f'{name} {check}: survived or failed to build; see {log}')
            print(f'{name} {check}: caught; see {log}', flush=True)
        path.write_text(original)
finally:
    path.write_text(original)
print('all 3 caught by C tests and backend oracles')
