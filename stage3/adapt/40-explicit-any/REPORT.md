Built: three explicit-any owner corrections, a per-site ledger, and mechanical proofs; partial, 207 tokens remain.
Base: ef3d907ecdc4c771b016f7d9c52372def057a340; pipeline: 8728405135d329efc12c837a7a6c293234abbe1c.
Commands/results: full build and stock checker pass; 10 JavaScript files identical; oracle 106,367 passing, baseline.diff empty.
Mutants: ten caught mutations covering source/type guards, census, JavaScript, API, idempotence, checker, and real-input baseline comparison.
Not completed: the remaining 207 owner contracts, units 10/20 composition, Adamic adapted-program re-census, and the uncached native gate.

## Completed contracts

Only src/compiler/core.ts is adapted. length and toOffset infer their array
item type as T; neither reads an item. hasProperty takes object, the common
contract of the actual arrays, functions, AST nodes, package metadata and
formatting objects whose key ownership it tests. All statements, imports and
trivia are preserved. No unknown replacement, runtime narrowing, or consumer
edit is made. See evidence/source.patch and README.md.

The census before is 210 tokens, 198 lines, 28 files. After it is 207 tokens,
195 lines, 28 files. The original 210 site locations and source lines match
stage3/census/data/sites.json exactly. evidence/sites.json lists each changed
or declined site. The declined sites were not fully analyzed; the ledger names
the unresolved contract work rather than claiming it cannot be done. I stopped
at the three owners whose types and consumer contracts I established. A count
reduction does not establish that the remaining program is Adamic.

## Verification

The exact commands are in README.md; these are the observed results:

- bash cloud/setup.sh: Go go1.27.1 ready at 1s; clang 20.1.8, Node v24.19.0
  and submodules ready at 1s; cache warm at 109s; total 109s. nproc=5;
  cgroup cpu.max=400000 100000.
- npm run build on the pristine prepared tree and on the final adapted tree:
  both exit 0, compiler/services/server/watch/typings and test projects checked.
  Upstream's locked build checker is 5.9.3; the independent stock checker is
  6.0.3, with the complete original compiler tsconfig and config-file context.
  Both stock pre-emit snapshots contain zero diagnostics.
- verify.cjs after: all 10 .js/.mjs/.cjs outputs under built have the same
  paths and SHA256 bytes as before. All 714 declaration files are compared;
  two differ by exactly the two emitted owner signatures, mechanically
  reconstructed from pristine declarations. See evidence/declarations.patch.
  Public typescript.d.ts is byte-identical to its pristine build and matches
  the public API reference with upstream's newline normalization.
- Final adapter first pass reproduces the final source exactly. Second pass
  reports zero owners/removed tokens and changes zero bytes.
- stage3/oracle/run.sh with default runners, four workers, --light=false and
  --lint=false: exit 0, 106,367 passing, zero failing/pending, 319.623 seconds.
  baseline.diff is zero bytes; no public API snapshot exception was needed.
- Following the baseline mutant, its test input was restored byte for byte;
  git diff --name-only in the upstream tree lists only src/compiler/core.ts.
  The checker and whole emitted-output comparison passed again.

Raw snapshots, output hashes, commands, phase logs and reports are in evidence.
An earlier overlapping exploratory build/oracle invocation was stopped; it is
not counted as a passing verification. Final comparison builds and oracle runs
were sequential.

## Every exposed consumer error

The final declaration edits introduce zero consumer errors:
evidence/consumer-errors.json is an empty array. An exploratory MapLike<T>
parameter for hasProperty was rejected by actual consumer checking. All 18
TS2345 diagnostics are preserved in evidence/exploratory-consumer-errors.json
and the full build log. Each says its argument lacks the string index signature
required by MapLike<unknown>:

| Consumer | Line:column | Actual argument |
| --- | --- | --- |
| compiler/commandLineParser.ts | 1716:130 | CommandLineOption |
| compiler/core.ts | 1354:29 | T extends object |
| compiler/debug.ts | 372:30 | AnyFunction |
| compiler/factory/nodeFactory.ts | 6138:54 | SourceFile |
| compiler/factory/nodeFactory.ts | 6400:29 | T extends Node |
| compiler/factory/nodeFactory.ts | 6400:57 | T extends Node |
| compiler/moduleNameResolver.ts | 362:22 | PackageJson |
| compiler/moduleNameResolver.ts | 953:25 | object |
| compiler/moduleNameResolver.ts | 2711:95 | object |
| compiler/moduleNameResolver.ts | 3103:26 | never[] or PackageJsonPathFields |
| compiler/utilitiesPublic.ts | 1488:24 | readonly T[] |
| compiler/utilitiesPublic.ts | 1488:53 | readonly T[] |
| services/formatting/rules.ts | 474:54 | FormatCodeSettings |
| services/formatting/rules.ts | 478:54 | FormatCodeSettings |
| services/formatting/rules.ts | 482:56 | FormatCodeSettings |
| services/formatting/rules.ts | 486:56 | FormatCodeSettings |
| services/formatting/rules.ts | 490:56 | FormatCodeSettings |
| services/stringCompletions.ts | 1394:33 | object |

