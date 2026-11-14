# Readonly writeFile view — #b8ffarm

Landing: **2bbc91af4a7dddbf31e14ddb5153ccd7acc4db3c** on `typescript/write-file-view-1385-1`, exactly **one commit** on **8e3a784aa5a601f57342c07b68f145fe56252c21**. Evidence: `typescript/write-file-view-1385-evidence`.

| Proof | Result |
|---|---|
| Ownership | No existing whole-owner pin; new 99-correctness-write-file-view; no public row |
| Contract | PASS: whole text/SHA, LF/CRLF, 15 drift/half/missing/duplicate/overload/moved/parse/mixed/bare-CR refusal controls; bytes/mtime preserved; zero-write replay |
| Node witness | PASS on real whole-owner pins: receiver-sensitive calls, getter, later reassignment, exact identity and own .call |
| Full single-file CommonJSBundle sys.ts census | **59 → 58** unique findings: **0 new, 1 gone, 0 changed** |
| Restoring mutant | **59**, exactly main again; original target returns |
| Stock tsc 6.0.3 | Actual CLI, all **79 roots**, main/candidate/mutant **exit 0** |
| adamic types | Production CLI, all **79 roots**, main/candidate/mutant **exit 0** |
| Emitted JavaScript | **604 source files**, **1,208 pairs / 2,416 emits**, comments kept/removed; only owner alias/read redirect differs; every other file byte-identical |
| Declarations | Separate upstream npm ci + hereby dts builds pass; typescript.d.ts and tsserverlibrary.d.ts byte-identical; exact diffs empty |
| Fresh whole apply | **PASS** inside the whole lane; every resulting src byte equals the measured candidate |
| Full unfiltered stage3/lane | **PASS**, **106,366 passing**, **1 sanctioned API failure**, **0 pending**, all runners/eight workers/no filter |
| Post-lane replay | **0 writes**, identical bytes and mtime |
| Carried lane count | **2,200 captures / 4,627 calls**, both sides, **0 added calls**; source hashes match the original counted trees |

## Exact change and ownership

Only patchWriteFileEnsuringDirectory's capture changes:

```ts
const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;
const originalWriteFile = writeFileView.writeFile;
```

Every other byte in the whole owner is pinned and unchanged. The readonly view
is the same object, not a copy or wrapper. The getter is read once with the same
receiver before the existing sys.writeFile reassignment. The exact function,
including its own .call behavior, is preserved; originalWriteFile.call(sys, ...)
remains unchanged. There is one extra local reference assignment per capture,
no extra function call, no new function identity and no public declaration edit.

The first search covered owners.json, targets.json and every adapt.cjs under
stage3/adapt. The timer-closures hit is a call in its TestServerHost owner, not
a pin of this function. 41, 93 and crypto own separate declarations/getNodeSystem.
[ownership.json](ownership.json) records the search and decision.

Whole-owner SHA-256:

- Before: `adbb2e3289ea7f35f76dc8f6baf618c27bf9c334a770dd5c83382816e93e76e4`
- After: `9859c28c313f8688c39620ca3aecef4bf51d5aa077f895a50aefe2c2d75ddf25`

The landing contains the new directory's seven files and exactly one
JavaScript-changing, value-identical ledger line citing #b8ffarm. It changes no
compiler, public row, API sanction, oracle, worker count or expected count.
[implementation-inputs.json](implementation-inputs.json) binds the exact eight
frozen implementation files tested by the lane and copied into the one-commit
landing tree.

## Census and restoring mutant

The production Go source is main 8e3a784aa5a601f57342c07b68f145fe56252c21. The same standard guarded census
binary scans each complete private sys.ts, all 79 compiler roots, with:

```text
LATENT_FULL=1
LATENT_ASSERT_NO_OUTPUT=1
LATENT_ONLY_FILE=<tree>/src/compiler/sys.ts
CENSUS_MODULE_FORMAT=CommonJSBundle
```

All headers have diagnostics=[] and checker_rejected:false. The meter's legacy
generic “measured on a checker-rejected program” label does not override that
header. Both no-output guards execute. No successful native compilation of the
remaining frontier is claimed.

[CENSUS.md](CENSUS.md) and [delta.json](delta.json) list every new/gone/changed
finding in full, including phase and attempt context. The sole gone finding is:

