Built: classified candidate unions; added overlapping/open-kind dispatch, own class-data probes, and top-four candidate selector fixtures.
Commits: production hooks 1baa2640; prior handoff 55b3fdef/2bb00f0f; this candidate group follows in Git history.
Commands: stock classification has 0 diagnostics; final lane IR/lower/backend/oracle checks pass, oracle 25.111s; getter/data control 0.783s.
Mutants: skip selection native/JS, accept wrong scalar kind native/JS, accept wrong literal native/JS, drop nested guards, ignore overlap; all eight caught and restored.
Uncovered: 57 candidate selector pairs/268 candidate reads remain; full unchanged compiler interfaces, non-kind member adapters, nominal targets, getters/static selection and exact allocation reachability are not certified.

Whole-family planning target remains October 12, 2026, 23:00 UTC. Candidate
progress is explicitly selector-projection coverage, not a claim of full compiler
interface conformance. User authorization removed the shared-hook waiting block.

The pinned stock checker at TypeScript 6.0.3/050880ce has zero diagnostics after
locked dependency installation and generated diagnostic declarations. No production
checker diagnostics were bypassed. classify-candidates.cjs refers to a detached
cohere/TypeScript worktree; no code was lifted from cohere. Its input is lazy's
228-pair/2,468-read candidate inventory. Actual shared disjoint literal fields split
it into 156 shared-discriminant pairs/2,019 reads, 61 untagged candidates/434 reads,
and 11 other consumer contracts/15 reads. The 61 retain array/callable overlaps.
A shared enum field with overlapping values is not a discriminant. Open whole
numeric enums preserve number semantics rather than requiring listed enum values.

First ranked group:
- FunctionLikeDeclaration.body: rank 8, 50 candidate reads.
- ForStatement.initializer: rank 11, 43 candidate reads.
- ModuleDeclaration.body: rank 13, 38 candidate reads.
- ForOfStatement.initializer: rank 19, 35 candidate reads.

Each has a separate source fixture set with valid member samples, named wrong-kind
exit 70, a nested wrong label, and permitted absence where declared. Source definitions
are independently written selector projections from measured kind metadata, with
unread unsupported fields retained. They do not copy original interfaces and do not
certify every unread compiler field. Four candidate selector pairs/166 candidate
reads are covered; 57/268 remain unvalidated. The full compiler-pair completion count
is still zero. Both scopes are recorded in lazy-candidate-progress.json.

ViewUnionHasDiscriminant is a new lane-owned IR query. The existing named lower
hook and both backend hooks use it; the shared change is listed in the plan. An own
required scalar kind is checked, including unrestricted numeric enum kind contracts.
Other unread payload fields are deferred. Pure probing now covers initialized own
data on class instances through shared metadata. It does not invoke getters or
resolve inherited static storage. Nominal target contracts remain unsupported by this
matcher, consistently in both backends. The returned id does not become a payload
certificate; subsequent reads keep shared guards. Helper, generic, callback and
stored-field source controls pass with exact holder.value wrong-shape messages.

candidate-refusals.json pins actual native/JS messages. A malformed nested label
fails at value0.payload.label, expected string, found boolean. Invalid kinds fail at
the actual body/initializer read, naming its source alias and found object. Node
controls observe ordinary true values; invalid checked reads intentionally refuse.
class-data-getter.a refuses without printing its getter marker, proving no getter
was invoked by membership probing. Valid class data and wrong class kinds are pinned.

Eight candidate mutations are separate and restored. Skipped selectors, wrong type
acceptance and wrong literal acceptance continue with exit 0; exact refusal pins catch
them. Dropping actual lowered payload/label guards makes native exit by signal and
JavaScript continue with true; both fail the exit-70 pin. Removing disjointness is
caught by TestViewUnionDiscriminantOverlaps. Logs are checked in. No full repository
gate was run; the scoped existing view/lazy checks passed (IR 0.015s, lower 1.371s,
native 4.733s, JavaScript 1.083s). Setup remains 214.165s, nproc 5.

