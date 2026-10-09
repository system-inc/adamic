"""Render reviewable tables from raw API evidence; never invent a proof for an unresolved row."""
import json
from pathlib import Path
from collections import Counter
import subprocess
base = Path(__file__).resolve().parent
data = base / 'data'
def read(name): return json.loads((data / (name + '.json')).read_text())
def loc(value): return f"{value['file']}:{value['line']}:{value['column']}"
rows, counts = read('factories'), read('counts')
builders = [r for r in rows if r['nodeResult'] and r['factoryName'] and not r['parser']]
deferred = read('deferred_fields')
names = {p['field'] for p in deferred}
late = [w for w in read('phase_writes') if w['field'] in names and w['plainAssignment'] and not w['undefined']]
(data / 'late_completion_candidates.json').write_text(json.dumps(late, indent=2) + '\n')
fixtures = read('fixtures')
# Keep command logs as text evidence rather than altering the repository's ignored-log policy.
for label in ['setup', 'build', 'inventory', 'fixtures', 'audit', 'api', 'surface', 'proofs-cli']:
    source = Path('/tmp/step24-factories-' + label + '.log')
    if source.exists(): (data / (label + '-output.txt')).write_bytes(source.read_bytes())
provenance = {
    'adamic_base_commit': '45487a809f89885a3fc651cd590e7dabf31362dc',
    'typescript_tag': 'v6.0.3', 'typescript_commit': '050880ce59e30b356b686bd3144efe24f875ebc8',
    'typescript_api': counts['apiVersion'],
    'cohere_commit': subprocess.check_output(['git', '-C', 'cohere', 'rev-parse', 'HEAD'], text=True).strip(),
    'source_root': '/workspace/cache/tsc-census/typescript',
    'prepared_root': '/workspace/cache/tsc-census/prepared',
    'generated_input': 'src/compiler/diagnosticInformationMap.generated.ts, official processDiagnosticMessages.mjs',
    'source_modifications': 'None in factory/, parser.ts, types.ts or utilities.ts; byte pins in source_manifest.json',
    'node_version': fixtures['nodeVersion'], 'nproc': 5, 'cpu_quota': 4,
}
(data / 'provenance.json').write_text(json.dumps(provenance, indent=2) + '\n')
lines = [
    'Built: every factory-directory callable indexed, per-field boundary evidence, parser paths, five .a fixtures and mutants.',
    'Commits: based on 45487a809f89885a3fc651cd590e7dabf31362dc; delivery SHA is recorded by the final branch commit.',
    'Commands/results: TypeScript API inventory and audit pass; five source Node goldens pass; two native/JS/sanitizer/count cases pass.',
    'Mutants: five fixture changes, three scanner changes and three artifact changes caught; exact catches are in data/.',
    'Not covered: an exhaustive interprocedural initialization proof or native acceptance of the three staged casts; unresolved rows stay explicit.',
    '', '# Step 24: factories completed before escape', '',
    'This scout is evidence for #ktz9fek and the October 8 ruling, not its compiler implementation. TypeScript is pinned to v6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`, as [stage3/source.json](../../../source.json) records. The source is outside this repository. The upstream compiler project is loaded with stock TypeScript 6.0.3 and its actual tsconfig; it has **zero diagnostics** after official diagnostic-map generation and the pinned Node declarations already used by the census.', '',
    'The important observed distinction: concrete factories finish syntax fields, while `parent`, `symbol`, and several SourceFile fields deliberately remain `undefined`. Assigning `undefined!` is a physical write, not proof of the declared non-undefined type. Under the ruling those fields require loud reads until valid completion is proven. A blanket claim that every required type is true at a factory return would be false.', '',
    '## Counts and their populations', '',
    '| Measurement | Count |', '| --- | ---: |',
    f"| Files in factory/ plus parser.ts | {counts['files']} |",
    f"| All callable bodies, including arrows/getters/methods | {counts['allCallableBodies']} |",
    f"| Detailed rows (builders, node results, node-parameter helpers) | {len(rows)} |",
    f"| Public NodeFactory members, aliases and overloads included | {len(read('node_factory_surface'))} |",
    f"| Node-returning create/update bodies in factory/ | {counts['factoryNodeBuilders']} |",
    '| Of those: create / update | 316 / 176 |',
    f"| All required runtime slots definitely written before this boundary, allowing explicit undefined placeholders | {counts['factoryAllRequiredSlotsWrittenBeforeEscape']} |",
    f"| Certified completion before escape without an invalid undefined placeholder | {counts['factoryCompleteBeforeEscape']} |",
    f"| Certified completion apart from parent, without another invalid undefined placeholder | {counts['factoryCompleteApartFromParent']} |",
    f"| Factory / required-field pairs explicitly undefined at escape | {counts['deferredRequiredFactoryFieldPairs']} |",
    f"| Distinct names in those deferred pairs | {counts['distinctDeferredRequiredFieldNames']} |",
    f"| Observed post-escape writes, including rewrites, in detailed bodies | {counts['afterEscapeWriteSites']} |",
    f"| Those post-escape writes targeting required fields | {counts['requiredAfterEscapeWriteSites']} |",
    f"| Later plain non-undefined assignment sites targeting deferred names, across compiler/ | {len(late)} |",
    f"| Distinct deferred names at those later assignment sites | {len({w['field'] for w in late})} |",
    f"| Parser Node-returning bodies | {counts['parserNodeProducingBodies']} |",
    '| Parser createBase* implementations (two five-member allocators) | 10 |',
    f"| Parser finishNode call sites | {counts['parserFinishCalls']} |", '',
    '**These are different counts and must not be added.** The zero is the number certified by this profile, not a claim that no caller-provided node can already be complete. Required `parent` is undefined at default fresh factory returns, and there are additional deferred semantic fields. The 254 write-complete bodies are a useful completion-work population; they are not 254 wholly trusted objects. The 129 count excludes only parent and does not erase symbol or SourceFile obligations. No successful native lowering is inferred from any inventory count.', '',
    'The 82 observed post-escape writes include range and flag rewrites; those are not all first initialization. The 30 later completion candidates are exact static sites, not a claim that every execution reaches a first write there. They include binder/checker parent/symbol writes. `lineMap` is lazy and emit metadata has additional helpers. Unknown effects, aliasing through members, allocator patching and phase-specific preconditions prevent an exhaustive exact global first-completion count with this scout. The lower bounds and unresolved evidence are kept, rather than treating missing proof as completion.', '',
    '### Factory body populations', '', '| File | Node-returning create/update bodies |', '| --- | ---: |',
]
for file, number in sorted(Counter(r['location']['file'] for r in builders).items()): lines.append(f'| {file} | {number} |')
lines += ['', '### Deferred required fields', '', '| Field | Factory/field pairs with explicit undefined at escape | Later candidate assignment sites |', '| --- | ---: | ---: |']
late_counts = Counter(w['field'] for w in late)
for field, number in Counter(p['field'] for p in deferred).most_common(): lines.append(f'| {field} | {number} | {late_counts[field]} |')
lines += ['', '## How to read the map', '',
    '[FACTORIES.md](FACTORIES.md) lists every detailed row and [FIELDS.tsv](FIELDS.tsv) gives one required field per row with status and evidence locations (`-` means no evidence in that column). [data/functions.json](data/functions.json) indexes every callable body, including the bodies outside the node-builder population. [data/node_factory_surface.json](data/node_factory_surface.json) independently lists all public factory members and their overload declarations, so aliases and convenience getters are not silently omitted.', '',
    'The full [data/factories.json](data/factories.json) contains each made/returned type, inherited required fields, union-member applicability, declaration file:line, direct writes, allocator write lines, summarized callee locations, first escapes, returns and unresolved control forms. Type-only `_...Brand` members are listed but excluded from runtime-slot counts: the stock allocator does not materialize them. Optional properties are excluded from required fields, even when a factory explicitly writes them. A required property whose type includes undefined may legitimately hold it.', '',
    'Statuses are `definitely-written-before-escape`, `written-undefined-before-escape`, `written-by-return-after-earlier-escape`, `conditional-or-path-incomplete-write`, `no-proven-write`, `unresolved-control`, and `phantom-type-brand`. `no-proven-write` means the scout lacks proof; it does not mean the source never writes the field. Helper/forwarding rows do not assume caller parameters are fully initialized merely because their declared types say so.', '',
    'Analysis intersects branch writes, follows direct local aliases and records return, store, pass and closure-capture boundaries. A compound assignment does not initialize an absent field. Loops, abrupt control and try paths disable completion claims and are explicitly marked unresolved. Calls with node arguments count as escapes. Field reads are not whole-object escapes. Returning a partial base node is an escape of that helper; the enclosing concrete factory starts another boundary with the returned write summary. Thus completion is reported per factory boundary, not by pretending the earlier helper return never occurred. A transferred summary is physical-write evidence only.', '',
    'Default allocator assumptions are explicit in each object origin. Custom BaseNodeFactory implementations and ObjectAllocator patchers are not proven by the default profile. The constructor write map is taken from utilities.ts:8505-8543, including `parent = undefined!`; createBaseDeclaration at nodeFactory.ts:1213-1217 explicitly puts `symbol` in the same state. Direct source factories are checked against their own subtype requirements, not just Node.', '',
    '## Parser base and finish paths', '',
    'The parser implements two BaseNodeFactory tables at parser.ts:433-437 and 1463-1467. Each has all five createBase* arrows; the second wraps allocation in countNode. There are no parser createBase* call expressions: those implementations are invoked through the node factory. Every one of the 214 `finishNode` calls, its concrete/generic Node type, required fields and exact expression is in [data/parser_paths.json](data/parser_paths.json).', '',
    '`finishNode` at parser.ts:2600 sets pos/end through setTextRangePosEnd and conditionally ORs flags. It never fills every field of generic T and never sets parent or symbol. The range-setter effects are explicitly audited from utilities.ts:10644, 10654, 10664 and 10674. Their write evidence remains a passed-node boundary; the scout does not infer that every helper is non-escaping.', '',
    'SourceFile illustrates real later completion: factory/nodeFactory.ts:6076-6098 initializes required arrays/maps/diagnostics to undefined. Parser setFields writes bindDiagnostics at parser.ts:2005. parseSourceFileWorker writes counts, identifiers and parseDiagnostics at parser.ts:1826-1829 after the creation path has returned. Parent installation is a separate fixup/binder path, described by parser.ts:1971-1975. The whole-compiler required-write ledger is [data/phase_writes.json](data/phase_writes.json); the deferred-name subset is [data/late_completion_candidates.json](data/late_completion_candidates.json).', '',
    '## Five real-source reductions and their mutants', '',
    '| Fixture | Upstream origin | Boundary | Mutant and catcher |', '| --- | --- | --- | --- |',
    '| 01-binary-complete.a | nodeFactory.ts:3419 | left/operatorToken/right completed before concrete return | omit right; source Node golden changes 4 to undefined |',
    '| 02-variable-complete.a | nodeFactory.ts:4251 | name and initializer completed, including permitted undefined initializer | omit name; source Node golden changes answer to undefined |',
    '| 03-numeric-conditional.a | nodeFactory.ts:1233, utilities.ts:8512 | base flags initialized on both paths; branch only augments | omit default flags; false branch prints undefined; source Node golden catches it |',
    '| 04-sourcefile-after-escape.a | nodeFactory.ts:6041, parser.ts:2003 | diagnostics completed after return; explicit loud read | omit late diagnostics write; source Node and native panic with exit 70 and the same reason |',
    '| 05-parser-finish.a | parser.ts:2600 | pos/end completion before finishNode returns, flag branches retained | omit end; source Node and native golden changes 7 to -1 |', '',
    'Children/enums are reduced to numbers/strings and only relevant fields are retained. The first three preserve the staged cast, rather than replacing their completion behavior with a finished literal. Main currently refuses that cast; each first-line a-check refusal header is measured against the real CLI. This is direct evidence that #ktz9fek still has native acceptance work on this base. The fourth makes the latent field an explicit union and uses `?? panic` to express the loud check; this is a safety translation and does not claim compiler-inserted checks exist. The fifth has an already allocated fixed shape. Clean cases have `a-check: clean` comments; the harness requires real clean compilation. All mutants are .a files with measured headers too.', '',
    'The five goldens run the source through unchanged oracle/node.mjs on Node, not through Adamic output. The two compilable fixtures also match generated JavaScript and native under ASan/UBSan with LeakSanitizer enabled. Counted outputs and the pending refused rows are in [counts.md](counts.md). No internal/oracle fixture was added and no file outside this territory was edited.', '',
    '[data/fixtures.json](data/fixtures.json) preserves every command, stdout, stderr, exit, mutant, native and JS observation. Five fixture mutants are independently caught; the two compilable mutants additionally match their source-Node observations. The late-diagnostics mutant is a panic check, not an unrelated compilation failure.', '',
    'Stock TypeScript itself is also sampled independently: [api_probe.cjs](api_probe.cjs) asserts real factory children, conditional flags, undefined deferred fields, parser diagnostics completion and child parent installation. [data/api_observations.json](data/api_observations.json) records required-field ownness and undefined states. This sample corroborates the boundary distinction; it does not prove every factory.', '',
    '## Inventory checks and mutants', '',
    '[proofs.a](proofs.a) supplies eight independent scanner witnesses: complete writes, one-sided conditional writes, both-sided writes, pass before completion, alias writes, a zero-iteration loop, compound-only writes and explicit undefined. These are scanner evidence, not eight new runtime goldens. The file itself has the measured cast-refusal header.', '',
    'Three scanner mutants run on scratch copies and are caught by the expected boundary-status assertions: union rather than intersection at a branch join falsely completes createConditional; ignoring passed objects falsely completes createPassed; ignoring undefined literal values falsely completes createUndefined. Three artifact mutants are also caught: inflated full-completion count, dropped right-field escape evidence, and changed source hash. [data/audit.json](data/audit.json) and [data/audit-output.txt](data/audit-output.txt) contain all exact catches. These are evidence-tool mutants, not mutants of the pending compiler pass or its memory-management code.', '',
    '## Commands and setup', '', 'All test output was redirected to log files, with text snapshots committed under data/. No whole-package or full-gate confirmation was run.', '', '```sh',
    "export GOPROXY='https://proxy.golang.org|direct'",
    'bash cloud/setup.sh > /tmp/step24-factories-setup.log 2>&1',
    'source /workspace/adamic-tools/env.sh',
    'go build -o /workspace/cache/step24-adamic ./cmd/adamic > /tmp/step24-factories-build.log 2>&1',
    'node stage3/scouts/step24/factories/inventory.cjs /workspace/cache/tsc-census/prepared stage3/scouts/step24/factories/data > /tmp/step24-factories-inventory.log 2>&1',
    'node stage3/scouts/step24/factories/surface.cjs /workspace/cache/tsc-census/prepared > /tmp/step24-factories-surface.log 2>&1',
    'node stage3/scouts/step24/factories/api_probe.cjs > /tmp/step24-factories-api.log 2>&1',
    'node stage3/scouts/step24/factories/check.cjs /workspace/cache/step24-adamic > /tmp/step24-factories-fixtures.log 2>&1',
    'node stage3/scouts/step24/factories/audit.cjs /workspace/cache/tsc-census/prepared > /tmp/step24-factories-audit.log 2>&1',
    'python3 stage3/scouts/step24/factories/render.py > /tmp/step24-factories-render.log 2>&1',
    '```', '',
    'The source cache from the census was reused and its exact tag SHA verified. The prepared tree copies src/scripts/package files, generates diagnosticInformationMap.generated.ts with the upstream processDiagnosticMessages.mjs, and resolves node_modules to the census npm cache containing TypeScript 6.0.3 and @types/node 25.3.3. TYPESCRIPT_API may select an equivalent pinned module path. Every scoped source file is byte-pinned in [data/source_manifest.json](data/source_manifest.json).', '',
    'Setup passed: Go 1.27.1, Node 24.19.0, clang 20.1.8. Timing lines: Go 0.138s; Node 0.203s; clang 1.451s; markdown install step 2.372s, ready 2.810s; submodules 6.421s; Go build 650.799s; tests deferred 651.228s; cache warm 651.231s; done 651.302s. nproc 5, cgroup quota four. The initial git fetch refreshed main promptly but then spent over fifteen minutes recursively fetching unrelated historical submodule commits; that recursive fetch was stopped after setup had successfully checked out the pinned submodules. The branch was created from refreshed origin/main 45487a80, not from the earlier census branch. See [data/setup-output.txt](data/setup-output.txt).', '',
    '## Limits and implementation handoff', '',
    'Exact observations: source populations, declarations/locations, the listed writes/escapes, local branch intersections, the field-status witnesses, the runtime samples and fixture outputs. Inference: applicability to #ktz9fek and which deferred reads it will need to check. The compiler worker must make the whole-program, ownership-aware and interprocedural proof; it must handle branches, aliases, stores, callbacks, allocator overrides and partial-helper returns, not whitelist factory names.', '',
    'The map does not decide fields behind computed keys, prototype-backed redirected SourceFiles, general helper effects, every global/container alias, recursive parser effects or phase-specific caller preconditions. There are 49 builder rows with recorded gaps, eight with unsupported control. Union requirements are listed per constituent, conservatively combined for completion. Therefore this unit cannot honestly label every unresolved field as definitely-before or definitely-after, nor give an exhaustive exact global count of first valid writes after escape. Every unresolved row is retained for follow-up. Three staged-cast fixtures remain current refusals; full native tsc and the pending compiler pass were not tested.', '',
    'No blanket trust should be granted from the declared Node subtype or from a physical undefined write. Once the compiler proves syntax completion, the semantic fields still need their loud checks until their own completion is proven.',
]
(base / 'REPORT.md').write_text('\n'.join(lines) + '\n')
# The compact index includes helpers and non-node creators, avoiding an undocumented selection gap.
index = ['# Factory and parser body index', '', 'All detailed rows from data/factories.json. Field statuses and evidence are in FIELDS.tsv; every remaining callable body is in data/functions.json. Completeness is for this function boundary under the stock allocator profile.', '', '| Body and source | Made/returned type | Role | Write-complete | Non-undefined complete |', '| --- | --- | --- | --- | --- |']
for row in rows:
    complete = row['nodeResult'] and bool(row['requiredFields']) and all(p['phantom'] or p['status'] in ['definitely-written-before-escape', 'written-undefined-before-escape'] for p in row['requiredFields'])
    index.append(f"| {row['name']} at {loc(row['location'])} | {row['nodeType'].replace('|', '&#124;')} | {row['classification']} | {'yes' if complete else 'unproven'} | {'yes' if row['completeBeforeEscape'] else 'unproven'} |")
(base / 'FACTORIES.md').write_text('\n'.join(index) + '\n')
fields = ['body\tnode_type\tfield\tfield_type\tstatus\tdeclaration\tdirect_writes\tallocator_writes\tcallee_evidence\treturn_evidence']
for row in rows:
    for p in row['requiredFields']:
        values = [row['id'], row['nodeType'], p['name'], p['type'], p['status'], loc(p['declaration']), ';'.join(loc(w['location']) + (' after-escape' if w['afterEscape'] else '') for w in p['writes']), ';'.join(loc(w['location']) for w in p['allocatorEvidence']), ';'.join(w['callee'] + ' at ' + loc(w['location']) for w in p['calleeEvidence']), ';'.join(loc(w) for w in p['returnEvidence'])]
        fields.append('\t'.join(v.replace('\t', ' ').replace('\n', ' ') or '-' for v in values))
(base / 'FIELDS.tsv').write_text('\n'.join(fields) + '\n')
print(f'PASS report, {len(rows)} indexed bodies, {len(fields)-1} required-field rows, {len(late)} later candidate sites')
