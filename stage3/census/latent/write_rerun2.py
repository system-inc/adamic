"""Render the second census, tonight's ratio, and the readonly-views declaration ledger."""
import json
from pathlib import Path
ROOT=Path(__file__).resolve().parent;data=json.loads((ROOT/'REPORT.json').read_text());base=data['baseline'];runs=data['runs']
LABEL='measured on a checker-rejected program'
def esc(s):return str(s).replace('|','\\|').replace('\r\n','<br>').replace('\n','<br>')
text=f'''Built: the same latent measurement on cumulative-2 and cumulative-2 plus newest nested functions, with adaptations 10/20/30/45/46.
Tonight: both configurations have {runs[0]['checker_clean_ratio']['fraction']} checker-clean files ({runs[0]['checker_clean_ratio']['percent']:.2f}%), up from {base['checker_clean_ratio']['fraction']}; measured on a checker-rejected program.
Commands/results: both builds and runs exit 0; checker total {runs[0]['checker_total']} each, delta {runs[0]['delta_vs_previous_cumulative']['checker']:+d} against previous cumulative, measured on a checker-rejected program.
Mutants: extra NotYet on the real corpus, output guards, body scope/misattribution, checker ratios, variance ownership and report artifacts tested and caught.
Limits: zero own-file diagnostics is the score, not project acceptance; first lowering error per unit remains; no native or full-gate correctness claim.

# Tonight's number

**{runs[0]['checker_clean_ratio']['fraction']} = {runs[0]['checker_clean_ratio']['percent']:.2f}%** for both configurations,
**measured on a checker-rejected program**. The previous cumulative run was
{base['checker_clean_ratio']['fraction']} = {base['checker_clean_ratio']['percent']:.2f}%. The increase is seven files.
A checker-clean file has zero diagnostics attributed to its own source in the same
whole-project checker run as before. The 78-root project still has checker errors;
this score does not imply that any file's complete imported program compiles.
No message has been sent to @system_adamic or another worker.

# Method and provenance

All checker totals, per-file/reason counts, ratios, unit counts, variance counts and
deltas here and in JSON are **measured on a checker-rejected program**. The method
is unchanged: one whole-project checker run, diagnose bodies by exact byte overlap,
skip/count top-level functions with their own body diagnostics, measure every other
function and every top-level statement independently, scan refusal syntax past its
first failure, and deduplicate `(kind, location, reason, exact text)` across attempts.
Checker-diagnosed dependencies remain SkippedDependency boundaries, separate from
NotYet/Refused. Ordinary errors/panics remain separate. Source declarations were
added as metadata only; the lowering and eligibility algorithms are unchanged.

The original driver, production Load/LoadOverlay, and lower.Lower output path remain
disabled in this scratch overlay binary. Every corpus run enables the nil-IR and
loader guards; no emitter/native backend is called. Delivery changes stay under
stage3/census/latent/. Scratch integration commits and branches are never pushed.

The previous cumulative head is `{base['head']}`, delivered by
`{base['delivery_commit']}`. Its unchanged raw data is `data/cumulative.jsonl.gz`;
the full previous report is archived as `data/previous-REPORT.md` and
`data/previous-REPORT.json.gz`. Its 10/20-only source hashes remain there.
This rerun's changed adapted-source hashes are recorded in REPORT.json. Deltas against
the previous run combine the new adaptations, checker rulings and compiler commits.
They cannot isolate each adaptation's contribution. Exact reason strings may shift
with changed types; exact locations may shift with inserted source lines.
The nested-vs-cumulative-2 delta uses identical source bytes and differs by the newly
merged nested-functions commits and their scratch conflict resolution.

The remote's default fetch refspec tracks only main. Requested branches were fetched
explicitly. Main at the fetch snapshot has no stage3/adapt directory, so the named
adaptations were taken from their branches. Flags at 246ecc07 includes 4097384;
nested-functions' newest fetched commit is 59fe216b. Cumulative-2 retains the previous
cumulative's older nested-functions implementation; the second run adds its new commits.

| Configuration | Resolved scratch head | Never-pushed branch |
|---|---|---|
'''
for r in runs:text+=f"| {r['name']} | {r['head']} | {r['branch']} |\n"
text+='\nPinned newly merged commits:\n\n| Branch | Commit |\n|---|---|\n'
for ref,sha in runs[1]['pinned_commits'].items():text+=f'| {ref} | {sha} |\n'
text+='''
The starting cumulative already contains taste at aa896b5d, namespaces at ce8a2acf,
and nested functions at b15216da, plus the base pipeline and original census.
All requested commit ancestries were checked before measurement.
Dependency pins remain cohere 715ba94f and typescript-go 8d550c83. Canonical dependency
paths in scratch Go workspaces share the warm cache; this does not change source or
checker options. Exact generated overlay and binary hashes are in JSON.

Scratch fallthrough conflicts preserve taste's label/continue support and enum
ErasableSyntaxOnly=false, while removing NoImplicitReturns/NoFallthroughCasesInSwitch
as ruled by the fallthrough branch. IR retains label and depth fields for the two
representations. The newest enum merge retains namespace initialization (which
already calls enum initialization) and both namespace/enum refusal hooks. The
nested merge combines undefined-only/narrowed-field storage with void storage and
retains the TypeScript notice. Whole compiler/native correctness was not tested;
these integrations are used only by the disabled-output measurement driver.
`data/rerun2/` preserves conflict-resolution source, merge logs, final integration
diffs, and each resolved branch's head. Unknown/new conflicts would require review.

# Adapted tree

`apply.sh` from cumulative-2 discovered setup plus exactly 10, 20, 30, 45 and 46,
in numeric order, on TypeScript 6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`. Both compilers use identical adapted bytes.
There are still exactly the same 78 compiler roots, including the generated diagnostic
map. Adaptation 46 is a deliberate fork fix for the empty single-quoted pragma argument,
not an erased soundness assertion. This rerun measures it without revalidating its oracle.
`data/rerun2/apply.log.gz` and `patch-set.md` preserve the exact adapter outputs.

'''
text+=(ROOT/'data/rerun2/patch-set.md').read_text().replace('# Patch set','Incremental source edits')+'\n'
text+='''
# Checker, lowering and variance totals

Every number is **measured on a checker-rejected program**. N/R are unique NotYet/
Refused sites, not attempted-unit counts. Dependency skips are excluded from N/R.

| Configuration | Checker | Clean files / 78 | NotYet | Refused | Variance Refused subset | Functions attempted | Bodies skipped | Dependency skips | Errors/panics |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
'''
for r in [base,*runs]:
 c=r['lowering_counts'];text+=f"| {r['name']} | {r['checker_total']} | {r['checker_clean_ratio']['fraction']} | {c['NotYet']} | {c['Refused']} | {r['variance']['total']} | {r['functions_attempted']} | {r['bodies_skipped']} | {c['SkippedDependency']} | {c['error']}/{c['panic']} |\n"