Reproduction:

```
source /workspace/adamic-tools/env.sh
NODE_PATH=/tmp/untagged-stock-api/node_modules node stage3/interface-downcasts/untagged/classify-candidates.cjs /tmp/untagged-typescript stage3/interface-downcasts/untagged/lazy-candidate-progress.json /tmp/untagged-classified-candidates.json > /tmp/untagged-classification.log 2>&1
python3 stage3/interface-downcasts/untagged/generate-candidate-fixtures.py /tmp/untagged-classified-candidates.json 8 11 13 19
python3 stage3/interface-downcasts/untagged/run-candidate-mutants.py > /tmp/untagged-candidate-mutants.log 2>&1
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run 'TestViewUnionDiscriminant|TestUntaggedView|TestCheckedViewUntagged(Selection|SourceDispatch|SourceFlows|CandidatePairs|OwnClassData)' -count=1 -v > /tmp/untagged-group1-shipping.log 2>&1
```

Next: continue most-read remaining own-kind candidates, then field-only/recursive
and array/callable overlaps. Full member adapter coverage is not replaced by these
projection receipts. No individual lane branch was merged.

Earlier checkpoint, historical:

Built: installed three minimal named production hooks; all 12 representative source cases pass without overlays.
Commits: component/source handoff 55b3fdef, evidence 2bb00f0f, integration base ba59427c; production hook commit follows in Git history.
Commands: production oracle 5.751s; scoped lower/JS/native 1.247s/0.975s/4.064s; setup 214.165s, nproc 5.
Mutants: skip native/JS selection, accept native/JS wrong tags, remove nested IR guards; all five caught on production source and restored.
Uncovered: unchanged compiler-pair fixtures and overlapping member adapters; candidate pool remains 228 pairs/2,468 reads pending classification.

Revised whole-family target: October 12, 2026, 23:00 UTC. The user now authorizes
minimal shared hooks directly. No owner handoff is blocking this lane. Only the
three listed shared hunks changed: supportsUntaggedRead in lower/view_contracts.go
and viewUntaggedObjectUnion in each backend's view_unions.go. The integrator can
merge this own-branch tip and reconcile those named hooks hunk by hunk.

The 12-member, five-member and field-only fixtures now compile through production
source loading/lowering and both backends. They include every representative member,
wrong shapes, nested wrong scalars and viewed optional absence. Exact refusal pins
are unchanged; positive results match Node. They are reduced shape fixtures, not
unchanged tsc pair receipts. Candidate completed pairs/reads remain zero, pending
classification and actual per-pair fixtures. Exact runtime reachability is still
unmeasured, but no longer treated as a prerequisite for candidate progress.

Production reproduction:

```
source /workspace/adamic-tools/env.sh
VIEW_UNTAGGED_SOURCE_REQUIRED=1 go test ./internal/oracle -run '^TestCheckedViewUntagged(Selection|SourceDispatch)$' -count=1 -v > /tmp/untagged-wired-source.log 2>&1
go test ./internal/lower ./internal/javascript ./internal/native -run 'TestUntaggedView|TestView|TestLazyView|TestSharedArrayContractAdapter' -count=1 > /tmp/untagged-wired-packages.log 2>&1
python3 stage3/interface-downcasts/untagged/run-source-mutants.py > /tmp/untagged-wired-mutants.log 2>&1
```

All five source mutations were run without GOFLAGS overlays. Individual source
mutation logs and wired summary logs are checked in. Full repository gate was not run.
The source test's no-hook skip is now unreachable on this branch; its required mode
fails if admission is removed. The old frontier test skips the obsolete refusal and
the source-dispatch test supplies executable evidence. No individual lane was merged.

Earlier overlay checkpoint, historical:

