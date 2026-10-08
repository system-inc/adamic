# Root provenance evidence

Measured compiler sources remain identical to base `44583d3283fdd8674085a7ddcce040cf2a73a94e`. Started this follow-up at report tip `dc6b1529ae9d2a2210672e105c8bb6374619a59d`. Only the census overlay and report tooling changed. The adapted 81-file tsc source bytes, original checker diagnostics and attempt ledger are unchanged. Toolchain setup and its timing lines are preserved in [the original evidence](../EVIDENCE.md); reused that environment (`nproc=5`, setup 187.482s).

## Implementation and counting

The overlay instruments declaration names in `variables` and `declareLocal`, preserving production code on disk. Active statement frames retain resolved checker symbols and declaration nodes within that same source-file span. Before rollback, `latentBlockDeclarations` links absent checkpoint locals to the actual failed finding. `latentReadNotYet` attaches `blocked_by` and `blocked_symbol_declaration` only for the same resolved symbol; shorthand uses its actual value symbol. Frames and provenance reset for each independent unit. Transitive failures retain the original root.

An echo must be tagged in every attempt to leave the root-site count. Mixed or untagged reads remain conservative root candidates. Referenced Refused and SkippedDependency roots are listed separately; the latter are measurement boundaries, not compiler lessons. Counts use exact `(kind, where, reason, text)` sites. The complete root-to-echo CSV includes declaration and attempt-unit provenance, including echoes rooted outside their diagnostic file.

The newly annotated stream has 14,705 raw NotYet observations versus the original 14,704, because distinct provenance retains one additional observation formerly deduplicated. Its underlying observation set, every source/checker input and every unit/declaration ledger are identical. Unique NotYet sites remain 10,426.

## Full census

