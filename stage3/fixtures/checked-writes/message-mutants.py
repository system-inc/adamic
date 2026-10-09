"""Swap only diagnostic sides in the emitted JavaScript write-check runtime."""
import argparse
import hashlib
import json
from pathlib import Path
import tempfile

from verify import ROOT, checked_write_accepts, run

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--compiler', type=Path, required=True)
parser.add_argument('--compiler-revision', required=True)
args = parser.parse_args()
compiler = args.compiler.resolve()
manifest = json.loads((ROOT / 'manifest.json').read_text())
family = next(row for row in manifest['families'] if row['rank'] == 3)
name = family['files']['out']
expected = family['expected']['out']['adamic_ts']
with tempfile.TemporaryDirectory(prefix='checked-write-message-mutant-') as directory:
    scratch = Path(directory)
    source = scratch / name.replace('.a', '.ts')
    source.write_text((ROOT / name).read_text())
    emitted = run([str(compiler), 'js', str(source)])
    assert emitted['exit'] == 0, emitted
    original_path = scratch / 'original.mjs'
    original_path.write_text(emitted['stdout'])
    original = run(['node', str(ROOT.parents[2] / 'oracle/node.mjs'), str(original_path)])
    assert checked_write_accepts(original, expected), original
    before = 'if (!valid) panic("write failed: "'
    after = 'if (!valid) adamicMessageSideSwap("write failed: "'
    assert emitted['stdout'].count(before) == 1
    mutated = emitted['stdout'].replace(before, after, 1)
    # Keep the actual runtime rejection, value contracts and output intact.
    # Only the panic formatter reverses allocated-shape and incoming labels.
    helper = 'const adamicMessageSideSwap = message => panic(message.replace(/expects (.*), got (.*)$/, "expects $2, got $1"));\n'
    mutated = helper + mutated
    mutant_path = scratch / 'swapped.mjs'
    mutant_path.write_text(mutated)
    actual = run(['node', str(ROOT.parents[2] / 'oracle/node.mjs'), str(mutant_path)])
    assert (actual['exit'], actual['stdout']) == (original['exit'], original['stdout']), actual
    wanted_swap = expected['stderr'].replace('expects BinaryExpression, got BindingElement', 'expects BindingElement, got BinaryExpression')
    assert actual['stderr'] == wanted_swap, actual
    assert not checked_write_accepts(actual, expected), 'swapped diagnostic accepted'
report = dict(compiler_revision=args.compiler_revision,
              compiler_sha256=hashlib.sha256(compiler.read_bytes()).hexdigest(),
              fixture=name, mutation='swap allocated-shape and incoming sides in emitted JS panic formatter',
              original=original, mutant=actual, caught=True,
              caught_by='verify.checked_write_accepts exact stderr')
(ROOT / 'message-mutants.json').write_text(json.dumps(report, indent=2) + '\n')
print('CAUGHT ' + name + ': ' + actual['stderr'].strip())
print('catcher: exact stderr; exit 70 and before-write stdout unchanged')
