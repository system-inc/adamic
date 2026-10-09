# `nexus/correctness-no-discarded-outcome`

| | |
|---|---|
| **Recommendation** | **Yes, at `error` once ahra's 4 sites are fixed.** 4 findings on ahra, 0 false |
| Findings | **ahra 4**, all `writeJsonFile` state writes (measured 2026-10-03) |
| Measured precision | 4 of 4 true: each drops a `JsonFileWriteOutcomeType` whose `Unwritable` arm is never read |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to read the call's type and trace each arm to its declaration; without a checker the rule registers nothing |

## What it checks

A call made as a statement on its own (or `await` of a call) whose value is one of Nexus's `{ outcome }`
types, so the outcome, failure arm included, is thrown away. The types are decided by declaration: the
alias name and the Nexus file that declares it, never a name pattern or the presence of an `outcome` field.
The five, which are every outcome union in Nexus on 2026-10-03:

| Type | Declared in | Returned by |
|---|---|---|
| `JsonParseOutcomeType` | `nexus/source/structured-text/json/Json.ts` | `parseJson` |
| `JsoncParseOutcomeType` | `nexus/source/structured-text/json/Jsonc.ts` | `parseJsonc` |
| `JsonFileReadOutcomeType` | `nexus/source/structured-text/json/JsonFile.ts` | `readJsonFile` |
| `JsonFileWriteOutcomeType` | `nexus/source/structured-text/json/JsonFile.ts` | `writeJsonFile` |
| `PackageJsonReadOutcomeType` | `nexus/source/system/PackageJson.ts` | `readPackageJson` |

A project wrapper returning one of these is reported where its result is dropped, written or inferred
return type alike. The test is on the arms, so an optional call's `JsonFileWriteOutcomeType | undefined`
and a project union adding an arm beside Nexus's still count; a value narrowed to some arms
(`Extract<..., { outcome: 'Written' }>`) does not. `void f()` is the opt-out. The doc comment carries the
declined shapes, each owned by another rule or a missed finding by design.

## Where it came from

Four state writes in ahra, each a bare `writeJsonFile(...)` in a function returning `void`:
`modules/meta/MetaPollingApi.ts:41`, `modules/google/youtube/YouTubeSafety.ts:57`,
`modules/pensieve/PensieveWindows.ts:43`, `modules/pensieve/PensieveWeeklies.ts:99`. Found by the
cross-language pass of the new-rules sweep (`#tevhg3f`, after rustc's `unused_must_use` and go vet's
`unusedresult`), built as `#1vf0qwd`.

## Reconciling with the research count

Research counted **15** on ahra, keyed on any discarded value whose type carries an `outcome` field.
cohere counts **4**, the four `writeJsonFile` sites, which the research named. A measurement-only probe
keyed the way research was (any constituent with an `outcome` member) reproduces research's 15 exactly
on today's tree, so the gap is the scope and nothing else: the other 11 discard outcome types ahra declares
itself, which this rule does not recognize by design.

| Site | Type, declared in ahra |
|---|---|
| `modules/os/wisdom/AhraOsWisdom.ts:4210` (`withdrawVerdict`) | `WithdrawVerdictOutcomeInterface` |
| `modules/os/wisdom/AhraOsWisdomApi.ts:1054` (`flagTaskAsIntent`) | `FlagTaskIntentOutcomeType` |
| `modules/os/lifecycle/AhraOsWake.ts:826` (`await wakePosition`) | `WakeOutcomeInterface` |
| `modules/os/minds/runtime/MindRuntimeSupervisor.ts:980` | `DeliverAndWakeResultInterface` |
| `app/(os-layout)/os/wisdom/_components/` x4 (`WisdomAcceptStrikeControls.tsx:41`, `:59`, `WisdomDecisionCard.tsx:68`, `WisdomFailedControls.tsx:38`) | `WisdomCompletionResponseInterface` |
| `app/(os-layout)/_components/detail/TaskDetailActions.tsx:139`, `:152`, `os/wisdom/_components/WisdomHome.tsx:297` | anonymous unions with `outcome` arms |

These are likely real too; a project naming its own outcome types to this rule is the follow-up, not a
widening of the type test.

## Verification

- Fixtures from the four real sites, before (one finding each, on the bare write) and after (the outcome
  kept and its `Unwritable` arm handled, silent). 12 more firing shapes, one per Nexus type plus the
  wrapper, async, method, optional-call, extended-union, loop and module-level forms; 10 silent shapes;
  and a lookalike `JsonFileWriteOutcomeType` declared outside Nexus that must stay silent beside a Nexus
  write that fires.
- Mutation check, each mutant a copy through `go test -overlay -count=1`, each killed: accepting any
  alias (same name and shape declared here; the lookalike), dropping the declaring-file test (the same
  two), accepting a value carrying any arm instead of every arm (narrowed to the success arm), not
  awaiting the `await` operand (async wrapper), not splitting a union (all four real sites, the async
  wrapper, the lookalike). The top-level-alias guard has no fixture reaching it; it stays because a
  nested alias of the same name in a Nexus file is not one this list has read.
- `gofmt -l` and `go vet` clean through the overlay.