The correction belongs at hasProperty's own declaration: it tests ownership,
not dictionary values. No consumer was fixed or silenced.

## Mutants and their catches

| Mutation run | Dedicated catch |
| --- | --- |
| JavaScript bytes return 1 to return 2 | Per-file byte hash comparison |
| Drop emitted JavaScript file | Exact output file-set comparison |
| Change real length body default 0 to 1 | Exact source edit reconstruction |
| Append explicit any to the real adapted source text | Stock AST census count |
| Change internal API length<T> to length<U> | Exact declaration reconstruction |
| Change real helper body before adapting | Adapter rejects unreviewed uses |
| Change original length array item type to string | Adapter rejects unreviewed signature |
| Add a source byte on the second pass | Idempotence byte comparison |
| Return string from real length function via checker host | Stock checker TS2322 |
| Append number = string declaration to real forInStatement1 test input | Stage 3 oracle: three failing tests, four baseline differences, 1501-byte diff |

The independent census alignment check also catches removal of one supplied
site. The baseline mutant's install/build both exit 0; only its test oracle
fails. It is a real fixture-input mutant, not an Adamic compiler or sanitizer
mutant. Its filtered invocation is --runners=compiler --tests=forInStatement1,
which uses one worker. Full passing and filtered failing reports are separate.

## Limits and handoff

This is deliberately a partial unit. I did not complete the remaining owner
contracts. There is no claim of no explicit any left, native tsc acceptance,
implicit-any elimination, or checked JSON/host inputs. No untyped input was
changed to unknown. The complete declined ledger is the handoff for further
work, not evidence of an impossibility.

Units 10 and 20 were fetched but not composed into this proof. No production
compiler, fixture bucket, fixture harness, stage 3 pipeline, or other adaptation
was edited. The full uncached Go/ASan gate and the adapted-tree Adamic census
with its sounder regex declarations were not run. The full upstream stage 3
suite was run. Source maps and build metadata are outside the JavaScript byte
claim. Every emitted JavaScript file is inside it.

## Follow-up classification

All remaining 207 tokens are classified into 33 disjoint semantic patterns in
CLASSIFICATION.md and evidence/classification.json. Each has an exact location,
observation and proposed rule. The largest class is JSON/config (38), followed
by phantom brands (36), enum display reflection (23), timer plumbing (16) and
incomplete constructors/initialization (11). There is no unclassified other.

The largest class requires unknown and checked narrowing at external boundaries,
then real validated value types internally. Existing raw config property reads
and writes do not prove those shapes. Extra runtime checks would change the
JavaScript, so the requested conditional source edit is declined. No further
adapter, TypeScript source, API reference or consumer edits are made.

classify-check.cjs compares all IDs, locations and source snippets against the
prior stock AST census, independently partitions every site exactly once, checks
class counts and their total, and hashes every compiler source against the prior
after snapshot. Five mutants remove a site, duplicate class membership, alter a
count, alter a location and append a comment to a scratch copy of real source; each must fail its own check.
See evidence/classification-check.log. These prove ledger completeness and
source preservation, not the soundness of proposed replacement types.

The earlier whole-build/oracle/idempotence proof still applies to the unchanged
three-owner adapter. It was not rerun as a new adaptation proof in this follow-up.
The remaining count is unchanged at 207, and no new consumer errors are elicited
because there are no new owner edits. Full native and composition limits above
remain.

## Type-only brand class

Applied all 36 brand sites after classification, on the combined 10+30+40
tree with 20 excluded. PROGRESS.md and evidence/brands record the current
proof; the 207-token statements above are historical. No runtime checks or
consumer changes were made. Census is now 171.

Brand-class full oracle: 106,367 passing, zero failing/pending, baseline.diff
0 bytes. Real API-reference mutant: one failed test and one baseline diff,
then exact restoration and zero remaining baseline differences. See
evidence/brands/oracle-report.json and baseline-restoration.json.

## Type-only enum display class

Applied 23 real enum-map/namespace projection types after the brand commit.
Census 171 -> 148; JavaScript identical; zero new consumers; two internal
declaration changes precisely explained; public API unchanged. Full oracle
106,367 passing, zero failures, empty baseline.diff. The enum-domain checker
mutant and real API-baseline mutant both fail their dedicated checks. Evidence
is in evidence/enum-display and current progress is in PROGRESS.md.

## Type-only diagnostic argument class

Applied six sites and related owner-only argument alias/generic constraints.
Initial rule exposed 16 consumer errors, all recorded verbatim. Final rule:
zero new errors, all emitted JavaScript identical, mechanically limited API
exception, idempotent, census 148 -> 142. Full 10+30+40 oracle without 20:
106,367 passing, zero failures, empty baseline diff. Real checker and API
reference mutants are caught; reference restored exactly. See PROGRESS.md and
evidence/diagnostic. The remaining 142 sites are classified but unfinished;
JSON/config and host-extension checks remain held for separate authorization.
