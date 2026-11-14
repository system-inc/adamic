# Resolution-only binding view

Implementation: `61636df700acf25afe229345128bfa4c193fda23`, exactly one commit on
`7534c1d7575db02bb7b6512119e0685987b7e333`. This commit was created before the
proof runs; the full lane and every recorded runner identify that exact source.
No implementation input is changed after testing.

The combined diagnostics overload candidate is excluded following the reported
full program census red on typescript/program-full-census: overload 1's
getDiagnostics parameter cannot be served by the implementation parameter.
This change makes no claim about that program lowering and does not edit
program.ts or amend 95-checker-axolotl-h.

## Scope and ownership

Only stage3/adapt/99-correctness-binding-views and one resolution-only ledger
line change relative to base. The first owner grep is retained. No prior whole
owner pin claims the edited resolution declarations. Complete text/SHA-256 pins
cover ResolutionCache, GetResolutionWithResolvedFileName and
createResolutionCache, the inserted view, and the complete source file.
The new helper uses the established `/** @internal */ export interface` form.
The original ResolutionWithResolvedFileName interface stays unchanged.

The candidate differs from baseline only in src/compiler/resolutionCache.ts.
All 747 source files are enumerated; program.ts is byte-identical. The adapter
and mutant have no program.ts dependency: controls pass with that file absent.
They write only resolutionCache.ts. The mutant restores its exact original.

## Source proofs

| Proof | Result |
| --- | --- |
| Full resolutionCache.ts census | 221 -> 217 raw; 133 -> 129 unique; four gone, zero new, zero changed |
| Target sites | Baseline 774, 1053, 1056, 1471; all four exact refusal_scan/Refused identities gone |
| Restoring mutant full census | All four return; empty baseline/mutant delta |
| Stock tsc 6.0.3 | 79 roots plus scout Node ambient inputs; 0 -> 0 diagnostics, 0 new |
| Adamic types | 79 sorted roots; exit 0, 0 diagnostics; binary revision is the implementation commit, modified=false |
| JavaScript | All 604 files, 1,208 emits with comments retained/removed, byte-identical |
| Adapter controls | Drift, duplicate, partial-owner and mixed-line-ending inputs refused before writes; LF/CRLF preserved |
| Replay and mutant scope | Zero-write replay; apply/restoration each writes only resolutionCache.ts; no program.ts dependency |
| Generated files | Canonical hashed inputs unchanged; no regeneration |
| Ledger | One type-only line for these four sites, #ty5myjt |

The full census uses CommonJSBundle, LATENT_FULL=1,
LATENT_ASSERT_NO_OUTPUT=1 and one exact LATENT_ONLY_FILE. All streams complete
with relation_report and exit 0. This proof includes lowering-phase findings;
it is not a scan-only result. No program census is claimed.

Both adapted checkouts run npm ci and upstream's unmodified Hereby dts. The two
public outputs are byte-identical:

| File | Shared baseline/candidate SHA-256 |
| --- | --- |
| typescript.d.ts | f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516 |
| tsserverlibrary.d.ts | cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd |

## Normal apply and whole lane

Normal apply passes. Its complete 747-file source tree matches the measured
candidate; program.ts is unchanged. Replay on that output writes zero files.

Whole unfiltered lane: **PASS**, 106,366 passing, exactly one sanctioned public
API failure, zero pending. The API guard accepts 273 sanctioned declarations.
Workers remain 8, runners is all, tests is null. Install/build exit 0; the
upstream tests/oracle exit 1 solely for the sanctioned row. The lane itself
exits 0. Expectations, workers, sanctions and filters are unchanged.

The lane execution records exactly `61636df700acf25afe229345128bfa4c193fda23`. The tested product key is
`a293afb5e9a03713b44910a7edf404fa36d7011b24e934c468e3a8c197fe5245`. No implementation input changes after these
proofs. Commands, complete logs and results are retained under lane/ and proofs/.
The implementation is pushed as typescript/binding-views-resolution-1; evidence
is fast-forwarded on typescript/binding-views-resolution-evidence.
Nothing is submitted to Loom.
