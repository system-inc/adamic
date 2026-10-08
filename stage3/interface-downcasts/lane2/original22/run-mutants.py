"""Run isolated semantic guard failures and restore every source file."""
import os
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[4]
logs = pathlib.Path(os.environ.get('TMPDIR', '/tmp')) / 'adamic-lane2-original22-mutants'
logs.mkdir(exist_ok=True)
cases = [
    ('native-number-field', 'internal/native/runtime/object.c',
     'unsigned char actual = adamic_object_field_types(owner)[cache->index];\n\t// Boxed',
     'unsigned char actual = adamic_object_field_types(owner)[cache->index]; if (wanted == 1) return *slot;\n\t// Boxed'),
    ('javascript-number-field', 'internal/javascript/readiness.go',
     '(type === 1 || type === 7) ? typeof value === "number"',
     '(type === 1 || type === 7) ? true'),

]
cases = [(name + "-" + family + "-" + tag, relative, before, after, family + "-" + tag + "-wrong-pos")
         for name, relative, before, after in cases
         for family, tag in [("signature", "180"), ("signature", "324")]]
for name, relative, before, after, witness in cases:
    file = root / relative
    original = file.read_bytes()
    assert original.decode().count(before) == 1, name
    try:
        file.write_text(original.decode().replace(before, after))
        log = logs / (name + '.log')
        with log.open('wb') as output:
            result = subprocess.run([
                'go', 'test', './internal/oracle', '-run',
                '^TestCheckedViewRanked22OriginalArrays$/^' + witness + '$',
                '-count=1', '-v', '-timeout', '10m'], cwd=root,
                env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'},
                stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        assert result.returncode != 0 and '--- FAIL:' in observed, observed
        assert witness in observed and 'stderr' in observed, observed
        assert all(marker not in observed for marker in [
            '[build failed]', 'clang failed', 'SyntaxError', 'compiler bug:',
            "stage 0 can't", 'SKIP']), observed
        print(name + ': executed guard mismatch; ' + str(log), flush=True)
    finally:
        file.write_bytes(original)