text+='''
Each new configuration records one ordinary error whose text says the checker gave
`src/compiler/parser.ts:1472:9` a declaration no symbol. It exposes no structured
location and is excluded from N/R. No panic was recorded.

# Deltas

Every delta is **measured on a checker-rejected program**. More functions become
eligible when their body diagnostics disappear, so added lowerer findings can be
newly exposed failures rather than regressions. The report makes no causal attribution
of the combined source/compiler delta. The optional-declaration revision is narrower
than the previous revision, and the indexed/regex adaptations discharge additional
obligations, so this is not a pure feature comparison against identical sources.

| Configuration relative to previous cumulative | Checker delta | Clean-file delta | NotYet delta | Refused delta | Variance delta |
|---|---:|---:|---:|---:|---:|
'''
for r in runs:
 d=r['delta_vs_previous_cumulative'];text+=f"| {r['name']} | {d['checker']:+d} | {d['checker_clean_files']:+d} | {d['counts']['NotYet']:+d} | {d['counts']['Refused']:+d} | {d['variance']:+d} |\n"
d=runs[1]['delta_vs_cumulative_2'];text+=f"\nNewest nested commits relative to cumulative-2: checker {d['checker']:+d}, clean files {d['checker_clean_files']:+d}, NotYet {d['counts']['NotYet']:+d}, Refused {d['counts']['Refused']:+d}, all **measured on a checker-rejected program**.\n"
text+='\nNewly checker-clean files (both configurations, **measured on a checker-rejected program**):\n\n'
for f in runs[0]['delta_vs_previous_cumulative']['new_checker_clean_files']:text+='- '+f+'\n'
text+='\nNo previously checker-clean file loses that status.\n\nAll 26 checker-clean files in both configurations:\n\n'
for f in runs[0]['checker_clean_files']:text+='- '+f+'\n'
text+='''
# Per file

Every cell is **measured on a checker-rejected program**. Checker columns count all
primary diagnostic chains in that file. N/R columns count unique actual diagnostic
sites. JSON includes per-file checker/lowering deltas, reasons, eligibility and exact
finding texts. Zero checker diagnostics are the ratio criterion above.

| File under src/compiler | Previous checker | Cum-2 checker | New nested checker | Previous N/R | Cum-2 N/R | New nested N/R |
|---|---:|---:|---:|---:|---:|---:|
'''
for f in data['source']['files']:
 name=f['file'];rows=[r['per_file'][name] for r in [base,*runs]]
 text+='| '+name.removeprefix('src/compiler/')+' | '+' | '.join(str(r['checker']) for r in rows)+' | '+' | '.join(f"{r['NotYet']}/{r['Refused']}" for r in rows)+' |\n'
