#!/usr/bin/env python3
"""Prove each CSS input cache component and validation check can fail."""
from pathlib import Path
import json
import os
import subprocess
import tempfile

cloud = Path(__file__).resolve().parent
source = (cloud / 'setup-gate-inputs.py').read_text()
mutants = {}
for name, text in [('commit', 'commit=CSS_COMMIT, '), ('url', 'url=CSS_URL,\n                     '),
                   ('sparse', 'sparse=CSS_SPARSE, '), ('counts', 'counts=CSS_COUNTS, '), ('helper', 'helper=source_hash()')]:
    # Restrict key mutations to the new fixture key, leaving other input keys intact.
    start = source.index('def css_fixture_key():');end = source.index('\ndef css_counts', start)
    body = source[start:end]
    changed = body.replace(text, '' if name != 'helper' else "helper='constant'")
    assert changed != body, name
    mutants['drop-' + name] = (source[:start] + changed + source[end:], 'test_every_fixture_key_component')
mutants['degraded-counts'] = (source.replace('if counts != CSS_COUNTS:', 'if False:'), 'test_counts_reject_degraded_corpus')
mutants['wrong-head'] = (source.replace('if actual != CSS_COMMIT:', 'if False:'), 'test_checkout_commit_and_sparse_checks')
mutants['wrong-sparse'] = (source.replace('if sorted(sparse) != sorted(CSS_SPARSE):', 'if False:'), 'test_checkout_commit_and_sparse_checks')
mutants['skip-checkout-bytes'] = (source.replace("saved['key'] == key and saved['tree'] == artifact_digest(destination)", "saved['key'] == key"), 'test_local_sparse_cache_repair_and_uncached_equality')
mutants['skip-prefix-prettier-version'] = (source.replace('if (actual !== expected) throw Error(`${name}:', 'if (false) throw Error(`${name}:'), 'test_shared_prettier_version_is_verified')
mutants['skip-package-prettier-version'] = (source.replace('if (actual !== expected) throw Error(`ADAMIC_MARKDOWNINLINE_LIBRARY:', 'if (false) throw Error(`ADAMIC_MARKDOWNINLINE_LIBRARY:'), 'test_shared_prettier_version_is_verified')
for name in ['ADAMIC_CSS_FIXTURES', 'ADAMIC_CSSNUMBERS_LIBRARY', 'ADAMIC_CSSSTRINGS_LIBRARY', 'ADAMIC_MARKDOWNINLINE_LIBRARY']:
    mutants['missing-' + name] = (source.replace("'" + name + "':", "'OMITTED_" + name + "':"), 'test_exports_and_ordinary_unset')
report = cloud / 'reports/css-gate-inputs';results = []
with tempfile.TemporaryDirectory(prefix='css-input-mutants-', dir='/tmp/adamic-gate') as temporary:
    scratch = Path(temporary)
    (scratch / 'setup-gate-npm.py').symlink_to(cloud / 'setup-gate-npm.py')
    for name, (code, test) in mutants.items():
        assert code != source, name
        file = scratch / (name + '.py');file.write_text(code)
        with (report / ('mutant-' + name + '.log')).open('w') as log:
            result = subprocess.run(['python3', '-m', 'unittest', 'test_css_gate_inputs.CSSInputs.' + test],
                                    cwd=cloud, env=dict(os.environ, ADAMIC_GATE_INPUTS_MODULE=str(file)),
                                    stdout=log, stderr=subprocess.STDOUT)
        output = (report / ('mutant-' + name + '.log')).read_text()
        assert result.returncode and ('FAIL:' in output or 'ERROR:' in output), name
        assert 'SyntaxError' not in output, name
        results.append(dict(mutant=name, test=test, exit=result.returncode));print(name, 'caught', flush=True)
(report / 'mutants.json').write_text(json.dumps(results, indent=2) + '\n')
