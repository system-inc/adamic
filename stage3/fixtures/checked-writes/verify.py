"""Verify the selected corpus, source observations, and per-pair input mutants.

--require-checked-writes adds the .ts compatibility runtime gate. The authored
.a files retain the stricter .a refusal headers; only scratch copies use .ts.
"""
import argparse
import collections
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[2]


def run(command, cwd=REPO, environment=None):
    result = subprocess.run(command, cwd=cwd, env=environment, capture_output=True, text=True, timeout=120)
    return dict(exit=result.returncode, stdout=result.stdout, stderr=result.stderr)


def header_accepts(header, observation):
    match = re.fullmatch(r'// a-check: refused (.+)', header)
    return bool(match and observation['exit'] != 0 and
                ('Adamic 0.1 refuses ' + match[1] + '; ') in observation['stderr'])


def verify_observation(actual, expected, label):
    if actual != expected:
        raise AssertionError(f'{label}: expected {expected!r}, observed {actual!r}')


def checked_write_accepts(actual, expected):
    if (actual['exit'], actual['stdout']) != (expected['exit'], expected['stdout']):
        return False
    if expected['exit'] == 0:
        return actual['stderr'] == ''
    if 'stderr' in expected:
        return actual['stderr'] == expected['stderr']
    return (actual['stderr'].startswith('adamic: panic: ') and
            all(word in actual['stderr'] for word in expected['stderr_fields']))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--compiler', type=Path, required=True, help='built Adamic CLI')
    parser.add_argument('--compiler-revision', help='source revision of the supplied compiler, recorded as provenance')
    parser.add_argument('--survey', type=Path, help='results.json from survey commit 9b8ebd77')
    parser.add_argument('--observe', action='store_true', help='refresh local evidence and counts')
    parser.add_argument('--require-checked-writes', action='store_true', help='fail until every .ts native/JS contract passes')
    arguments = parser.parse_args()
    compiler = arguments.compiler.resolve()
    manifest = json.loads((ROOT / 'manifest.json').read_text())
    families = manifest['families']
    assert len(families) == 20 and len({r['family'] for r in families}) == 20
    assert families == sorted(families, key=lambda r: (-r['count'], r['family']))
    assert sum(r['count'] for r in families) == manifest['selected_sites'] == 381
    assert len({i for r in families for i in r['site_ids']}) == 381
    names = {name for row in families for name in row['files'].values()}
    assert names == {file.name for file in ROOT.glob('*.a')} and len(names) == 40
    if arguments.survey:
        raw = arguments.survey.read_bytes()
        assert hashlib.sha256(raw).hexdigest() == manifest['survey_sha256'], 'survey pin changed'
        survey = json.loads(raw)
        records = [r for r in survey['records'] if r['classification'] == 'd']
        assert len(records) == manifest['population'] == 515
        ranked = sorted(collections.Counter(r['source_family'] for r in records).items(), key=lambda item: (-item[1], item[0]))[:20]
        assert [(r['family'], r['count']) for r in families] == ranked
        for row in families:
            selected = [r for r in records if r['source_family'] == row['family']]
            assert row['site_ids'] == [r['id'] for r in selected]
            assert row['representative'] == selected[0]
        print('PASS survey: all 515 class-d records grouped; top 20 cover 381 sites')

    observations = []
    mutations = []
    failures = []
    with tempfile.TemporaryDirectory(prefix='checked-writes-') as scratch_name:
        scratch = Path(scratch_name)
        for family in families:
            positive_source = None
            for side, name in family['files'].items():
                source = (ROOT / name).read_text()
                header = source.splitlines()[0]
                assert header == family['a_check'][side], name
                node = run(['node', '--disable-warning=ExperimentalWarning', str(REPO / 'oracle/node.mjs'), str(ROOT / name)])
                verify_observation(node, family['expected'][side]['node'], name + ' Node')
                if side == 'out':
                    # Node's silent-store result models a missing inserted check.
                    # This proves the acceptance predicate, not a compiler mutant.
                    assert not checked_write_accepts(node, family['expected'][side]['adamic_ts']), name
                a_check = run([str(compiler), 'c', str(ROOT / name)])
                assert header_accepts(header, a_check), (name, a_check)
                # The exact header predicate must reject a missing or false reason.
                assert not header_accepts('', a_check), name
                assert not header_accepts('// a-check: refused missing-check-mutant', a_check), name
                ts_source = scratch / name.replace('.a', '.ts')
                ts_source.write_text(source)
                executable = scratch / name.replace('.a', '')
                build = run([str(compiler), 'build', str(ts_source), '-o', str(executable), '--sanitize'])
                row = dict(file=name, source_sha256=hashlib.sha256(source.encode()).hexdigest(), node=node,
                           a_check=a_check, ts_build=build, ts_runtime=None, ts_javascript=None, ts_counted=None, native_counts=None, checked_write_pass=False)
                if build['exit'] == 0:
                    environment = dict(os.environ)
                    # Panic intentionally ends without cleanup. Finished positives
                    # must be leak-clean; neither mode may hide an ASan/UBSan fault.
                    environment['ASAN_OPTIONS'] = 'detect_leaks=' + ('1' if side == 'in' else '0')
                    environment['UBSAN_OPTIONS'] = 'halt_on_error=1'
                    native = run([str(executable)], environment=environment)
                    row['ts_runtime'] = native
                    counted_binary = Path(str(executable) + '.counted')
                    counted_build = run([str(compiler), 'build', str(ts_source), '-o', str(counted_binary), '--count'])
                    assert counted_build['exit'] == 0, counted_build
                    counted = run([str(counted_binary)])
                    pattern = r'adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n$'
                    counts = re.search(pattern, counted['stderr'])
                    assert counts, counted
                    keys = ['allocations', 'frees', 'retains', 'releases', 'peak', 'regions']
                    row['native_counts'] = dict(zip(keys, map(int, counts.groups())))
                    row['ts_counted'] = counted
                    counted_result = dict(counted, stderr=counted['stderr'][:counts.start()])
                    if side == 'in':
                        values = row['native_counts']
                        assert values['allocations'] == values['frees'] + values['regions'], values
                    js = run([str(compiler), 'js', str(ts_source)])
                    if js['exit'] == 0:
                        js_path = scratch / (name + '.mjs')
                        js_path.write_text(js['stdout'])
                        row['ts_javascript'] = run(['node', '--disable-warning=ExperimentalWarning', str(REPO / 'oracle/node.mjs'), str(js_path)])
                    else:
                        row['ts_javascript'] = js
                    expected = family['expected'][side]['adamic_ts']
                    row['checked_write_pass'] = (checked_write_accepts(native, expected) and
                                                 checked_write_accepts(row['ts_javascript'], expected) and
                                                 checked_write_accepts(counted_result, expected))
                if not row['checked_write_pass']:
                    failures.append(name)
                observations.append(row)
                if side == 'in':
                    positive_source = source
                print(f'PASS {name}: Node exact; .a header exact; .ts checked-write={row["checked_write_pass"]}')
            mutant = family['mutant']
            assert positive_source.count(mutant['before']) == 1, mutant
            mutated = positive_source.replace(mutant['before'], mutant['after'], 1)
            mutant_path = scratch / mutant['file']
            mutant_path.write_text(mutated)
            actual = run(['node', '--disable-warning=ExperimentalWarning', str(REPO / 'oracle/node.mjs'), str(mutant_path)])
            # The fitting-input expectation must fail only on stdout: Node still
            # finishes silently, with no exception or unrelated checker failure.
            assert actual['exit'] == 0 and actual['stderr'] == '', actual
            expected = family['expected']['in']['node']
            assert actual['stdout'] != expected['stdout'], mutant
            verify_observation(actual, family['expected']['out']['node'], mutant['file'] + ' mutant control')
            mutations.append(dict(rank=family['rank'], family=family['family'], mutation=mutant,
                                  observed=actual, caught_by='exact fitting-fixture Node stdout', caught=True))
            print(f'CAUGHT {family["rank"]:02}: {family["family"]} input mutant by exact Node stdout')
    report = dict(fixture_base_commit='45487a809f89885a3fc651cd590e7dabf31362dc',
                  compiler_revision=arguments.compiler_revision,
                  compiler_sha256=hashlib.sha256(compiler.read_bytes()).hexdigest(),
                  node_version=subprocess.check_output(['node', '--version'], text=True).strip(),
                  fixtures=observations, mutants=mutations,
                  totals=dict(node_pass=40, a_headers_pass=40, header_predicate_mutants_caught=80,
                              input_mutants_caught=20, checked_write_pass=40-len(failures),
                              checked_write_pending=len(failures)),
                  checked_write_pending=failures)
    if arguments.observe:
        (ROOT / 'observations.json').write_text(json.dumps(report, indent=2) + '\n')
        lines = ['# Counts', '', 'Local stage-3 corpus, not registered in internal/oracle.', '',
                 '| Family | Class-d sites | Node fixtures passing | Input mutants caught | .ts runtime contracts passing |',
                 '|---|---:|---:|---:|---:|']
        for row in families:
            passing = sum(r['checked_write_pass'] for r in observations if r['file'] in row['files'].values())
            family_label = row["family"].replace("|", "\\|")
            lines.append(f'| {family_label} | {row["count"]} | 2 | 1 | {passing} |')
        lines += [f'| **Total** | **381 / 515** | **40** | **20** | **{40-len(failures)} / 40** |', '',
                  'Allocation rows below are release counted builds; panic rows stop at exit 70.',
                  'Finished positives also pass ASan/UBSan and leak detection.',
                  'No internal/oracle registry or counts rows were added by this unit.', '']
        lines += ['| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |',
                  '|---|---:|---:|---:|---:|---:|---:|']
        for row in observations:
            if row['native_counts'] is not None:
                values = row['native_counts']
                lines.append('| ' + row['file'] + ' | ' + ' | '.join(str(values[k]) for k in ['allocations', 'frees', 'retains', 'releases', 'peak', 'regions']) + ' |')
        lines += ['', 'Uncompiled compatibility inputs have no native allocation row.', '']
        (ROOT / 'counts.md').write_text('\n'.join(lines))
    print('TOTALS ' + json.dumps(report['totals'], sort_keys=True))
    if arguments.require_checked_writes and failures:
        raise SystemExit('checked-write runtime gate pending: ' + ', '.join(failures))


if __name__ == '__main__':
    main()
