"""Audit and publish explicitly scoped directory census results, never a full claim.

Usage: partial.py ADAPTED_COMPILER_DIRECTORY RUN_DIRECTORY OUTPUT_DIRECTORY BINARY
RUN_DIRECTORY contains directory/speculative.jsonl from four foreground runs.
Each directory is a separate checker project using the original source bytes.
"""
from collections import Counter, defaultdict
from copy import deepcopy
import hashlib
import html
import json
import os
from pathlib import Path
import subprocess
import sys

root, runs, output, binary = (Path(p).resolve() for p in sys.argv[1:5])
scopes = ('_namespaces', 'factory', 'transformers/declarations', 'transformers/module')
scripts = Path(__file__).resolve().parent
output.mkdir(parents=True, exist_ok=True)


def run(command, log, **flags):
    env = dict(os.environ)
    for name in ('LATENT_SPECULATIVE', 'LATENT_MUTANT_NO_STUBS', 'LATENT_MUTANT_DEPTH_ZERO'):
        env.pop(name, None)
    env.update(flags)
    with log.open('w') as stream:
        subprocess.run(command, env=env, stdout=stream, stderr=subprocess.STDOUT, check=True)


projects = []
unique = {}
files = []
for scope in scopes:
    source = root / scope
    folder = runs / scope
    destination = output / 'directories' / scope
    raw = folder / 'speculative.jsonl'
    assert raw.exists(), scope
    rows = [json.loads(line) for line in raw.read_text().splitlines()]
    expected = {str(p) for p in source.rglob('*') if p.is_file() and p.suffix in ('.ts', '.a')}
    assert {row['file'] for row in rows[1:]} == expected, (scope, 'unfinished run')
    stock = folder / 'stock.json'
    run(['node', str(scripts / 'stock.cjs'), str(source), str(raw), str(stock)], folder / 'stock.log')
    run([sys.executable, str(scripts / 'report.py'), str(source), str(raw), str(stock), str(destination)], folder / 'report.log')
    run([sys.executable, str(scripts / 'verify.py'), str(raw), str(destination / 'RESULT.json'), str(source)], folder / 'verify.log')
    result = json.loads((destination / 'RESULT.json').read_text())
    assert result['independent_depth_unmatched'] == 0, scope
    for mode, flags in [('full', {}), ('no-stubs', {'LATENT_SPECULATIVE': '1', 'LATENT_MUTANT_NO_STUBS': '1'})]:
        run([str(binary), str(source), str(folder / (mode + '.jsonl'))], folder / (mode + '.log'),
            LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1', **flags)
    baseline = (folder / 'full.jsonl').read_bytes()
    assert baseline == (folder / 'no-stubs.jsonl').read_bytes(), (scope, 'no-stubs changed baseline')
    full = [json.loads(line) for line in baseline.splitlines()]
    assert full[0]['diagnostics'] == rows[0]['diagnostics'], scope
    assert full[0]['diagnostic_sites'] == rows[0]['diagnostic_sites'], scope
    baseline_sites = {tuple(f[k] for k in ('kind', 'where', 'reason', 'text'))
                      for row in full[1:] for f in row['findings']
                      if f['kind'] in ('NotYet', 'Refused') and f['where'].startswith(str(source) + '/')}
    local = {}
    for row in rows[1:]:
        for finding in row['findings']:
            if finding['kind'] not in ('NotYet', 'Refused') or not finding['where'].startswith(str(source) + '/'):
                continue
            identity = tuple(finding[k] for k in ('kind', 'where', 'reason', 'text', 'site_where', 'site_kind')) + (finding.get('site_start', 0), finding['site_end'])
            previous = unique.setdefault(identity, finding)
            assert previous['depth'] == finding['depth'], identity
            local[identity] = finding
    assert len(local) >= len(baseline_sites), scope
    if local:
        assert len(local) > len(baseline_sites), scope
    for row in result['files']:
        row = dict(row, file=scope + '/' + row['file'])
        files.append(row)
    projects.append(dict(directory=scope, files=len(result['files']), counts=result['counts'],
                         full_sites=len(baseline_sites), speculative_sites=len(local),
                         no_stubs_byte_identical=True, checker_diagnostics_identical=True,
                         independent_depth_matches=result['independent_depth_matches'],
                         independent_depth_unmatched=result['independent_depth_unmatched'],
                         outside_scope_attempts=result['independent_depth_without_source']))
    print(f"PASS {scope}: {len(local)} speculative sites versus {len(baseline_sites)} full; stock ancestry and applicable artifact mutants passed; no-stubs byte-identical", flush=True)

counts = {kind: [sum(f['kind'] == kind and min(4, f['depth']) == d for f in unique.values()) for d in range(5)]
          for kind in ('NotYet', 'Refused')}
reasons = defaultdict(Counter)
for finding in unique.values():
    reasons[(finding['kind'], finding['root_kind'])][min(4, finding['depth'])] += 1
ranking = sorted(reasons, key=lambda key: (-sum(reasons[key].values()), key))[:20]
covered = {row['file'] for row in files}
all_files = {str(p.relative_to(root)): p for p in root.rglob('*') if p.is_file()}
source_bytes = sum(p.stat().st_size for p in all_files.values())
typescript_bytes = sum(p.stat().st_size for p in all_files.values() if p.suffix in ('.ts', '.a'))
examined = sum(all_files[name].stat().st_size for name in covered)
result = dict(status='partial; four separately loaded directory census projects, not a whole-compiler run',
              base='a5630a90d05abe85670ba1e00bff3643a2742e53',
              mapper_fix='69501280a81259fb512edbb8dd0e52c6eb0d88c8',
              upstream='050880ce59e30b356b686bd3144efe24f875ebc8',
              covered_directories=list(scopes), projects=projects, counts=counts,
              top20=[dict(kind=kind, root_kind=reason, counts=[reasons[(kind, reason)][d] for d in range(5)]) for kind, reason in ranking],
              files=sorted(files, key=lambda row: row['file']), source_bytes=source_bytes,
              typescript_source_bytes=typescript_bytes, examined_bytes=examined,
              examined_share=examined/source_bytes, typescript_examined_share=examined/typescript_bytes,
              unexamined=[dict(file=name, bytes=p.stat().st_size,
                              reason='outside covered directories; full census interrupted before buffered observations were emitted'
                              if p.suffix in ('.ts', '.a') else 'non-TypeScript input; no lowering sites')
                          for name, p in sorted(all_files.items()) if name not in covered])


def verify(artifact):
    # Independent arithmetic over raw observations and the original on-disk inventory.
    assert artifact['covered_directories'] == list(scopes)
    assert {row['file'] for row in artifact['files']} == {str(p.relative_to(root)) for scope in scopes for p in (root/scope).rglob('*') if p.is_file() and p.suffix in ('.ts', '.a')}
    assert all(row['unvisited_nodes'] == 0 for row in artifact['files'])
    for row in artifact['files']:
        data = all_files[row['file']].read_bytes()
        assert row['source_bytes'] == len(data)
        assert row['sha256'] == hashlib.sha256(data).hexdigest()
    for kind in ('NotYet', 'Refused'):
        expected = Counter(min(4, f['depth']) for f in unique.values() if f['kind'] == kind)
        assert artifact['counts'][kind] == [expected[d] for d in range(5)]
    expected_reasons = Counter((f['kind'], f['reason']) for f in unique.values())
    expected_ranking = sorted(expected_reasons, key=lambda key: (-expected_reasons[key], key))[:20]
    assert [(row['kind'], row['root_kind']) for row in artifact['top20']] == expected_ranking
    for row in artifact['top20']:
        assert row['counts'] == [sum(f['kind'] == row['kind'] and f['reason'] == row['root_kind'] and min(4, f['depth']) == d for f in unique.values()) for d in range(5)]
    assert artifact['examined_bytes'] == sum(all_files[row['file']].stat().st_size for row in artifact['files'])
    assert artifact['source_bytes'] == sum(p.stat().st_size for p in all_files.values())
    assert artifact['examined_share'] == artifact['examined_bytes']/artifact['source_bytes']
    assert artifact['typescript_examined_share'] == artifact['examined_bytes']/typescript_bytes
    assert {row['file'] for row in artifact['unexamined']} == set(all_files) - covered


verify(result)
for name, mutate in [
    ('headline-depth-total', lambda x: x['counts']['NotYet'].__setitem__(0, x['counts']['NotYet'][0] + 1)),
    ('top20-depth-count', lambda x: x['top20'][0]['counts'].__setitem__(0, x['top20'][0]['counts'][0] + 1)),
    ('omit-covered-file', lambda x: x['files'].pop()),
    ('byte-share', lambda x: x.__setitem__('examined_bytes', x['examined_bytes'] + 1)),
    ('omit-unexamined-file', lambda x: x['unexamined'].pop()),
    ('claim-additional-directory', lambda x: x['covered_directories'].append('checker.ts')),
]:
    mutant = deepcopy(result)
    mutate(mutant)
    try:
        verify(mutant)
    except AssertionError:
        print(name + ' partial artifact mutant caught', flush=True)
    else:
        raise AssertionError(name + ' mutant survived')

(output/'RESULT.json').write_text(json.dumps(result, indent=2) + '\n')
text = '''Published audited partial speculative depth tables for four compiler directories.
Base `a5630a90` includes `ed6e2975` and mapper fix `69501280`; production compiler files are unchanged.
Coverage: `_namespaces/`, `factory/`, `transformers/declarations/`, `transformers/module/`, each loaded separately.
Stock TypeScript ancestry, exact depth/top-20 recount and no-stubs baselines pass for each directory.
The whole-compiler run was interrupted in `checker.ts`; remaining files are explicitly unexamined below.

'''
text += '| Depth | NotYet | Refused |\n|---|---:|---:|\n'
for d in range(5):
    text += f'| {"4+" if d == 4 else d} | {counts["NotYet"][d]:,} | {counts["Refused"][d]:,} |\n'
text += '| Total | ' + ' | '.join(f'{sum(counts[k]):,}' for k in ('NotYet', 'Refused')) + ' |\n\n'
text += '| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |\n|---|---|---:|---:|---:|---:|---:|---:|\n'
for row in result['top20']:
    reason = html.escape(row['root_kind'], quote=False).replace('|', '&#124;').replace('\n', '<br>')
    text += f'| {row["kind"]} | {reason} | ' + ' | '.join(f'{n:,}' for n in row['counts'] + [sum(row['counts'])]) + ' |\n'
text += '\n| Covered directory | Files | Speculative sites | Full/no-stubs sites |\n|---|---:|---:|---:|\n'
for project in projects:
    text += f'| `{project["directory"]}/` | {project["files"]} | {project["speculative_sites"]:,} | {project["full_sites"]:,} |\n'
text += f'\nExamined **{examined:,} / {typescript_bytes:,} TypeScript bytes ({100*examined/typescript_bytes:.6f}%)**, in **{len(files)} / 79 files**. Including non-TypeScript files: **{examined:,} / {source_bytes:,} bytes ({100*examined/source_bytes:.6f}%)**. All covered AST children were visited. Comments and whitespace are included in parsed source bytes; type and structural nodes are traversed rather than passed to a value lowerer. This is speculative examination, not successful production lowering.\n'
text += '\nThese are four independent directory-root projects on the original adapted bytes. Imports remain available to the checker, but registration and the census walk use each directory\'s roots. Isolated binding context can change observed reasons and depths compared with a whole-compiler project; these counts must not be presented as the final full-corpus totals. Attempts attributed outside a project directory are excluded from its site tables.\n'
text += '\nThe previous whole-compiler run stopped during checker.ts without emitting its buffered records. Completed progress lines are not depth evidence and are not included. The earlier PARTIAL.json and legacy raw evidence from e46fa73d have unverified depth tags and are superseded only for these measured directory projects.\n'
text += '\nUnexamined files, including all remaining TypeScript sources:\n\n'
for row in result['unexamined']:
    text += f'- `{row["file"]}`: {row["bytes"]:,} bytes; {row["reason"]}.\n'
text += '\nNormal C and JS output with the overlay off remains byte-identical to base a5630a90 on the isolation witness. Nested controls prove depth 1 and 2, checker-clean controls reach depth 3 and 4, and the equal-span witness distinguishes identifier/parameter ancestry. All six .a controls pass local a-check with zero mismatches; eight header mutants are caught. The original 13:00 MDT deadline was missed.\n'
text += '\nSee [validation and mutants](validation.md), [method](method.md), [fixture counts](counts.md), [header validation](header-validation.md), and [evidence manifest](evidence/manifest.json). Reproduce this partial report with `partial.py ADAPTED_COMPILER_DIRECTORY RUN_DIRECTORY OUTPUT_DIRECTORY BINARY`, after running the four directory speculative measurements. Per-directory depth tables are in `directories/`; raw observations and stock ancestry are retained under `evidence/directories/`.\n'
(output/'README.md').write_text(text)
print(f'PASS: partial combined report, {len(unique)} sites, {len(files)} files; six partial artifact mutants caught', flush=True)
