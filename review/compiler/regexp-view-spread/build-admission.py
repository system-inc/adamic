"""Build the census from final sources, or from the recorded main lowering sources.

Run from the repository root with the setup environment sourced. The temporary
package is evidence tooling, not a registered compiler consumer in the product.
"""
import argparse
import json
from pathlib import Path
import subprocess
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument('--baseline', action='store_true')
parser.add_argument('--output', required=True)
args = parser.parse_args()
root = Path.cwd()
evidence = root / 'review/compiler/regexp-view-spread'
with tempfile.TemporaryDirectory(prefix='admission-build-', dir=evidence) as temporary:
    scratch = Path(temporary)
    source = scratch / 'main.go'
    source.write_text((evidence / 'admission.go.txt').read_text())
    command = ['go', 'build', '-o', args.output]
    if args.baseline:
        replacements = {}
        for name in ['lower.go', 'object.go']:
            original = scratch / (name + '.txt')
            original.write_bytes(subprocess.check_output([
                'git', 'show', 'd79302e50c022c997c00f7c68ddb7475f9ee65fc:internal/lower/' + name
            ], timeout=30))
            replacements[str(root / 'internal/lower' / name)] = str(original)
        empty = scratch / 'library_spread.go.txt'
        empty.write_text('package lower\n')
        replacements[str(root / 'internal/lower/library_spread.go')] = str(empty)
        overlay = scratch / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': replacements}))
        command += ['-overlay=' + str(overlay)]
    command.append(str(source))
    subprocess.run(command, check=True, timeout=90)
