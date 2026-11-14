# Resolve host 1612 — accepted proof

| Path exception | Counter | Why unreachable; direct check |
| --- | --- | --- |
| getDirectories | 0 | moduleNameResolver.ts:820,825 are in getAutomaticTypeDirectiveNames, outside the resolver closure. Exact path and nodeSystem receiver checked directly. |
| trace | 0 | Fixed options at moduleNameResolver.ts:1755 omit traceResolution; :124 short-circuits before host.trace. Exact string and nodeSystem receiver checked directly. |
| getCurrentDirectory | 0 | Successful resolutions use fixed Node10/None/JavaScript settings (:1750–1764), excluding all current-directory routes. Failed retry's extension mask is 0 (:1897), so both lookup passes skip (:3017,:3023). Zero arguments and nodeSystem receiver checked directly. |

No fourth exception. These are **forwarder checks, not path proofs**. A module
that successfully resolves through nodeSystem.require does not read
getCurrentDirectory. Every possible site, including paths/type roots and
imports/exports, is accounted for in [CHAIN.md](CHAIN.md); both successful
fixture cases independently record zero. The lane never enters this path,
so its path counts are vacuous.

Implementation: **635baacb5dfa6c4d61aa6fb9f26d720552b1520e**, exactly one commit
on ad8ca5a3801d979ca8486b73f3b0ea3cbec0b7d8. Pushed to
`typescript/resolve-host-1612-1` without force. This stacks after the require
result annotation, amending 99's same whole-owner after pin. The initial grep
found whole-owner pins in 93 and 99 and narrower checks in 92. 93 stays earlier
and unchanged; the fresh full apply proves their composition. Measured road:
3ecc182a, stage3/census/price-sys-public/REPORT.md site 10, **#b8ffarm**.
No Loom submission.

## Presence-read inventory

Counts are syntactic callback presence-test/optional-call sites, not ordinary
callee reads. Whole resolver-module sites are included even when outside the
entry's closure. Additional sites are from the symbol-bound conservative
closure. `presence.json`, `resolver-audit.json` and CHAIN.md give the complete
locations, branches and function/edge list.

| Member | moduleNameResolver.ts presence sites (count) | Additional closure sites (count) | Feasible require presence-test sites |
| --- | --- | --- | ---: |
| fileExists | none (0) | none (0) | 0 |
| readFile | none (0) | none (0) | 0 |
| directoryExists | :820 (1; outside closure) | utilities.ts:7819 (1) | 1 |
| realpath | :1975 (1) | none (0) | 1 |
| getDirectories | :820 (1; outside closure); call :825 | none (0) | 0 |
| trace | :124 (1; short-circuited) | none (0) | 0 |
| getCurrentDirectory | :488,2331,2582,2663,2870,2898,2924,2974 (8) | utilities.ts:6609 (1; paths absent) | 0 |
| getGlobalTypingsCacheLocation | :3056 (1) | none (0) | 1 |
| useCaseSensitiveFileNames | no explicit absence/optional-call test (0) | none (0) | 0 |

The boolean row is a copied value, not a function forwarder. Its separate
type/truthiness tests are :512,3425,3426, plus utilities.ts:6482's callback-host
adapter. :512 reads the actual host's type and selected boolean value; the
other routes are excluded by the fixed settings. All value/type accesses,
including the cache boolean at :2463, are classified in the audit. Every one
of the seven optional host members has its own conditional initialization;
useCaseSensitiveFileNames uses `in`, preserving a present false value. The
128-case contract verifies exact member sets and zero present-undefined slots.

## Node fixture counters

| Forwarder | realpath present | realpath absent | Direct check |
| --- | ---: | ---: | --- |
| fileExists | 17 | 17 | path proof |
| readFile | 3 | 3 | path proof |
| directoryExists | 23 | 23 | path proof |
| realpath | 1 | 0 (absent) | path proof when present |
| getCurrentDirectory | 0 | 0 | 1 each case, exact zero arguments, nodeSystem receiver |
| getDirectories | 0 | 0 | 1 each case, exact path, nodeSystem receiver |
| trace | 0 | 0 | 1 when present, exact string, nodeSystem receiver; absent otherwise |
| getGlobalTypingsCacheLocation | 2 | 2 | path proof |

