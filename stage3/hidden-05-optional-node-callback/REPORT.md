Archived hidden-05 replay and safety evidence for roadmap step 30; no lowering behavior changed.
Base: dcdbb9098f77f30ad41790c56df1bd63ad462b63; evidence delivery branch: codex/hidden-05-optional-node-callback.
Checks: 82 pinned hashes match; exact new-boundary replay succeeds; focused overload relation tests pass before and after the mutant.
Mutant: removing the overload parameter guard fails parameter contravariance and mutable-parameter checks; source restored exactly.
Uncovered: historical callback stop still does not reproduce; optional calls, callback-drop mutant, both backend fixtures and counts update remain blocked.

The recorded caller es2018.ts:827:5 now reaches a Refused at visitorPublic.ts:123:5: overload 1 of visitNode parameter node cannot be served by implementation parameter node. The overload admits TIn extends Node | undefined, while the implementation declares node: Node. Removing this contract check also accepts a concrete reduction dereferencing undefined: Node exits 1 with TypeError, while Adamic correctly refuses it. The implementation's early undefined return in the real source cannot by itself prove the coupled visitor argument and result contracts for every generic overload. Those contracts need a body-level proof; no blanket relaxation is delivered.

The documented diagnostic-position replay selects visitorPublic.ts:148:1 and fails its exact match. A separate scratch-only overlay selected the recorded caller at es2018.ts:827:5 while retaining the original diagnostic-position, kind and reason matcher. Its historical-pin run also mismatches, encountering the overload refusal first. The same caller selection on area-next exactly reproduces the new refusal. Neither run is claimed as reproducing the historical optional-callback diagnostic. Both workers preserve the no-output guards. Scratch overlays and isolated worktrees were not merged into the delivery branch.

The independent witness prints true|false on Node with empty stderr and exit 0. Adamic on this base stops at an optional call at the witness's line 2. This witness does not reproduce the census head and was not added as a supported fixture.

The file-scoped hidden census loads the whole hash-pinned compiler project and retains its whole-project registration, but runs every attempted unit only for transformers/es2018.ts. Its 63 units and 483 findings produce an intersection of 6,078 hidden bytes with [35690,41768), matching the old intersection of 6,078: difference 0 bytes. This does not claim a whole-corpus remeasurement or newly revealed bytes. The smallest current head boundary is [35690,35769), with the overload refusal above. An enclosing transformFunctionBody result-overload refusal remains [3955,69978).

Commands wrote output to logs. Exact selectors and logs:
- go run ./stage3/census/latent/replay -project /tmp/hidden-adapted/src/compiler -where /tmp/hidden-adapted/src/compiler/visitorPublic.ts:151:5 -kind NotYet -reason 'a value of type ((node: Node) => boolean) | undefined': exit 1, evidence/diagnostic-selector-replay.log.txt.
- Historical scratch caller worker with the same selector: exit 1, evidence/pin-caller-replay.log.txt.
- Area-next scratch caller worker with -where /tmp/hidden-adapted/src/compiler/visitorPublic.ts:123:5 -kind Refused -reason 'overload 1 of visitNode parameter node cannot be served by implementation parameter node': exit 0, evidence/new-boundary-replay.log.txt.
- LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/hidden-05-region-census /tmp/hidden-adapted/src/compiler /tmp/hidden-05-region.jsonl: exit 0, evidence/region-run.log.txt.
- go test ./internal/lower -run '^TestCensusOverloadRelation$' -count=1 -timeout 30m: exit 0 before and after restoration, evidence/overload-guard.log.txt and evidence/overload-restored.log.txt.
- Independent guard mutant: exit 1 at two intended assertions, evidence/overload-mutant.log.txt.
- Node semantic observation: evidence/node.log.txt and evidence/node.stderr.log.txt.
- Unsafe overload reduction: the unsafe reduction below; Node exit 1, evidence/unsafe-node.log.txt; compiler refusal, evidence/unsafe-native.log.txt.

No full package or repository gate ran. Two broad census attempts were stopped: the first overlapped input preparation, and the second was replaced by the file-scoped measurement. Their partial output is not used as completed census evidence. No fixture was registered and no counts row was changed, so no counts update is claimed. The requested drop-present-callback mutant could not be run through either backend while this stop remains; the overload mutant is separate evidence, not a substitute.

Setup used GOPROXY=https://proxy.golang.org|direct, cloud/setup.sh, then /workspace/adamic-tools/env.sh. It ran on the previous feature tip before the requested branch checkout was corrected; subsequent commands rebuilt against area-next. Setup lines: Go 0.028s, Node 0.028s, submodules 0.075s, markdown step-duration 0.011s and ready 0.086s, clang ready 0.173s, Go build ready 40.703s, test binaries deferred 40.963s, build cache warm 40.965s, done 40.993s. nproc 5, quota 4; Go 1.27.1, Node 24.19.0, clang 20.1.8.

No main or area branch was modified, merged into or pushed. This delivery archives evidence only. The parameter and result overload family, including visitNode, belongs to the overload-results unit, which builds per-overload specialization and checked boundaries.

The two safety catchers are `TestCensusOverloadRelation/parameter_contravariance` and `TestCensusOverloadRelation/mutable_parameters`. The temporary mutation changed `if !l.censusRelated(given, takes)` to `if !l.censusRelated(given, takes) && false`; no compiler mutation is included. Baseline and restored test logs surround the independent mutant log.

The caller replay used a scratch-only replacement of `latentReplaySelect(program, where)` with `latentReplaySelect(program, "/tmp/hidden-adapted/src/compiler/transformers/es2018.ts:827:5")`. Diagnostic position, kind and reason matching stayed exact. File-scoped measurement added only a file filter in `latentFullSelected`, before creating each per-file lowering state; all 63 target units and whole-project registration remained unchanged. Raw ledger: `evidence/region-census.jsonl.gz`.

Source of the independent optional-call witness (unregistered, still NotYet):

```typescript
interface Node { readonly value: number; }
function test(node: Node, predicate: ((node: Node) => boolean) | undefined): boolean { return predicate?.(node) ?? false; }
console.log(`${test({ value: 7 }, node => node.value === 7)}|${test({ value: 7 }, undefined)}`);
```

Source of the unsafe overload reduction (intentionally refused, unregistered):

```typescript
interface Node { readonly value: number; }
function visitNode<TIn extends Node | undefined>(node: TIn): Node | undefined;
function visitNode(node: Node): Node | undefined { return { value: node.value }; }
console.log(`${visitNode(undefined)?.value ?? 0}`);
```
