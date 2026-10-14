#!/usr/bin/env python3
"""Prove the lane product keys and artifact identity guards can fail."""
from pathlib import Path
import subprocess
import os

ROOT = Path(__file__).resolve().parents[2]


def failure(command, log, message):
    with log.open('w') as output:
        result = subprocess.run(command, cwd=ROOT, stdout=output,
                                stderr=subprocess.STDOUT, timeout=90,
                                env=dict(os.environ, PYTHONPATH=str(ROOT / 'stage3/lane')))
    print(f'{command!r}: exit={result.returncode}; {log}', flush=True)
    if result.returncode != 1 or message not in log.read_text():
        raise RuntimeError('mutant survived or failed for another reason')


def main():
    recipe = ROOT / 'stage3/lane/products.go'
    original = recipe.read_text()
    try:
        for index, phase in enumerate(('Install', 'Compiler', 'Harness')):
            marker = '\n\terr := filepath.WalkDir'
            assert marker in original
            recipe.write_text(original.replace(marker,
                f'\n\tif index == {index} {{ files = append(files[:1], files[2:]...) }}' + marker, 1))
            failure(['go', 'test', './stage3/lane', '-run', f'^TestStage3Lane{phase}TracksInputs$',
                     '-count=1', '-timeout', '90s', '-v'], Path('/tmp/hefhych-mutant-B' + phase + '.log'),
                    'changed lane worker reused product key')
    finally:
        recipe.write_text(original)
    artifact = ROOT / 'stage3/lane/artifact.py'
    original = artifact.read_text()
    probes = [
        ('upstream', '                  upstream_packages=upstream_packages_digest(root),\n',
         'test_upstream_installed_bytes_invalidate_key'),
        ('parser-location', "\n             and file.relative_to(stage).parts[:2] != ('api', 'node_modules')",
         'test_installed_parser_is_keyed_by_its_byte_pin'),
    ]
    try:
        for name, removed, test in probes:
            assert removed in original
            artifact.write_text(original.replace(removed, '', 1))
            failure(['python3', '-B', '-m', 'unittest', 'test_artifact.ArtifactTests.' + test],
                    Path('/tmp/hefhych-mutant-' + name + '.log'), 'AssertionError')
    finally:
        artifact.write_text(original)
    worker = ROOT / 'stage3/lane/products.py'
    original = worker.read_text()
    try:
        assert 'if actual != wanted:' in original
        worker.write_text(original.replace('if actual != wanted:', 'if False:', 1))
        failure(['python3', '-B', '-m', 'unittest',
                 'test_products.ProductPackageTests.test_changed_installed_bytes_are_refused_before_copy'],
                Path('/tmp/hefhych-mutant-package-pin.log'), 'ValueError not raised')
    finally:
        worker.write_text(original)


if __name__ == '__main__':
    main()
