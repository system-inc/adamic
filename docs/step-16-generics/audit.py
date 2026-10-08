"""Select exact generic census reasons and count distinct attempted roots."""
import argparse
from collections import defaultdict
import gzip
import hashlib
import json
from pathlib import Path
import re
import subprocess
import unittest

# These names come from TypeParameterDeclaration nodes in the pinned tsc AST.
# Names also used by concrete compiler types are kept in the collision appendix.
COLLISIONS = {'Child', 'Children', 'Entry', 'Resolution', 'ResolutionCache', 'Source', 'SourceFile', 'Target'}

def scoped(reason, names):
    if MUTANT_SELECTION:
        return True
    if re.search(r'generic|polymorphic|type parameter', reason):
        return True
    if reason.startswith('reading '):
        return False  # A lexical name is not a type-parameter diagnostic.
    return any(re.search(r'\b' + re.escape(name) + r'\b', reason) for name in set(names) - COLLISIONS)

def table_fields(line):
    escaped = line if MUTANT_MARKDOWN else line.replace(chr(92) + '|', 'PIPE_ESCAPE')
    return [part.strip().replace('PIPE_ESCAPE', '|') for part in escaped.split('|')[1:-1]]

def three_witnesses(sites):
    result = []
    for site in sites:
        parts = site.rsplit(':', 2)
        witness = ':'.join(parts[:2]) if len(parts) == 3 and parts[-1].isdigit() and parts[-2].isdigit() else site
        if witness not in result:
            result.append(witness)
    return result[:3]

def summarize(findings, mutant=False):
    groups = defaultdict(list)
    for finding in findings:
        if finding['kind'] in ('Refused', 'NotYet'):
            groups[(finding['kind'], finding['reason'])].append(finding)
    result = []
    for (kind, reason), rows in sorted(groups.items()):
        roots = {row['unit'] for row in rows}
        sites = sorted({row['where'] for row in rows})
        result.append(dict(kind=kind, reason=reason, roots=len(rows) if mutant else len(roots),
                           sites=len(sites), witnesses=three_witnesses(sites)))
    return result

def bound_collision(reason, where, scopes):
    parts = where.rsplit(':', 2)
    if len(parts) != 3 or not parts[-1].isdigit() or not parts[-2].isdigit():
        return False
    filename, line, column = parts
    filename = filename.split('/src/compiler/', 1)[-1]
    position = (int(line), int(column))
    return any(tuple(r['start']) <= position <= tuple(r['end']) and any(
        name in COLLISIONS and re.search(r'\b' + re.escape(name) + r'\b', reason)
        for name in r['names']) for r in scopes.get(filename, []))

class AuditTests(unittest.TestCase):
    def test_refusal_table_union(self):
        fields = table_fields('| a function returning T ' + chr(92) + '| undefined | owner | 2 | 3 | -1 | core | a:1 | a:2 | adaptation | proof |')
        self.assertEqual(len(fields), 10)
        self.assertEqual(fields[0], 'a function returning T | undefined')
        self.assertEqual(fields[2], '2')
    def test_roots_are_not_attempts_or_sites(self):
        rows = [dict(kind='NotYet', reason='a function returning T', unit=u, where=w)
                for u, w in [('a:1:1', 'b:2:1'), ('a:1:1', 'b:2:1'),
                             ('a:1:1', 'b:3:1'), ('c:1:1', 'b:2:1')]]
        result = summarize(rows, MUTANT)[0]
        self.assertEqual(result['roots'], 2)
        self.assertEqual(result['sites'], 2)
        self.assertEqual(three_witnesses(['a:1:2', 'a:1:9', 'a:2:1', 'b:3:1']), ['a:1', 'a:2', 'b:3'])
    def test_collision_scope(self):
        scopes = {'a.ts': [dict(start=[2, 1], end=[5, 9], names=['SourceFile'])]}
        self.assertTrue(bound_collision('a value of type SourceFile', 'a.ts:3:1', scopes))
        self.assertFalse(bound_collision('a value of type SourceFile', 'a.ts:6:1', scopes))
        self.assertFalse(bound_collision('a value of type SourceFile', 'b.ts:3:1', scopes))
        self.assertFalse(bound_collision('a value of type SourceFile', 'none (one file:line site):1', scopes))
    def test_selection(self):
        names = ['T', 'TData', 'K', 'SourceFile']
        for reason in ['a function returning T | undefined', 'a value of type Box<TData>',
                       'a generic function as a value', 'a value of type T["kind"]']:
            self.assertTrue(scoped(reason, names))
        for reason in ['a function returning Type', 'reading T', 'a value of type SourceFile']:
            self.assertFalse(scoped(reason, names))

