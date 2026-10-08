"""Prepare a census-only loader overlay; never modify compiler files on disk."""
import json
import pathlib
import sys

repository = pathlib.Path(__file__).resolve().parents[2]
scratch = pathlib.Path(sys.argv[1]).resolve()
scratch.mkdir(parents=True, exist_ok=True)
original = (repository / 'internal/load/load.go').read_text()
if len(sys.argv) > 2 and sys.argv[2] == 'policy-probes':
    enabled = 'ErasableSyntaxOnly:         core.TSTrue,'
    disabled = 'ErasableSyntaxOnly:         core.TSFalse,'
    assert original.count(enabled) + original.count(disabled) == 1
    # This historical counterfactual is now the production setting too.
    (scratch / 'load.go').write_text(original.replace(enabled, disabled))
    (scratch / 'overlay.json').write_text(json.dumps({'Replace': {str(repository / 'internal/load/load.go'): str(scratch / 'load.go')}}, indent=2) + '\n')
    sys.exit(0)
start = original.index('func compilerOptions() *core.CompilerOptions {')
end = original.index('\n// Load checks', start)
modified = original[:start] + 'func compilerOptions() *core.CompilerOptions { return censusOptions() }\n' + original[end:]
assert modified.count('roots = append(roots, preludePath)') == 1
modified = modified.replace('roots = append(roots, preludePath)', 'ownedRoots := append([]string{}, roots...)\n\troots = censusRoots() // All upstream project roots, even for a single entry.')
assert modified.count('roots[:len(roots)-1]') == 1
modified = modified.replace('roots[:len(roots)-1]', 'ownedRoots')
if len(sys.argv) > 2 and sys.argv[2] == 'stock-library':
    needle = 'cachedvfs.From(&regexpLibraryFS{FS: bundled.WrapFS(fs)})'
    assert modified.count(needle) == 1
    modified = modified.replace(needle, 'cachedvfs.From(bundled.WrapFS(fs))')
(scratch / 'load.go').write_text(modified)
(scratch / 'overlay.json').write_text(json.dumps({'Replace': {
 str(repository / 'internal/load/load.go'): str(scratch / 'load.go'),
 str(repository / 'internal/load/census_config_hook.go'): str(repository / 'stage3/census/config_hook.go.txt'),
}}, indent=2) + '\n')
