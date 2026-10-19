"""Build independent scratch census mutants while preserving delivered sources."""
import json
from pathlib import Path
import subprocess

root = Path.cwd()
evidence = root / 'review/compiler/loom-census-deferral'
overlay = Path('/tmp/loom-census-overlay')
original = json.loads((overlay / 'overlay.json').read_text())
mutants = {
    'no-caller': [('internal/lower/latent_hook.go', 'unit.Instantiated = latentInstantiated[unit.Where]', 'unit.Instantiated = true')],
    'plain-map': [('internal/lower/latent_full.go', 'latentCopyIR(function.OptionalParameters, seen)', 'copyValue(reflect.ValueOf(function.OptionalParameters)).Interface().(map[int]bool)')],
    'premature': [
        ('internal/lower/latent_hook.go', 'if node.Kind == ast.KindFunctionDeclaration && len(node.TypeParameters()) > 0 {', 'if false && node.Kind == ast.KindFunctionDeclaration && len(node.TypeParameters()) > 0 {'),
        ('internal/lower/latent_units.go', 'if deferred {', 'if false && deferred {'),
    ],
}
for name, edits in mutants.items():
    folder = evidence / ('mutant-' + name)
    folder.mkdir(exist_ok=True)
    configuration = {'Replace': dict(original['Replace'])}
    for source, old, new in edits:
        key = str(root / source)
        text = Path(configuration['Replace'][key]).read_text()
        assert text.count(old) == 1, (name, source, old, text.count(old))
        path = folder / (Path(source).stem + '.go.txt')
        path.write_text(text.replace(old, new))
        configuration['Replace'][key] = str(path)
    manifest = folder / 'overlay.json'
    manifest.write_text(json.dumps(configuration, indent=2) + '\n')
    with (folder / 'build.log').open('w') as log:
        subprocess.run(['go', 'build', '-overlay=' + str(manifest), '-o', '/tmp/loom-census-' + name,
                        './stage3/census/latent/tool'], stdout=log, stderr=subprocess.STDOUT,
                       timeout=90, check=True)
    print(name + ': built')
