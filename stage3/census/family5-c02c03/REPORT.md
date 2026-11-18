# Family 5 C02/C03: stopped at the whole-lane proof

Start: origin/main **fbe5b72ba115a8f3060d4b5229fd578574f4505f**. TypeScript v6.0.3 **050880ce59e30b356b686bd3144efe24f875ebc8**. Recursive submodules were initialized at their recorded pins; see [setup/submodules.txt](setup/submodules.txt).

**No landing was pushed.** The unfiltered C02 whole lane exited **1**, with **106,359 passing, 3 failing, 0 pending**, against the required 106,366 passing and one sanctioned failure. Two failures were 40-second timeouts, in addition to the sanctioned public API acknowledgement failure. The paired censuses and the product test were then terminated with SIGTERM; their incomplete results are not successful proofs.

## Candidates and ownership

C02 local candidate: **28a2e22015d0cf51667fcefa7acb246cab553817**, parent main above. Its only changes are the four files in `stage3/adapt/99-correctness-line-and-character` and one appended ledger line. [Complete candidate patch](c02-candidate.patch). Node remains at types.ts:942; the exact requested member is inserted at :955.

C03 has **no landing commit**. Its external [draft](draft-c03/README.md) and one isolated adapter-contract receipt are retained. Its Type owner is at main's types.ts:6450; C02 would move the declaration to :6451. Its required proofs on C02's tree were not run.

[Pin search](setup/pin-search.txt): adaptation 80 pins Node's heritage, and 71 pins individual members; neither pins the complete Node interface. No complete Node, Type or types.ts pin was found. This uses the same ownership distinction documented by main's 99-correctness-sourcefile-like-position. C02 therefore uses the requested new directory, with the complete owner pinned before and after. Both JavaScript emits are asserted identical before writing; CRLF, exact restoration, replay without writes, and rejection of owner drift, malformed candidate text, duplicate owners and mixed endings passed the scratch contract checks.

The candidate READMEs name adaptation 72's certificate, construction premises and escapes. [Source excerpts](setup/premise-excerpts.txt) and the [scout site list](setup/scout-sites.json) are retained. No internal/ change was made.

## Required proof lines

| Proof | C02 result | Real child exit |
|---|---|---:|
| 1: stock tsc 6.0.3, ordinary, 79 roots, 0 diagnostics | PASS | 0 |
| 1: stock tsc 6.0.3, strict, 79 roots, 0 diagnostics | PASS | 0 |
| 1: adamic types over the 79 roots | PASS | 0 |
| 2: full before/after census of all four files | INCOMPLETE | see below |
| 3: mutant brings the sites back | NOT RUN; only byte restoration in the contract check passed | — |
| 4: stage3/apply.sh whole | PASS | 0 |
| 4: TestProduct_Stage3AdaptedTree, -timeout 45m | INTERRUPTED after the lane failure, before the 45-minute timeout | -15 (SIGTERM) |
| 5: whole stage3 lane, all runners, eight workers | FAIL: 106,359 passing / 3 failing / 0 pending | 1 |
| 6: public typescript.d.ts, cmp | PASS: byte-identical | 0 |

C03 required proofs: **NOT RUN**. Its isolated draft contract check exited 0; that is not a proof of the requested stack.

The complete child receipts include command, cwd, environment, elapsed time, stdout and stderr. Stock checks enforce both zero diagnostics and 79 roots. The public declaration SHA-256 before and after C02 is `65c210e9d7553165606df3ac67544ea7ddcbccf2c322ed924e8e8f4c4622c5fd`.

## Blocking lane failures

[Lane report](c02/lane-receipts/report.json), [oracle report](c02/lane-receipts/oracle/report.json), and complete compressed logs are retained. The lane took 1,864.450 seconds. The npm test command exited 1; its inner test runner reported exit 3. The oracle and validated lane each exited 1.

| Failure | Cause | Measured duration |
|---|---|---:|
| `fourslash tests tests/cases/fourslash/codeFixClassImplementInterfaceNoTruncation.ts fourslash test codeFixClassImplementInterfaceNoTruncation.ts runs correctly` | `Timeout of 40000ms exceeded. For async tests and hooks, ensure "done()" is called; if returning a Promise, ensure it resolves.` | 46,964.015 ms |
| `compiler tests compiler tests for tests/cases/compiler/relationComplexityError.ts "before all" hook for "Correct errors for tests/cases/compiler/relationComplexityError.ts"` | Same 40,000 ms timeout message | 55,109.134 ms |
| `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` | Sanctioned baseline-content failure, `api/typescript.d.ts` | 4,085.339 ms |

The lane ran alongside the product build and two censuses. Contention is a plausible contributor to the timeouts; it was not established as their cause. No quiescent lane rerun was made. The checked-in lane fixes eight workers. This report does not classify either timeout as an adaptation regression or waive it.

## Census coverage and reached stops

Exact environment: `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1`, `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_ONLY_FILE` set to each file's exact absolute path, `LATENT_TIMES=/dev/stderr`, rooted at `src/tsc/tsc.ts`. The entry tool was built from starting main with the repository's scratch measurement overlay. Tool hashes are in [setup/tool-provenance.json](setup/tool-provenance.json).

The scout requires C02 coverage of program.ts (2 sites), utilities.ts (7), parser.ts (4), checker.ts (1). C03 requires checker.ts (14 sites across G03 and G12).

| Run | Result | Real child exit | Wall seconds |
|---|---|---:|---:|
| C02 before, program.ts | Complete, checker-clean, normal relation_report trailer | 0 | 2482.599 |
| C02 before, utilities.ts | Interrupted after lane failure | -15 | 93.624 |
| C02 after, program.ts | Interrupted after lane failure | -15 | 1582.934 |
| Remaining before/after files | Not run | — | — |

The complete baseline program.ts stream contains the two G02 sites at **program.ts:738:57** and **program.ts:739:76**. Its createProgram implementation attempt completed in 969.486 seconds; it was not established to be hung. All partial streams and stderr were kept. The optional debugger attempt failed to start because its Python 3.10 shared library was missing and did not attach.

**No paired finding comparison completed. Target clearance, zero new findings, and reached stops are unestablished.** There is no claimed list of zero reached stops. The prepared comparator requires an old boundary or a later-phase proof before treating an unmatched finding as reached.

The early baseline stock receipt was taken while adaptation was running and reported one missing source-map-support declaration. It is preliminary. Both settled baseline stock configurations and baseline adamic types subsequently passed with zero diagnostics; their receipts are retained separately.

Large logs and streams are compressed losslessly with gzip. Executable orchestration scripts are retained as evidence of the method; unexecuted commands in those scripts are not proof receipts.
