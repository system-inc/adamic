# Run from the repository root after sourcing the printed cloud toolchain env.
# Every subprocess writes test output to a log; mutants use Go overlays.
from pathlib import Path
import json
import os
import subprocess

root = Path.cwd()
scratch = Path('/tmp/stable-emitter-reproduce')
scratch.mkdir(exist_ok=True)
probe = root / 'internal/native/unit_changes_probe_test.go'
assert not probe.exists(), 'remove the scratch measurement probe first'
probe.write_text((root / 'internal/native/stable_emitter_evidence/unit_changes_probe.go.txt').read_text())
mutants = [
    ('readability', 'internal/native/names.go', 'text = text[:20] + "_" + text[len(text)-39:]', 'text = text[:60]', './internal/native', '^TestStableIdentifierBoundsAndEscaping$', 'declaration spelling lost'),
    ('accessor-storage', 'internal/native/names.go', 'return "#accessor:" + e.functionName(function)', 'return fmt.Sprintf("#accessor:%d", function)', './internal/native', '^TestSourceNamesDoNotMove$', 'edit moved unrelated module-level code'),
    ('generated-role', 'internal/native/names.go', 'strconv.Quote(source.Role)', 'strconv.Quote("")', './internal/native', '^TestGeneratedRolesDoNotCollide$', 'generated role collided'),
    ('method-adapter', 'internal/native/emit_objects.go', 'e.thunks[name]', 'e.thunks[fmt.Sprint(function)]', './internal/native', '^TestMethodAdaptersShareStableDefinition$', 'duplicate helper adapters'),
    ('type-qualification', 'internal/lower/source_identity.go', 'key += fmt.Sprintf("@%q/%q", source.Module, source.Declaration)', 'key += fmt.Sprintf("@%q/%q", "", source.Declaration)', './internal/native', '^TestImportedSpecializationsHaveSourceIdentity$', 'imported same-spelling types collided'),
    ('temporary', 'internal/native/emit.go', '\te.temporaries = 0\n', '', './internal/native', 'TestFunctionCountersDoNotMove|TestReportLintUnitChanges', 'function-local edit changed'),
    ('cache', 'internal/native/emit.go', '\te.caches = 0\n', '', './internal/native', '^TestFunctionCountersDoNotMove$', 'cache moved'),
    ('region-facts', 'internal/native/emit.go', '\te.regionValues = map[string]bool{}\n', '', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regions.a$', 'heap-use-after-free'),
    ('ordinal', 'internal/native/emit_functions.go', 'name := e.namedFunction(function)', 'name := fmt.Sprintf("adamic_function_%d", function)', './internal/native', '^TestSourceNamesDoNotMove$', 'edit moved unrelated'),
    ('digest', 'internal/native/names.go', 'hex.EncodeToString(digest[:16])', 'hex.EncodeToString(digest[:0])', './internal/native', '^TestStableIdentifierBoundsAndEscaping$', 'escaping collision'),
    ('module-main', 'internal/native/emit.go', '\t\t\te.resetCounters()\n\t\t\treadable := module', '\t\t\treadable := module', './internal/native', '^TestSourceNamesDoNotMove$', 'edit moved unrelated module-level code'),
    ('regexp-token', 'internal/native/regexp.go', 'source := e.program.Regexps[index].Declarations', 'source := e.program.Regexps[index].Declarations\n source = strings.ReplaceAll(source, fmt.Sprintf("adamic_regex_%d", index), e.regexName(index))', './internal/native', '^TestRegexpSymbolsDoNotMove$', 'token rename damaged declaration'),
    ('absolute-root', 'internal/lower/source_identity.go', 'return filepath.ToSlash(relative)', 'return filepath.ToSlash(relative + "/" + root)', './internal/native', '^TestImportedSpecializationsHaveSourceIdentity$', 'program root relocation changed'),
    ('module-qualification', 'internal/native/names.go', 'func sourceKey(source ir.SourceIdentity) string {', 'func sourceKey(source ir.SourceIdentity) string {\n source.Module = ""', './internal/native', '^TestSourceNamesDoNotMove$', 'colliding declaration'),
]
try:
    for label, file, old, new, package, pattern, expected in mutants:
        original = root / file
        source = original.read_text()
        assert old in source, label
        replacement = scratch / (label + '.go')
        replacement.write_text(source.replace(old, new))
        overlay = scratch / (label + '.json')
        overlay.write_text(json.dumps({'Replace': {str(original): str(replacement)}}))
        log_path = scratch / (label + '.log')
        with log_path.open('w') as log:
            result = subprocess.run(['go', 'test', '-overlay=' + str(overlay), package, '-run', pattern, '-count=1', '-timeout', '30m', '-v'], stdout=log, stderr=subprocess.STDOUT, env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1', 'STABLE_NAMES_ACCEPTANCE': '1'})
        assert result.returncode and expected in log_path.read_text(), (label, result.returncode, log_path)
        print(label, 'caught', result.returncode, flush=True)
    with (scratch / 'units-after.log').open('w') as log:
        result = subprocess.run(['go', 'test', './internal/native', '-run', '^TestReportLintUnitChanges$', '-count=1', '-v'], stdout=log, stderr=subprocess.STDOUT, env={**os.environ, 'STABLE_NAMES_ACCEPTANCE': '1'})
    assert result.returncode == 0
    print((scratch / 'units-after.log').read_text())
finally:
    probe.unlink()

# Historical baseline: the saved units-before.log was measured at 52f9bae,
# before either naming or counter changes. Do not overlay those two old emitter
# files onto the new multi-file implementation: use that exact commit instead.
