"""Retain exact comparisons and a native-output byte mutant for the meter."""
import json
from pathlib import Path
import re
import shutil


def evidence(directory, node, native, comparator, adamic_sha, source_sha, run, *, stand_in=False):
    for label, sha in [('adamic_sha', adamic_sha), ('source_sha', source_sha)]:
        if not re.fullmatch('[0-9a-f]{40}', sha):
            raise ValueError(label + ' must be a full commit SHA')
    directory.mkdir(parents=True, exist_ok=False)
    # Keep outputs next to the reports so the meter can commit a self-contained run.
    shutil.copyfile(node, directory / 'node.stdout')
    shutil.copyfile(native, directory / 'native.stdout')
    for source, name in [(Path(node), 'node'), (Path(native), 'native')]:
        stderr = source.with_suffix('.stderr')
        if stderr.exists():
            shutil.copyfile(stderr, directory / (name + '.stderr'))

    def compare(right, name):
        code = run([comparator, directory / 'node.stdout', right], name, cwd=directory)
        report = json.loads((directory / (name + '.stdout')).read_text())
        if code not in (0, 1) or report['equal'] != (code == 0):
            raise RuntimeError(name + ': comparator failed')
        return code, report

    code, report = compare(directory / 'native.stdout', 'comparison')
    if code:
        raise RuntimeError('native output differs: ' + json.dumps(report['first_difference']))
    mutant = directory / 'native-byte-mutant.stdout'
    with (directory / 'native.stdout').open('rb') as source, mutant.open('wb') as target:
        first = source.read(1)
        if not first:
            raise RuntimeError('empty native output cannot provide a one-byte mutant')
        target.write(bytes([first[0] ^ 1]))
        shutil.copyfileobj(source, target)
    killed, changed = compare(mutant, 'mutant-comparison')
    expected = dict(offset=0, line=1, column=1, left=first[0], right=first[0] ^ 1)
    if killed != 1 or changed['first_difference'] != expected:
        raise RuntimeError('one-byte native mutant survived or wrong first difference')
    block = dict(adamic_sha=adamic_sha, source_sha=source_sha, run_directory=str(directory),
                 node_sha256=report['left']['sha256'], native_sha256=report['right']['sha256'],
                 comparison='comparison.stdout',
                 mutant=dict(output='native-byte-mutant.stdout', offset=0,
                             original_byte=first[0], changed_byte=first[0] ^ 1,
                             comparison='mutant-comparison.stdout', comparison_exit=killed,
                             first_difference=changed['first_difference']))
    result = dict(evidence={'scanner_native': block}, stand_in=stand_in,
                  milestone_eligible=not stand_in)
    (directory / 'scanner-native-evidence.json').write_text(json.dumps(result, indent=2) + '\n')
    return result