The fixture supplies optional trace/global-cache methods as receiver-checking
own members; realpath and trace are removed for the absent case. Setup ends
before host construction, and slots stay stable throughout resolution. The
actual built compiler's sys.require calls resolveJSModule. Each case contains
a successful package load, a successful relative-file load and a failed package
with an exports map. Per-resolution counts and exact argument logs are in
node-fixture-accepted.json. Parent/candidate output bytes are equal in both
cases; the four .out files are retained. Runtime is Node v24.19.0.

## Complete proof gates

| Proof | Result |
| --- | --- |
| Full single-file sys.ts census, CommonJSBundle | **184 -> 183; exactly 1612:77 gone, new 0, changed 0** |
| Restoring mutant | Exact parent source bytes; sibling ModuleImportResult annotation retained; all 184 findings restored |
| Receiver mutant | `(0, nodeSystem.readFile)(fileName)` fails specifically with `receiver: readFile` |
| Exact forwarders | 8/8 have exact named parameters, no hidden rest (#2t3h708), direct call through nodeSystem, unchanged return |
| Presence | All 128 optional-member combinations pass; absent slots remain absent; false boolean retained |
| Writes after construction | **0** through nodeSystem and production System-typed aliases across 712 source roots; virtual test write excluded |
| Function identity / .length reads | **0 / 0**; computed host reads, host reflection and host destructuring also 0 |
| Stock compiler API | 0 diagnostics for parent and candidate, 79 roots |
| Literal stock tsc CLI | **exit 0**, 79 roots, empty diagnostic output |
| adamic types | **exit 0**, same 79 roots |
| Fresh full stage3/apply.sh | **exit 0** in the whole lane at the implementation commit |
| Source/JavaScript scope | 747 source files; only sys.ts differs. 604 emitted files / 1208 comparisons; JavaScript differs only at host initialization/argument |
| Actual built Node bundle | Exact byte equality after undoing that one site; lane and fixture bundles byte-identical |
| hereby dts | Pass both sides; **all 391 generated .d.ts files byte-identical**, including public and internal bundles |
| Existing 99 crypto source/transpile oracles | Pass, full 747-file inventory |
| Owner contract | LF/CRLF pass; drift, half-state, duplicate and mixed-ending cases refuse before writes |
| Actual final-tree replay | **0 files, 0 writes**, sys.ts bytes and nanosecond mtime unchanged; all 747 files equal the measured source |
| Whole unfiltered lane | **PASS: 106366 passing, 1 existing sanctioned API-baseline failure, 0 pending** |

The lane's sole expected test failure is `unittests:: Public APIs for
typescript.d.ts should be acknowledged when they change`, with baseline diff
exactly `api/typescript.d.ts` and the unchanged 273-declaration sanction.
Oracle/test exit 1 is expected; the landing lane exits 0. No compiler rule,
checker flag, API sanction, filter, test count or expectation changed.

The complete census streams retain every original field and phase. delta.json
compares full finding identities after exact equal-line mapping and explicit
mapping of the one type-only nodeSystem declaration line; byte spans remain
raw observations. Every pre-existing finding except the requested refusal is
identical. The source is pinned by source-manifest.json, final-source.json,
raw parent/candidate sys.ts bytes and the implementation's exact owner texts.

The ledger calls this **JavaScript-changing and value-identical**, citing
#b8ffarm. New function identities are unobserved; member presence, arguments,
receiver and result are preserved, and zero later source-slot writes make
call-time reads equal to the original earlier reads.

See commands.json, provenance.json, presence.json, observers.json, CHAIN.md,
all-declarations.json, raw/, helpers/ and the full lane/ receipts. Superseded
drafts are explicitly excluded in DRAFTS.md. The evidence branch was
checkpoint-pushed, then appended without force after the lane passed.