Built: merged ba59427c; source selectors and three owner-hook handoffs; 12 source cases pass with an overlay.
Commits: integration base ba59427ccc7afecae29a305c41e6e9c7867e5610; source checkpoint 55b3fdef; evidence checkpoint follows in this branch's Git history.
Commands: overlay source/component oracle 4.509s; scoped lower/JS/native 1.058s/0.858s/3.875s; production frontier 1.436s, source dispatch explicitly skipped.
Mutants: skip native/JS selection, accept native/JS wrong tags, drop lowered nested guards; all five caught by exact source pins and restored.
Uncovered: production hooks and unchanged tsc pairs; exact family pairs/reads remaining are unmeasured, not zero.

Revised whole-family estimate: October 14, 2026 UTC, conditional on the integrator
installing the three hooks and a family-specific reachability inventory. This is a
planning estimate. The current census cannot support an unconditional delivery date.

The designated integration merge was a clean fast-forward. It brought lazy admission
and all pushed lane tips; no individual lane branch was merged. Production shared
files remain untouched, as required by territory ownership. source-hooks.patch contains
three concrete hunks for lower/view_contracts.go, native/view_unions.go and
javascript/view_unions.go. The integrator must apply them, format the shared files and
replace the old source-refusal frontier with the source-dispatch evidence.

The new helpers use the interned registry and shared slot initialization/type metadata.
A member's own required finite tags select it lazily; payload reads keep shared guards.
Field-only members require complete acyclic scalar/plain-object contracts. Other
families, classes and accessor-based membership remain refused. Required shared tags
keep the existing dispatch. JavaScript honors allowed undefined before object selection.
The selected id is not stored as dynamic provenance: subsequent reads rely on the
shared conservative checked-read machinery. This checkpoint does not certify every
helper/generic/callback/field flow for these unions.

Actual .a source programs exercise every member of the reduced 12-member name union,
every member of the five-member option union, and both structural alternatives. Each
shape also tests wrong membership, a nested wrong scalar, and viewed optional absence.
Node controls provide the observed ordinary behavior; compiled invalid views instead
stop with exact exit 70/stderr pins in source-refusals.json. These are representative
shape fixtures, not unchanged full compiler interface fixtures or completed census pairs.
The earlier component tests still include sanitized native execution.

Measurement uses lazy/census/read-demand-pairs.json.gz and lazy/ADAPTED-CENSUS.md.
The latter explicitly states at lines 64-65 that allocation reachability is unmeasured.
lazy-candidate-progress.json ranks all 228 static object-union pairs/2,468 candidate
reads, including tagged unions. All remain pending as census obligations; zero
production pairs/reads are claimed complete. Exact untagged remaining counts are null.
The old pair-progress.json 84/186 queue is historical and must not be used as the
current exact denominator. The fresh pool's leading BindingName reads (112, 96, 57)
have shared discriminants, so they cannot simply be charged to this lane. The reduced
12-member name fixture is not evidence that these tagged pairs are completed here.

Source mutation evidence:
- source-skip-native and source-skip-javascript: binding-name/wrong continues with exit 0; exact refusal pin fails.
- source-wrong-shape-native and source-wrong-shape-javascript: invalid kind 99 is admitted; the same exact refusal pin fails.
- source-drop-transitive: remove payload/label guards in the actual lowered program; native exits by signal (-1), JavaScript prints true with exit 0. Both fail the expected exit 70 pin. No sanitizer failure is being treated as the check.

Reproduction (all test output goes to files):

```
source /workspace/adamic-tools/env.sh
python3 stage3/interface-downcasts/untagged/source-overlay.py /tmp/untagged-source-overlay
VIEW_UNTAGGED_SOURCE_REQUIRED=1 GOFLAGS=-overlay=/tmp/untagged-source-overlay/overlay.json go test ./internal/oracle -run '^TestCheckedViewUntagged(Selection|SourceDispatch)$' -count=1 -v > /tmp/untagged-source-final.log 2>&1
GOFLAGS=-overlay=/tmp/untagged-source-overlay/overlay.json go test ./internal/lower ./internal/javascript ./internal/native -run 'TestUntaggedView|TestView|TestLazyView|TestSharedArrayContractAdapter' -count=1 > /tmp/untagged-source-packages.log 2>&1
GOFLAGS=-overlay=/tmp/untagged-source-overlay/overlay.json python3 stage3/interface-downcasts/untagged/run-source-mutants.py > /tmp/untagged-source-mutants.log 2>&1
go test ./internal/lower ./internal/javascript ./internal/native ./internal/oracle -run 'TestUntaggedView|TestCheckedViewUntaggedSourceFrontier|TestCheckedViewUntaggedSourceDispatch' -count=1 -v > /tmp/untagged-production-frontier.log 2>&1
```