text+='''
# Variance refusals for readonly views

Every number is **measured on a checker-rejected program**. The subset contains
mutable invariance, method-parameter bivariance, contravariant overrides, widened
return overrides and readonly-to-mutable overrides. Classification uses the exact
rule tags/fixes and named override reasons. Enum domain, nominal ancestry, casts
and type predicates are excluded. There are 440 mutable-invariance and 54 method-
parameter-bivariance sites in each run; the other variance families are zero.
Previous cumulative had 422 mutable-invariance and 65 bivariance sites (487 total).

The owning declaration is the nearest named AST declaration containing the refusal's
actual diagnostic byte position: function/method, variable/parameter, property,
interface/type alias, class, enum or namespace. Names are qualified by enclosing named
declarations. The attempting top-level unit is retained separately, so a dependency
failure is charged once to its source declaration, not once to every caller. Owning
declarations can contain several distinct sites/reasons. Arrow/function expressions
without a source name belong to the nearest enclosing named declaration. File scope
is an explicit fallback when no named declaration contains a site.

Stock typescript@6.0.3 independently parses all 78 adapted sources. Every Go overlay
declaration span/name/kind/location matches that AST ledger. The report audit calculates
owners from the stock spans and UTF-16 diagnostic locations, independently of the
reported ownership. Each run has 494 sites in 304 owning declarations. Per-owner
findings, family counts, locations, exact text and attempting units are in JSON.
The previous run did not record declaration spans; its variance total is recounted,
but no previous per-owner breakdown is invented.

| Owning declaration location | Qualified declaration | Cum-2 variance sites | New nested variance sites |
|---|---|---:|---:|
'''
owners=set(runs[0]['variance']['per_declaration'])|set(runs[1]['variance']['per_declaration'])
for where in sorted(owners):
 entries=[r['variance']['per_declaration'].get(where) for r in runs];entry=next(e for e in entries if e)
 text+=f"| {where} | {esc(entry['declaration']['qualified_name'])} | {entries[0]['count'] if entries[0] else 0} | {entries[1]['count'] if entries[1] else 0} |\n"