```text
kind: Refused
phase: refusal_scan
unit: src/compiler/sys.ts:1383:1
where: src/compiler/sys.ts:1385:31
reason: a method read as a value (writeFile would lose its object, and this with it)
```

Main has 41 NotYet + 18 Refused; candidate has 41 + 17. The restoring mutant
reinstates the exact whole before owner and reproduces main's entire finding
set, including this target. Boundary recorder events are excluded from unique
compiler-finding counts; raw streams retain actual coordinates and all attempts.
No widening or new finding is hidden. [runs.json](runs.json) preserves actual
commands, exits, timings and flags; all raw streams/logs are in evidence/.

## JavaScript, declarations and witness

[emitted-javascript.json](emitted-javascript.json) records every source hash and
both comment-mode emit hashes. The only differing file is sys.ts. Its exact
commented and stripped JavaScript hunks are [here](sys-comments.js.diff) and
[here](sys-stripped.js.diff): only patchWriteFileEnsuringDirectory changes, by
one alias statement and the read through that same alias. No other file differs.
The adapter enforces that exact owner-only emit delta before writing.

Both main and candidate separately ran upstream's npm ci and unmodified
node_modules/.bin/hereby dts. The [typescript.d.ts diff](typescript.d.ts.diff)
and [tsserverlibrary.d.ts diff](tsserverlibrary.d.ts.diff) are empty. Hashes:

- typescript.d.ts: `f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516`
- tsserverlibrary.d.ts: `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd`

[declarations.json](declarations.json) records sizes and identity. The complete
lane's own declaration artifacts match the same hashes as well.

[witness.json](witness.json) records the real pinned bodies executed in Node.
Only the filesystem helper is stubbed; a temporary in-memory hook observes the
private originalWriteFile identity. The hook and stub are not in the adapted
source. The two cases include an own .call changed between invocations.
[contract.json](contract.json) records refusal controls, precise write targets,
line endings, restoring replay, and byte/mtime equality.

## Whole apply, lane and carried counts

The frozen evidence implementation ran unmodified stage3/lane/run.sh, including
a fresh full apply, upstream install/build and all unfiltered tests. The new
patch-table row is one file, two added lines and one removed line. The full lane
returns PASS. Its only failing case is:

```text
unittests:: Public APIs for typescript.d.ts should be acknowledged when they change
```

The only baseline diff is api/typescript.d.ts. The unchanged sanction guard
passes; install/build exit 0, tests/oracle exit 1, and lane exit 0. Actual counts
are 106,366 passing, one failing and zero pending. The eight workers, complete
runners and absence of a filter are retained in the oracle report.
[lane-report.json](lane-report.json), [lane-execution.json](lane-execution.json),
[lane-verdict.txt](lane-verdict.txt) and compressed full logs preserve the receipts.
[post-lane.json](post-lane.json) verifies the frozen inputs, public declarations
and zero-write replay. [source-integrity.json](source-integrity.json) verifies
every src byte against the measured candidate and mutant against main.

The requested count is carried from typescript/price-sys-public at
3ecc182ad4c79a8783e4b673c01d977d046e515e. Both counted sys.ts inputs are
**byte-identical** to this task's main/candidate inputs: 2,200 captures and 4,627
callback calls on each side. [carried-lane-counts.json](carried-lane-counts.json)
binds those identities and the full counted-oracle results; raw counters and
reports are preserved in evidence/. This is a carried source-matched count,
not a claim that the new unmodified whole lane contains counter instrumentation.
The readonly alias adds zero calls; its 2,200 extra local assignments are not
function calls.

## Pins and publication

A fresh detached private worktree of the requested base initialized recursive
shallow submodules. A fresh v6.0.3 upstream clone was confirmed at
050880ce59e30b356b686bd3144efe24f875ebc8 and exported as
ADAMIC_TYPESCRIPT_SOURCE. stage3/api npm ci succeeded. Go/temp/npm caches and
all generated files stayed in writable scratch. Main apply restored a
hash-verified immutable baseline product; the candidate whole apply built the
new keyed product through the entire adaptation sequence.

[provenance.json](provenance.json) records pins, binaries, tool identities and
unchanged guarded files. [landing.json](landing.json) records the one-parent,
one-commit landing tree. Helpers are audit text with this run's fixed scratch
paths. [SHA256SUMS](SHA256SUMS) covers every published evidence file except itself.
No Go source is added under stage3. Nothing is submitted to Loom.
