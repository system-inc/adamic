#!/usr/bin/env python3
"""Restore the single-root census bug without modifying the working tree."""
import json
import pathlib
import subprocess
import tempfile

root = pathlib.Path(__file__).resolve().parents[4]
source = root / 'internal/load/load.go'
text = source.read_text()
start = text.index('\t\t// Composite membership')
end = text.index('\n\t}\n\troots = append(roots, preludePath)', start)
with tempfile.TemporaryDirectory(prefix='adamic-census-mutant-') as directory:
    mutant = pathlib.Path(directory) / 'load.go'
    mutant.write_text(text[:start] + '\t\troots = userRoots\n' + text[end:])
    overlay = pathlib.Path(directory) / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {str(source): str(mutant)}}))
    log = pathlib.Path('/tmp/load-census-mutant.log')
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', '-overlay', str(overlay), './internal/load', '-run', '^TestCompositeProjectCensus$', '-count=1'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    if result.returncode != 1 or 'error TS6307' not in log.read_text():
        raise SystemExit('mutant did not fail with TS6307; inspect ' + str(log))
    print('single-root mutant caught by TS6307; ' + str(log))
