"""Actual corpus metadata mutants must fail the independent per-site audit."""
import gzip
import importlib.util
import json
import pathlib
import sys
import traceback

base = pathlib.Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location('artifacts', base / 'dynamic-keys/artifacts.py')
artifacts = importlib.util.module_from_spec(spec)
spec.loader.exec_module(artifacts)
with gzip.open(sys.argv[1], 'rt') as stream:
    value = artifacts.decode(json.load(stream))
mapped = json.loads(pathlib.Path(sys.argv[2]).read_text())
root = pathlib.Path(sys.argv[3])
original_sites = value['sites']
original_counts = value['counts']


def caught(name, expected_line):
    try:
        artifacts.audit(value, mapped, root)
    except AssertionError as failure:
        frame = traceback.extract_tb(failure.__traceback__)[-1]
        assert frame.filename.endswith('/latent/audit.py') and expected_line in frame.line, (name, frame)
        print(name + ': valid corpus metadata; independent audit catches ' + frame.line.strip(), flush=True)
    else:
        raise AssertionError(name + ' escaped')


value['sites'] = original_sites[:-1]
caught('drop-corpus-site', "len(result['sites'])==len(mapped)")
value['sites'] = original_sites[:-1] + [original_sites[0]]
caught('duplicate-corpus-site', 'assert key not in seen')
value['sites'] = original_sites
value['counts'] = json.loads(json.dumps(original_counts))
value['counts']['tagged']['conforms and ready (free)'] += 1
value['counts']['tagged']['unknown'] -= 1
caught('invent-aggregate-free', "assert counts==result['counts']")
value['counts'] = original_counts
print('PASS three actual-corpus audit mutants; original artifact was not modified', flush=True)
