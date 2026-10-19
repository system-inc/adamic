'use strict';
const fs=require('node:fs'),path=require('node:path');
const directory=__dirname,data=JSON.parse(fs.readFileSync(path.join(directory,'void-reads-2026-10-19.json'),'utf8'));
const coverage=data.runtimeCoverage,oracle=coverage.oracle;
const rows=data.sites.filter(s=>s.judgment.kind==='unclassified');
const unhit=data.sites.filter(s=>s.runtime.instrumented&&s.runtime.hits===0);
const escape=text=>String(text).replace(/\r?\n/g,' ').replace(/\|/g,'\\|');
let text=`Measured 79 adapted compiler roots from main ${data.mainSha}.<br>
Found 491 void-typed reads: 383 returns, 98 concise arrow bodies, 7 conditional branches, 3 truthiness operands.<br>
Proved ${data.countsByJudgment.adaptation} annotation-only adaptations with byte-identical JavaScript; ${rows.length} sites remain unclassified.<br>
Reserved all 98 concise arrow bodies and all 16 union-with-void sites for the language lead.<br>
Node corpus: 106366 passed, one independently reproduced API baseline failure; ${coverage.hitSites}/393 sites observed; mutant adds exactly one site.

# Void reads, October 19, 2026

This is a measurement, with a partial classification. No compiler source repair is committed. The table is [void-reads-2026-10-19.json](void-reads-2026-10-19.json). Observations (locations, types, forms, resolved declarations, runtime counters) and judgments are separate fields. Lines and UTF-16 columns are one-based. Root hashes identify the exact adapted inputs.

| Form | Count |
|---|---:|
${Object.entries(data.countsByForm).map(([form,count])=>`| ${form} | ${count} |`).join('\n')}
| All other forms in the brief | 0 |
| **Total** | **491** |

| Judgment | Count |
|---|---:|
| adaptation | ${data.countsByJudgment.adaptation} |
| real read, with impossibility proved | 0 |
| concise arrow body, unjudged | 98 |
| unclassified | ${rows.length} |

## Construction and boundaries

Stock TypeScript **6.0.3**, installed from stage3/api's lockfile, runs the checker. Program construction is reused from checker-census.cjs at **21f6268d6dad5d39c985558b8469473d4f68a303**, using its loader library profile and compiler options. This includes its loader-compatible RegExp/string declaration edits. The prelude is loaded with exactly the MapIterator and SetIterator next declarations removed, preserving line positions. Its original and filtered hashes are in the table. These are measurement-only overlays outside the checkout, not edits under internal/.

The census visits ${data.expressionCount} runtime expression nodes in the 79 roots. It uses TypeScript's context-aware expression predicate, skips erased type syntax, and does not mistake declaration names for reads. It discards expression-statement results, void operands, comma left operands, for initializer/incrementor expressions, and comma results or parentheses carrying a discarded value. Nested expressions remain separate sites. There is no deduplication by source line or statement. Unions containing void are excluded from the 491.

The measurement program has **${data.adaptationProof.baselineDiagnostics} pre-existing checker diagnostics** under this construction. This is not a claim that the adapted tree is checker-clean. The annotation overlay also has ${data.adaptationProof.rewrittenDiagnostics} diagnostics and introduces zero diagnostics, comparing file, code, message and multiplicity before exploratory type queries.

## Adaptations

For 350 of the proved sites, the resolved implementation has no value-returning statement, excluding nested functions. Replace its explicit void return annotation with undefined, or add an undefined return annotation when void was inferred. For example, binder.ts's bind changes from:

\`\`\`typescript
function bind(node: Node | undefined): void {
\`\`\`

to:

\`\`\`typescript
function bind(node: Node | undefined): undefined {
\`\`\`

Its body is untouched. Every site records the exact signature replacement, offsets in the measured file, declaration line, and the checker result **undefined** (or **boolean** for the marker handler below). The combined overlay changes ${Object.values(data.adaptationProof.files).reduce((sum,p)=>sum+p.edits.length,0)} annotations across eight files; edits are listed once under adaptationProof.files. This is a batch proof: apply the complete listed overlay, rather than assuming that any arbitrary subset was separately checked.

Full-file ts.transpileModule output, using the measurement compiler options, is compared byte for byte before and after. Each pair of SHA-256 hashes is recorded. verify.cjs independently reconstructs all eight outputs and rejects eight mutants, each appending a runtime 0; statement to one rewritten file. No adaptation is credited solely because it looks erasable. The overlay is never written to tsc's source.

One additional adaptation refines outofbandVarianceMarkerHandler from a void-returning callback to a boolean-returning callback. Its assignments either return true from assignments to unreliable/unmeasurable or forward to the saved handler. Runtime coverage observes true at checker.ts:23401, so treating void as necessarily undefined would be wrong here. The table records the implementations and the checked annotation rewrite. This refinement is included in the same byte-identical overlay proof.

## Runtime observations and limits

The original measured source, without annotation adaptations, was copied to a separate scratch tree. Every non-arrow site was wrapped in a value-preserving observer. The standard stage3 oracle built the instrumented compiler and ran its default upstream test corpus on Node with four workers. The observer records completed evaluations: an expression that throws or exits before returning produces no value hit. Thus these are completed-read counts, not attempted-call counts.

Oracle status: **${oracle.status}**. Reported test totals: **${JSON.stringify(oracle.counts||{})}**. Baseline differences: **${(oracle.baseline_diffs||[]).length}**. Across ${coverage.processFiles} process snapshots, **${coverage.totalHits} completed reads** cover **${coverage.hitSites} sites**; **${unhit.length} instrumented sites have zero hits**. Each site's runtime field contains its count and every distinct observed value. Undefined is represented explicitly. Object, function and symbol identities, if present, are process-local; numbers retain NaN, infinities and negative zero.

Numeric values 0, 1, 2 and 4 were also observed at sys.exit call sites. Test host overrides use an expression-bodied assignment that returns the assigned exit code; production and throwing host implementations are recorded separately as implementation evidence. The observed numeric results must not be replaced by undefined.

A finite corpus cannot prove that an arbitrary host callback always returns undefined. Zero hits do not prove unreachability. This unit does not claim that the remaining ${rows.length} reads have no JavaScript-identical adaptation: the conservative no-value-return analysis did not prove one, and no impossibility proof was established. They include callbacks and host hooks, interface/overload signatures without an implementation body, and forwarding functions with return expressions. The table includes resolved declarations, including implementation bodies where the checker resolves directly to one. An exhaustive implementation inventory for external callbacks was not established. No site is labeled real read by default.

The sole corpus failure is the public API declaration baseline. An uninstrumented adapted-tree control ran the Public APIs tests: 30 passed and one failed. Its baseline diff is byte-identical to the instrumented run (SHA-256 3da14ced9f2c7eea0ae8b4dabbe6067a91510b0ab0d950ddb83903a9eb50fa28). This establishes the failure independently of the counter instrumentation. No baseline was accepted or changed.

The counter control passes eight values with six distinct value kinds, preserves return identity, and catches removal of the hit-count increment. This is recorded in counter-proof.json. Raw process snapshots and oracle logs are retained in evidence/.

## Unclassified sites

| Location | Form | Completed hits | Values seen | Expression |
|---|---|---:|---|---|
${rows.map(s=>`| ${s.file}:${s.line}:${s.column} | ${s.form} | ${s.runtime.hits} | ${escape(JSON.stringify(s.runtime.distinctValues))} | ${escape(s.expression)} |`).join('\n')}

## Union-with-void table

There are **16** sites, recorded in full under unionWithVoid and not judged.

| Location | Form | Type | Expression |
|---|---|---|---|
${data.unionWithVoid.map(s=>`| ${s.file}:${s.line}:${s.column} | ${s.form} | ${escape(s.type)} | ${escape(s.expression)} |`).join('\n')}

## Commands and environment

All long-running commands had a timeout and were checked while active. Test output went directly to log files. The final commands were:

\`\`\`bash
git fetch origin
git switch -c typescript/stage3-void-reads origin/main
export GOPROXY='https://proxy.golang.org|direct'
timeout 840 bash cloud/setup.sh > /tmp/void-reads-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
timeout 120 git fetch origin typescript/stage3-onediag-evidence
timeout 180 npm ci --prefix stage3/api > /tmp/void-reads-api-install.log 2>&1
timeout 180 git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/void-reads-upstream > /tmp/void-reads-clone.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/void-reads-upstream timeout 600 bash stage3/apply.sh /tmp/void-reads-tree > /tmp/void-reads-apply-retry.log 2>&1
timeout 300 node stage3/census/void-reads/census.cjs /workspace/adamic /tmp/void-reads-tree /tmp/void-reads-census stage3/census/void-reads/void-reads-2026-10-19.json --adaptations > /tmp/void-reads-census.log 2>&1
timeout 240 node stage3/census/void-reads/mutant.cjs /workspace/adamic /tmp/void-reads-tree /tmp/void-reads-mutant-final stage3/census/void-reads/void-reads-2026-10-19.json stage3/census/void-reads/mutant-proof.json > /tmp/void-reads-mutant.log 2>&1
timeout 120 cp -a --reflink=auto /tmp/void-reads-tree /tmp/void-reads-runtime
timeout 60 node stage3/census/void-reads/instrument.cjs /tmp/void-reads-runtime stage3/census/void-reads/void-reads-2026-10-19.json > /tmp/void-reads-instrument.log 2>&1
VOID_READS_COVERAGE=/tmp/void-reads-coverage NODE_OPTIONS='--require=/workspace/adamic/stage3/census/void-reads/observe.cjs' timeout 720 bash stage3/oracle/run.sh /tmp/void-reads-runtime /tmp/void-reads-oracle --workers 4 --limit-seconds 540 > /tmp/void-reads-oracle.log 2>&1
timeout 30 node stage3/census/void-reads/counter-control.cjs /tmp/void-reads-counter-control stage3/census/void-reads/counter-proof.json > /tmp/void-reads-counter-control.log 2>&1
timeout 90 node stage3/census/void-reads/verify.cjs /workspace/adamic /tmp/void-reads-tree stage3/census/void-reads/void-reads-2026-10-19.json > /tmp/void-reads-verify.log 2>&1
PYTHONDONTWRITEBYTECODE=1 timeout 240 bash stage3/oracle/run.sh /tmp/void-reads-tree /tmp/void-reads-api-control --tests 'Public APIs' --limit-seconds 60 > /tmp/void-reads-api-control.log 2>&1
timeout 30 node stage3/census/void-reads/merge-coverage.cjs stage3/census/void-reads/void-reads-2026-10-19.json /tmp/void-reads-coverage /tmp/void-reads-oracle/report.json /tmp/void-reads-api-control/report.json > /tmp/void-reads-merge.log 2>&1
node stage3/census/void-reads/report.cjs
\`\`\`

The first apply attempt failed because the product cache was cold and ADAMIC_TYPESCRIPT_SOURCE was unset. The pinned checkout above fixed that. An early census against the unfinished tree rejected 78 roots; the completed product has 79. An initial adaptation diagnostic comparison was rejected because its baseline diagnostics were collected after exploratory queries; the final run collects both diagnostic sets at the same phase and introduces none. These rejected attempts were not counted as proofs.

Setup timing lines: Node 0.074s; Go 0.097s; clang 0.621s; markdown dependencies 1.328s; submodules 231.481s; shared cache 241.970s; Go build 632.756s; cache warm 633.246s; done 633.604s. nproc=5, CPU quota=4, Node v24.19.0, Go 1.27.1, clang 20.1.8. Setup's printed environment path was /workspace/adamic-tools/env.sh and was sourced. The full Go test gate was not run for this measurement-only unit; setup built all Go packages.

The scratch mutant appends a void-returning function and a binding of its call result to the first root. mutant-proof.json records **491 -> 492**, with the sole added read at **src/compiler/_namespaces/ts.moduleSpecifiers.ts:6:25**, no lost sites, and the union count unchanged. The scratch file is restored in a finally block and its hash is rechecked by verify.cjs.
`;
fs.writeFileSync(path.join(directory,'README.md'),text);