text+='''
# Per checker reason

Every number is **measured on a checker-rejected program**.

| TS code | Previous | Cum-2 | New nested |
|---|---:|---:|---:|
'''
for reason in sorted(set().union(*(r['checker_per_reason'] for r in [base,*runs]))):text+='| '+reason+' | '+' | '.join(str(r['checker_per_reason'].get(reason,0)) for r in [base,*runs])+' |\n'
text+='''
# Per NotYet/Refused reason

Every number is **measured on a checker-rejected program**. These are the exact
reason families as before, including concrete types. SkippedDependency and ordinary
errors are preserved separately. Their exact text and every site are in JSON.

| Reason | Previous cumulative | Cum-2 | New nested |
|---|---:|---:|---:|
'''
for reason in sorted(set().union(*(r['per_reason'] for r in [base,*runs]))):text+='| '+esc(reason)+' | '+' | '.join(str(r['per_reason'].get(reason,0)) for r in [base,*runs])+' |\n'
text+='''
# Commands, checks and mutants

Required toolchain setup ran again: Go/clang/Node ready in 0s, submodules in 1s,
build cache warm and total 24s. `nproc` is 5, cgroup CPU quota four CPUs.
Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup/fetch logs are preserved.
All test output was redirected to logs, never piped.

```sh
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/tsc-latent2-adapted > /tmp/latent2-apply.log 2>&1 # in the resolved cumulative-2 tree
python3 stage3/census/latent/run_comparisons.py /workspace/adamic /tmp/tsc-latent2-adapted /tmp/latent2-runs /tmp/latent2-worktrees.json > /tmp/latent2-comparisons.log 2>&1
python3 stage3/census/latent/rerun2.py /tmp/latent2-runs /tmp/tsc-latent2-adapted > /tmp/latent2-summary.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/census/latent/declarations.cjs /tmp/tsc-latent2-adapted /tmp/latent2-manifest.json /tmp/latent2-declarations.json > /tmp/latent2-declarations.log 2>&1
python3 stage3/census/latent/audit_rerun2.py /tmp/tsc-latent2-adapted /tmp/latent2-declarations.json > /tmp/latent2-report-audit.log 2>&1
python3 stage3/census/latent/audit_corpus.py /tmp/latent2-runs /tmp/tsc-latent2-adapted cumulative-2 > /tmp/latent2-corpus-audit.log 2>&1
```

The stock AST oracle initially failed because NODE_PATH pointed at /root's cache;
using the actual /home/agent cache fixed it. The scratch merge helper initially lacked
gofmt on PATH; sourcing the toolchain and completing the same resolutions fixed it.
Neither failed attempt produced census results. Logs preserve the successful final runs.

Report mutants alter source hashes, checker totals/per-file attribution, zero-file
membership, the ratio numerator/denominator, NotYet/reason totals, skipped-body counts,
variance totals, owner metadata/counts, checker deltas and measurement labels. A raw
AST declaration-end mutant is caught only by the independent stock-span check.
All named mutants fail the intended check. `data/rerun2/report-audit.log` records them.

The real-corpus overlay-only NotYet at binder.ts:330:1 adds one site in binder.ts
and leaves every other file's entire record and all unit eligibility unchanged.
The baseline contains no mutant. Mutant audit JSON and compressed raw evidence are
preserved. The prior synthetic body-scope/misattribution and IR/loader guard mutants
were rerun with this instrumentation. Overlay vet and Python/Node syntax checks pass.
The older six-configuration report/audits remain archived as historical evidence.

# Limits

Measurement cannot emit native output. Lowering stops at its first returned error
within a unit; the refusal pre-scan sees additional syntax failures. Generic declarations
are attempted without invented specializations; registration is best effort. Diagnosed
dependency bodies remain skipped boundaries. Final module order, ownership, initializer
readiness, specialization completeness and backend passes are not validated. Native
runtime equivalence, adaptation oracle suites and the full Adamic gate were not rerun.
The ratio is zero own-file checker diagnostics out of 78, not native tsc readiness.
'''
(ROOT/'REPORT.md').write_text(text)