MUTANT = False
MUTANT_SELECTION = False
MUTANT_MARKDOWN = False

def build(census, parameters):
    parameter_data = json.loads(Path(parameters).read_text())
    names = parameter_data['names']
    scopes = parameter_data['scopes']
    def generic_hatch(reason, where):
        if 'cast' not in reason and 'predicate' not in reason:
            return False
        filename, line, column = where.rsplit(':', 2)
        filename = filename.split('/src/compiler/', 1)[-1]
        position = (int(line), int(column))
        return any(tuple(r['start']) <= position <= tuple(r['end']) for r in parameter_data['genericUses'].get(filename, []))
    opener = gzip.open if str(census).endswith('.gz') else open
    with opener(census, 'rt') as stream:
        records = [json.loads(line) for line in stream]
    def local(where):
        return where.split('/src/compiler/', 1)[-1]
    findings = []
    collisions = []
    for record in records[1:]:
        for row in record.get('findings', []):
            if row['kind'] not in ('NotYet', 'Refused'):
                continue
            selected = scoped(row['reason'], names) or bound_collision(row['reason'], row['where'], scopes) or generic_hatch(row['reason'], row['where'])
            collided = not selected and not row['reason'].startswith('reading ') and any(
                re.search(r'\b' + re.escape(name) + r'\b', row['reason']) for name in COLLISIONS)
            if selected or collided:
                kept = {k: row[k] for k in ['kind', 'reason', 'unit', 'where', 'phase']}
                kept['unit'], kept['where'] = local(kept['unit']), local(kept['where'])
                (findings if selected else collisions).append(kept)
    frozen_bytes = subprocess.check_output(['git', 'show', '6c4fc1af:stage3/census/hidden-ranking/RESULT.json'])
    frozen = json.loads(frozen_bytes)
    ranked = {(r['kind'], r['reason']): r for r in frozen['ranked_reasons']}
    refusal_text = subprocess.check_output(['git', 'show', '6c4fc1af:stage3/census/hidden-ranking/evidence/refusal-table.md']).decode()
    refusal_rows = {}
    for line in refusal_text.splitlines():
        fields = table_fields(line)
        if len(fields) != 10 or not fields[2].isdigit():
            continue
        reason = fields[0]
        if scoped(reason, names) or any(bound_collision(reason, w + ':1', scopes) for w in fields[6:8] if re.search(r':\d+$', w)):
            refusal_rows[('Refused', reason)] = dict(sites=int(fields[2]), witnesses=[w for w in fields[6:8] if re.search(r':\d+$', w)])
    # Keep historic-only exact reasons too, including zero-credit reasons.
    combined = {(r['kind'], r['reason']): r for r in summarize(findings)}
    for key, row in ranked.items():
        if scoped(row['reason'], names) or any(bound_collision(row['reason'], b['where'], scopes) for b in frozen['boundaries'] if (b['kind'], b['reason']) == key):
            combined.setdefault(key, dict(kind=key[0], reason=key[1], roots=0, sites=0, witnesses=[]))
    for key in refusal_rows:
        combined.setdefault(key, dict(kind=key[0], reason=key[1], roots=0, sites=0, witnesses=[]))
    for key, row in combined.items():
        table = refusal_rows.get(key)
        row['historical_refusal_table_sites'] = table['sites'] if table else None
        if table:
            row['witnesses'] = list(dict.fromkeys(row['witnesses'] + table['witnesses']))[:3]
        old = ranked.get(key)
        row['historical_hidden_bytes'] = old['bytes_revealed_if_fixed_alone'] if old else None
        row['historical_boundaries'] = old['boundary_count'] if old else None
        witnesses = list(row['witnesses'])
        for boundary in frozen['boundaries']:
            if (boundary['kind'], boundary['reason']) == key:
                where = boundary['where']
                if where not in witnesses:
                    witnesses.append(where)
        row['witnesses'] = three_witnesses(witnesses)
    root_positions = {f['unit'] for f in findings}
    root_units = [{**u, 'where': local(u['where'])} for record in records[1:] for u in record.get('units', []) if local(u['where']) in root_positions]
    result = dict(root_units=root_units, base='8cb5e7c189fc431ab3cab7e0be2c0530e722a598',
                  census_sha256=hashlib.sha256(Path(census).read_bytes()).hexdigest(),
                  ranking_commit='6c4fc1afb019d0a16fea97082e45fef8fcbe1746',
                  ranking_sha256=hashlib.sha256(frozen_bytes).hexdigest(),
                  refusal_table_sha256=hashlib.sha256(refusal_text.encode()).hexdigest(),
                  ranking_compiler=frozen['provenance']['compiler_commit'],
                  checker_diagnostics=len(records[0].get('diagnostics', [])),
                  parameters=names, source_sha256=parameter_data['hashes'], collisions=sorted(COLLISIONS),
                  rows=sorted(combined.values(), key=lambda r: (-(r['historical_hidden_bytes'] or 0), r['kind'], r['reason'])),
                  findings=findings, collision_rows=summarize(collisions))
    dest = Path(__file__).parent
    with (dest / 'baseline.json.gz').open('wb') as raw:
        with gzip.GzipFile(filename='', fileobj=raw, mode='wb', mtime=0) as stream:
            stream.write(json.dumps(result, indent=2).encode())
    lines = ['# Step 16: generics at runtime and inside generics', '',
             '## Evidence and counting', '',
             'The compiler base is `8cb5e7c1`. The guarded full census was rerun on this base over the adapted pinned TypeScript compiler directory. It is measurement on a checker-rejected program, with ' + str(result['checker_diagnostics']) + ' diagnostics, and cannot emit IR or invoke either backend.', '',
             'Roots count distinct attempted `unit` positions for each exact reason. Sites count distinct diagnostic positions. Repeated imported-helper observations and refusal-scan/lowering repetitions do not add roots. Root counts across reasons overlap and must not be summed as distinct programs.', '',
             'Hidden bytes are the frozen outermost-cause attribution from `codex/stage3-hidden-ranking` at `6c4fc1af`, measured on `ed6e2975`, not this base. They are exact-reason-wide totals, so short cast/predicate reasons can include historical nongeneric sites. They estimate exposure after removing one outer reason; they do not claim successful lowering or measured retirement. `unmeasured` means this exact reason has no row in that historical ledger. Zero means a measured row with no credited bytes.', '',
             'The AST type-parameter names select exact type-bearing diagnostics, plus all explicit generic/type-parameter/monomorphization diagnostics. Lexical reading-X failures are excluded. Concrete-name collisions enter the scoped table only where TypeScript AST ranges prove a binder with that spelling at the witness. The remaining collisions are retained separately below. Cast and predicate findings with an AST target containing an enclosing binder are included even when the short reason omits the target type. Generic refusals remain soundness obligations, not automatic compiler lessons.', '',
             'The frozen giant-body refusal table was also reconciled by exact reason. Its site counts are retained in the machine ledger, not substituted for attempted roots. It measures compiler 0d3f2715 over eight selected giant bodies. Table-only reasons stay visible with zero current roots and unmeasured hidden bytes. Witnesses use diagnostic positions on this base where available, then historical table/boundary positions. Fewer than three distinct witnesses exist for several reasons; those rows retain every available witness rather than inventing locations. Historical-only rows have zero current roots. Full provenance and selected observations are in [baseline.json.gz](step-16-generics/baseline.json.gz).', '',
             '| Kind and exact reason | Base roots | Base sites | Historical hidden bytes | Historical boundaries | Up to three witnesses |',
             '| --- | ---: | ---: | ---: | ---: | --- |']
    for r in result['rows']:
        reason = r['reason'].replace('|', chr(92) + '|')
        witnesses = '<br>'.join(r['witnesses']) or 'unavailable'
        lines.append(f"| {r['kind']}: {reason} | {r['roots']} | {r['sites']} | {r['historical_hidden_bytes'] if r['historical_hidden_bytes'] is not None else 'unmeasured'} | {r['historical_boundaries'] if r['historical_boundaries'] is not None else 'unmeasured'} | {witnesses} |")
    lines += ['', '## Concrete-name collision audit', '',
              'These exact reasons contain a spelling also used as a type parameter. Their spelling alone does not establish a generic root. They are excluded from the step totals because the witness has no enclosing type-parameter binder with that spelling. This is a lexical classification, not checker-identity proof.', '',
              '| Exact reason | Roots | Witnesses |', '| --- | ---: | --- |']
    for r in result['collision_rows']:
        lines.append('| ' + r['kind'] + ': ' + r['reason'].replace('|', chr(92) + '|') + ' | ' + str(r['roots']) + ' | ' + '<br>'.join(r['witnesses']) + ' |')
    lines += ['', '## Reproduction', '', 'Run the setup, adaptation, overlay build and census commands in a checkout of compiler base `8cb5e7c1`. Run the checked-in report scripts from this delivery branch, where the historical ranking object must also be available. Rebuilding the census on a newer compiler is a different measurement.', '', '```sh',
              "export GOPROXY='https://proxy.golang.org|direct'", 'bash cloud/setup.sh > /tmp/scout-setup.log 2>&1',
              'source /workspace/adamic-tools/env.sh', 'bash stage3/apply.sh /tmp/scout-adapted > /tmp/scout-adapt.log 2>&1',
              'python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/scout-overlay > /tmp/scout-overlay.log 2>&1',
              'go build -buildvcs=false -overlay=/tmp/scout-overlay/overlay.json -o /tmp/scout-census ./stage3/census/latent/tool > /tmp/scout-census-build.log 2>&1',
              'LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/scout-census /tmp/scout-adapted/src/compiler /tmp/scout-base.jsonl > /tmp/scout-base-census.log 2>&1',
              'node docs/step-16-generics/parameters.cjs "$HOME/.cache/adamic-stage3/api/node_modules/typescript" /tmp/scout-adapted/src/compiler > /tmp/scout-parameters.json',
              'python3 docs/step-16-generics/audit.py /tmp/scout-base.jsonl /tmp/scout-parameters.json',
              'python3 docs/step-16-generics/audit.py --test > /tmp/scout-audit.log 2>&1',
              'python3 docs/step-16-generics/parameters_test.py "$HOME/.cache/adamic-stage3/api/node_modules/typescript" > /tmp/scout-parameters-test.log 2>&1',
              'python3 docs/step-16-generics/parameters_test.py "$HOME/.cache/adamic-stage3/api/node_modules/typescript" --mutant > /tmp/scout-parameters-mutant.log 2>&1',
              'python3 docs/step-16-generics/audit.py --mutant > /tmp/scout-audit-mutant.log 2>&1',
              'python3 docs/step-16-generics/audit.py --mutant-selection > /tmp/scout-selection-mutant.log 2>&1', '```', '',
              'The parameters file is the sorted unique names of every TypeParameterDeclaration in the pinned compiler AST, parsed with TypeScript 6.0.3. The complete name list is retained in baseline.json.gz. The census SHA-256 pins the complete scratch measurement; selected findings remain reviewable without that scratch file.', '']
    document = dest.parent / 'step-16-generics.md'
    suffix = ''
    if document.exists():
        previous = document.read_text()
        marker = '## Design and disposition'
        if marker in previous:
            suffix = '\n' + previous[previous.index(marker):]
    document.write_text('\n'.join(lines) + suffix)
    print(f"{len(result['rows'])} exact reasons; {len(findings)} selected observations; {len(result['collision_rows'])} collision reasons")

if __name__ == '__main__':
    import sys
    if '--test' in sys.argv or '--mutant' in sys.argv or '--mutant-selection' in sys.argv or '--mutant-markdown' in sys.argv:
        MUTANT = '--mutant' in sys.argv
        MUTANT_SELECTION = '--mutant-selection' in sys.argv
        MUTANT_MARKDOWN = '--mutant-markdown' in sys.argv
        unittest.main(argv=[sys.argv[0]])
    else:
        parser = argparse.ArgumentParser()
        parser.add_argument('census')
        parser.add_argument('parameters')
        args = parser.parse_args()
        build(args.census, args.parameters)