Setup: GOPROXY=https://proxy.golang.org|direct; submodules ready 9.847s,
go build ready 213.987s, deferred test binaries 214.121s, build cache warm
214.124s, done 214.165s; nproc=5, cgroup quota=4 CPUs. Environment source is
/workspace/adamic-tools/env.sh. Logs are copied into this lane's logs directory.
Full repository gate was not run. Whole-tsc checker diagnostics do not prevent these
source fixtures, but prevent the claimed exact whole-program reachability census.

Pending owner action is concrete and reviewable: apply source-hooks.patch on the
integration branch and wake this lane with that tip. Until then, production source
admission remains blocked by the shared-hook ownership rule, not lazy cast admission.

Earlier checkpoint, retained as historical evidence:

Built: member-specific tag filtering, selected-contract preservation and structural fallback components; no source admission.
Commits: territory 6c321f8a; lane 4 refresh d15b4206; selector 566aad67; mutant checkpoint recorded in Git history.
Commands: focused lower/oracle passed (0.006s/3.470s); counted absent controls passed (1.299s); touched-package vet passed.
Mutants: skip check, accept wrong shape, drop nested check, separately in native and JavaScript; all six caught by semantic output/exit pins.
Uncovered: all 84 pairs and 186 reads remain; full compiler interfaces, lazy read dispatch and transitive propagation are pending.

Working date for the whole family: October 12, 2026 UTC, conditional on the lazy
read hooks and normalized probes arriving by October 9. This is an estimate,
not a promise that unsupported callable, array or branded descendants disappear.
The shared-file merge conflicts make an unconditional completion date indefensible.

The deterministic queue in pair-progress.json retains all overlapping families,
receiver identities, declared types and source witnesses. Its leading pairs are:

| Rank | Receiver / field | Reads |
| --- | --- | ---: |
| 1 | NamedDeclaration & { name: DeclarationName }.name | 25 |
| 2 | CommandLineOptionOfListType.element | 22 |
| 3 | ElementWithComputedPropertyName.name | 12 |
| 4 | TypeNode & LiteralTypeNode & { literal: StringLiteral }.literal | 11 |
| 5 | ParameterPropertyDeclaration.name | 8 |

The 12-member declaration-name shape has another one-read pair: 26 reads total.
The inventory also labels narrowed receivers, array alternatives and callable
alternatives as this family. None is silently removed from the 84-pair queue.

UntaggedViewMembers consumes the existing IR registry and preserves member ids.
Each candidate's required finite scalar fields can filter it independently:
there is no requirement that every alternative names the same discriminant.
Optional tags cannot exclude legal absent values. A tag never certifies payloads.
The C and JavaScript selectors try later alternatives after failure, require a
nonzero contract and a structural matcher, and return the selected contract id.
Shared dispatch must preserve that id on every subsequent checked read.

The component oracle uses reduced 12-member and 5-member shapes, ordered by the
most-read demand, plus two interfaces distinguished by fields alone. These are
representative member-count controls, not copies of tsc's full interfaces.
Source Node establishes their ordinary values and exposes malformed payloads.
Normalized C/JavaScript component snapshots are checked against those outputs;
negative cases pin the field, expected union, found object, and exit 70.
Native positives run release and ASan/UBSan/LeakSanitizer. Panics disable leak
checking because exiting at a panic intentionally does not unwind live data.