All commands ran from the repository root with output redirected to logs.

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/stage3-notyet-roots-overlay > /tmp/stage3-notyet-roots-build.log 2>&1
python3 stage3/meter/entry_overlay.py /tmp/stage3-notyet-roots-overlay /tmp/stage3-notyet-roots-entry-overlay >> /tmp/stage3-notyet-roots-build.log 2>&1
go build -buildvcs=false -overlay=/tmp/stage3-notyet-roots-entry-overlay/overlay.json -o /tmp/stage3-notyet-roots-census ./stage3/census/latent/tool >> /tmp/stage3-notyet-roots-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/stage3-notyet-roots-census /tmp/stage3-notyet-adapted/src/tsc/tsc.ts /tmp/stage3-notyet-roots-full.jsonl > /tmp/stage3-notyet-roots-full.log 2>&1
python3 stage3/notyet-table/regroup-roots.py /tmp/stage3-notyet-roots-full.jsonl > /tmp/stage3-notyet-roots-regroup.log 2>&1
python3 stage3/notyet-table/render-roots.py > /tmp/stage3-notyet-roots-render.log 2>&1
python3 stage3/notyet-table/audit-roots.py > /tmp/stage3-notyet-roots-accounting.log 2>&1
```

Exit 0. The full stream has one header and 81 file records. [Build](evidence/build.log.txt), [full run](evidence/full.log.txt), [regrouping](evidence/regroup.log.txt), [root accounting](evidence/accounting.log.txt). [Unfiltered annotated stream](full.jsonl.gz), [annotated raw CSV](raw.csv), [root-to-echo ledger](echoes.csv), [summary](summary.json).

## Poison and mutants

```sh
python3 stage3/census/latent/provenance_audit.py /tmp/stage3-notyet-roots-census stage3/notyet-table/roots/evidence/poison > /tmp/stage3-notyet-roots-poison.log 2>&1
```

Exit 0. [Poison audit](evidence/poison-audit.log.txt) prints:

```text
drop-tagging mutant caught by poison blocked_by assertion: poison echo must carry blocked_by pointing at its actual root
PASS: three poison reads including shorthand carry the actual failed initializer root; secondary flattens to that root; shadow and independent parameter symbols are not poisoned; no-output guards
```

The actual root is NotYet `a function without a body` at probe.a:1:1 in the rollback unit. The read at 4:31 carries that exact root and `blocked_symbol_declaration` at 3:11. [Positive findings](evidence/poison/positive.jsonl), [drop-tag mutant findings](evidence/poison/drop-tag-mutant.jsonl), [checker-clean witness](evidence/poison/probe.a). The mutant leaves root and read findings intact and fails specifically because `blocked_by` is missing. Shorthand, transitive secondary binding, shadowing and an unrelated parameter are checked by the same audit.

The independent real-data root accounting audit catches seven accounting mutants: root total, removed echo count, mixed-site count, per-reason root count, echo attribution, coverage and non-NotYet root count. Three rendered-table mutants change the headline, a root row and an echo count; all fail their respective assertions. It verifies every root against an actual finding in its same attempt record and compares the annotated census to the original observation set.

Additional commands:

```sh
go build -buildvcs=false -overlay=/tmp/stage3-notyet-roots-overlay/overlay.json -o /tmp/stage3-notyet-roots-unit-census ./stage3/census/latent/tool > /tmp/stage3-notyet-roots-unit-build.log 2>&1
python3 stage3/notyet-table/validate-tooling.py full /tmp/stage3-notyet-roots-unit-census stage3/notyet-table/roots/evidence/full-audit > /tmp/stage3-notyet-roots-full-audit.log 2>&1
python3 stage3/notyet-table/validate-tooling.py replay > /tmp/stage3-notyet-roots-replay-tests.log 2>&1
python3 stage3/census/latent/audit_output_guards.py "$PWD" /tmp/stage3-notyet-roots-overlay /tmp/stage3-notyet-roots-output-mutants > /tmp/stage3-notyet-roots-output-mutants.log 2>&1
go vet -overlay=/tmp/stage3-notyet-roots-overlay/overlay.json ./stage3/census/latent/tool ./stage3/census/latent/replay/worker > /tmp/stage3-notyet-roots-vet.log 2>&1
```

All exit 0. [Continuation/rollback audit](evidence/full-audit.log.txt) catches first-error-only and retained-failed-state mutants. [Replay integration](evidence/replay-tests.log.txt) prints `Ran 2 tests in 17.147s`, `OK`, including the sibling-selection negative control. [Output-guard audit](evidence/output-mutants.log.txt) catches non-nil IR and permissive production loader mutants. [Vet](evidence/vet.log.txt) is empty. The directory-based full audit uses the directory driver, while the tsc run and poison entry use the entry-reach driver.

## Real-project replay roots

```sh
go run ./stage3/census/latent/replay -project /tmp/stage3-notyet-adapted/src/tsc/tsc.ts -where /tmp/stage3-notyet-adapted/src/compiler/factory/nodeFactory.ts:1237:9 -kind NotYet -reason 'reading node' > stage3/notyet-table/roots/evidence/nodeFactory.json 2> stage3/notyet-table/roots/evidence/nodeFactory.log.txt
go run ./stage3/census/latent/replay -project /tmp/stage3-notyet-adapted/src/tsc/tsc.ts -where /tmp/stage3-notyet-adapted/src/compiler/factory/nodeFactory.ts:1246:9 -kind NotYet -reason 'reading node' > stage3/notyet-table/roots/evidence/nodeFactoryBigInt.json 2> stage3/notyet-table/roots/evidence/nodeFactoryBigInt.log.txt
```

Both exit 0 and print `reproduced NotYet: reading node` at the exact requested positions. [Numeric literal JSON](evidence/nodeFactory.json), [log](evidence/nodeFactory.log.txt), [command](evidence/nodeFactory-command.json); [big-int literal JSON](evidence/nodeFactoryBigInt.json), [log](evidence/nodeFactoryBigInt.log.txt), [command](evidence/nodeFactoryBigInt-command.json). The first root is NotYet `a value of type T["kind"]` at nodeFactory.ts:1213:59; the second has the same reason at 1436:46. The read and full root signature, including attempt unit and failed declaration identity, match the full census in both independent units. Stage 0 needs the checked generic factory representation; the later node reads are echoes.

An additional distinct-file command is preserved: [checker command](evidence/checker-command.json), [JSON](evidence/checker.json), [log](evidence/checker.log.txt). `reading node` at checker.ts:2043:22 reproduces with the same failed declaration at 2041:15. Standalone replay reaches Refused `overload 2 of getParseTreeNode result T | undefined cannot be served by implementation result Node | undefined` at utilitiesPublic.ts:826:1. The full census instead reaches NotYet `a value of type ((node: Node) => boolean) | undefined` at 827:58. Both are actual initializer failures, but their root signatures differ with isolated replay context. The table uses the full census's observed root and does not claim exact root reproduction for this extra example.

## Limits

The earlier plain replay sampling was superseded by the user's provenance request and does not decide these counts. Untagged binding failures, namespace registration, separately attempted bodies and skipped checker dependencies are not guessed into a rollback root. No production compiler source, oracle fixture registration or counts.md changed. No whole compiler package or full gate ran; the prior unchanged-base counts failures remain documented in the original evidence. This report does not establish whole-program lowering or backend semantics.
