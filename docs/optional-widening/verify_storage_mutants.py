#!/usr/bin/env python3
"""Run storage mutants separately and restore each source in finally."""
import os
from pathlib import Path
import subprocess

repository = Path(__file__).resolve().parents[2]
evidence = repository / 'docs/optional-widening/evidence/view-write-storage'
evidence.mkdir(parents=True, exist_ok=True)
scalar = '''    if (actual == 9 && wanted == 2) {
        adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(slot->maybe_boolean);
        if (boolean.present) { return (adamic_value){.boolean = boolean.boolean}; }
    }
'''
snapshot = '''    } else if (storage == 9) {
        adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(slot->maybe_boolean);
        value.kind = boolean.present ? adamic_view_union_boolean : adamic_view_union_undefined;
        value.payload.boolean = boolean.boolean;
'''
dictionary = '''        } else if (actual == 9) {
            adamic_maybe_boolean maybe = adamic_maybe_boolean_unpack(slot->maybe_boolean);
            if (maybe.present) value = (adamic_view_union_value){adamic_view_union_boolean, {.boolean = maybe.boolean}};
'''
mutants = [
 ('raw-undefined', 'internal/native/view_writes.go', 'adamic_maybe_boolean_pack((adamic_maybe_boolean){false,false})', '(uint8_t)0', './internal/oracle', '^TestViewWriteStorageCoherence$/boolean-reader$'),
 ('raw-maybe-pair', 'internal/native/view_writes.go', 'packed = "adamic_maybe_boolean_pack(" + value + ")"', 'packed = "(uint8_t)(" + value + ").boolean"', './internal/oracle', '^TestViewWriteStorageCoherence$/boolean$'),
 ('boxed-nullable', 'internal/native/runtime/view_nullish.c', 'else if(actual==9){adamic_maybe_boolean value=adamic_maybe_boolean_unpack(slot->maybe_boolean);kind=value.present?2:13;boolean=value.boolean;}', 'else if(actual==9){reference=slot->reference;kind=reference==NULL?13:logical_kind(reference);if(kind==2)boolean=((adamic_boolean_box *)reference)->boolean;}', './internal/oracle', '^TestViewWriteStorageCoherence$/boolean-nullish$'),
 ('omit-required-presence', 'internal/native/runtime/object.c', 'if (boolean.present) { return (adamic_value){.boolean = boolean.boolean}; }', 'return (adamic_value){.boolean = boolean.boolean};', './internal/oracle', '^TestViewWriteStorageCoherence$/boolean-required-undefined$'),
 ('omit-scalar-unpack', 'internal/native/runtime/object.c', scalar, '', './internal/oracle', '^TestViewWriteStorageCoherence$/boolean-required$'),
 ('omit-optional-absence', 'internal/native/runtime/object.c', ' || (actual == 9 && !adamic_maybe_boolean_unpack(slot->maybe_boolean).present)', '', './internal/native', '^TestPackedBooleanViewSnapshot$'),
 ('omit-snapshot-unpack', 'internal/native/runtime/object.c', snapshot, '', './internal/native', '^TestPackedBooleanViewSnapshot$'),
 ('omit-dictionary-unpack', 'internal/native/runtime/view_dictionaries.c', dictionary, '', './internal/oracle', '^TestViewWriteStorageCoherence$/boolean-dictionary$'),
]
for name, relative, old, new, package, pattern in mutants:
    path = repository / relative
    original = path.read_text()
    assert original.count(old) == 1, (name, original.count(old))
    print('running ' + name, flush=True)
    try:
        path.write_text(original.replace(old, new, 1))
        with (evidence / (name + '.log')).open('w') as output:
            result = subprocess.run(['go', 'test', '-count=1', '-timeout', '15m', package, '-run', pattern, '-v'], cwd=repository, stdout=output, stderr=subprocess.STDOUT, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'))
        log = (evidence / (name + '.log')).read_text()
        assert result.returncode != 0, name + ' survived'
        assert 'native: clang failed' not in log and '[build failed]' not in log, name + ' did not execute'
        assert '--- FAIL:' in log, name + ' did not reach its test'
        if name in ('raw-undefined', 'raw-maybe-pair'):
            assert 'stdout="false\\n' in log or 'stdout="undefined\\ntrue\\nfalse\\nundefined\\nfalse' in log
            assert 'exit=0' in log
        print(name + ': caught, exit=' + str(result.returncode), flush=True)
    finally:
        path.write_text(original)