The supplied C matcher reads normalized sample data, not production object slots.
The supplied JavaScript matcher uses own data descriptors, without getters.
Shared readiness and complete structural membership remain owner-supplied
adapters. The component does not implement a second readiness bitmap or flow
solver. Its nested-check mutant is an adapter mutation, not proof that every
compiler-generated nested field read is checked. Actual helpers, generics,
callbacks and stored-value propagation remain unverified for this family.

TestCheckedViewUntaggedSourceFrontier records the honest integration frontier:
source Node prints true for each viewed-source probe; common lowering refuses
at the cast with `object union checked view without a common finite discriminant`.
Both backends share that lowering refusal. Allowed absent controls compile as
ordinary source and print true on source Node, native, and emitted JavaScript;
this is not evidence for absence through a viewed union. Count-row additions
are supplied for the shared owner rather than modifying counts.md.

Required concrete hooks and the lazy merge's six conflicting files are recorded
in docs/checked-views-plan.md. Lazy admission 5002bfe0 was fetched and the merge
aborted. Lane 4 9ecdda53 was merged, resolving only the plan conflict by preserving
both additions. No other lane's compiler file was edited by this unit.

Setup: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh,
source /workspace/adamic-tools/env.sh. Go ready 0.231s, Node ready 0.292s,
clang ready 0.742s, markdown ready 1.378s, submodules ready 27.998s,
Go build ready 66.995s, cache warm 67.085s, done 67.116s; nproc 5,
cgroup quota 4 CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Commands wrote complete output to logs, never through a pipe:

```sh
python3 stage3/interface-downcasts/untagged/rank-pairs.py
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestUntaggedView|TestCheckedViewUntagged' -count=1 -v -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewUntaggedSourceFrontier$' -count=1 -v -timeout 10m
go vet ./internal/lower ./internal/javascript ./internal/native ./internal/oracle
```

The full repository gate was not run. The full touched-package run is recorded
separately once complete; the inherited TestSharedArrayContractAdapter failure
on readonly (number | string)[] is still present and was already documented
on the baseline lane. No source pair completion or complete-family gate is claimed.

The user's later coordination ruling supersedes direct lane merging. Only
codex/views-integration is now consumed for other lanes' code. At the final
remote check the integration branch had not yet been published (ls-remote
returned no ref). Further direct lane merges were stopped. This unit is blocked
on that shared integration/lazy read dispatch and can rest after pushing its
component checkpoint. The user will wake it when lazy admission lands.


Final touched-package result: native passed all tests in 123.992s; JavaScript
passed all tests in 1.093s; lower failed in 16.774s only at the inherited
TestSharedArrayContractAdapter/readonly_(number_|_string)[] disagreement.
The new lower tests passed in the focused run. The package gate therefore exits
1 and is not reported as green. Focused oracle validation and touched-package
vet are green. The full repository gate and full oracle package were not run.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/javascript ./internal/native -count=1 -timeout 30m
python3 stage3/interface-downcasts/untagged/run-mutants.py
```

Each mutation is made independently and restored in a finally block. Logs
include the semantic mismatch and a successfully executed release binary;
none is killed by clang or a sanitizer. The mutants and their catchers are:

| Mutant | Catcher |
| --- | --- |
| skip-check-native | nested binding-name input exits 0 instead of pinned 70 |
| skip-check-javascript | nested binding-name input exits 0 instead of pinned 70 |
| accept-wrong-shape-native | wrong binding-name kind accepted, exits 0 |
| accept-wrong-shape-javascript | wrong binding-name kind accepted, exits 0 |
| drop-transitive-native | component nested boolean label accepted, exits 0 |
| drop-transitive-javascript | component nested boolean label accepted, exits 0 |

The nested mutations modify the component adapter in the oracle harness, whose
complete matcher is a required future shared hook. They do not claim a mutant
of production compiler-wide transitive propagation. Selector mutations modify
this unit's runtime files. Source-frontier pins should be replaced with true
source admission tests once the shared integration implements the hooks.

After every push: 84 pairs and 186 reads remain. No pair is reduced based on
these components. All files and mutation sources are restored. Branch publication
uses only codex/views-untagged-object-unions and opens no pull request.
