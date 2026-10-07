Withdrawn duplicate: slot 05 owns NewTheme under earlier claim b9d90803 (02:30:29 UTC), before b07ef13e (02:30:45 UTC). The passing observations below are historical evidence, not an additional delivered helper. Active constructor sources and tests were removed.

Built collapse.NewTheme in one .a file, removing six dependency entries across six Tailwind rules, zero final blockers alone.
Commits: b07ef13e claim pushed before implementation; implementation commit follows this report.
Commands: helper parity PASS 9.12s with 13,272 identical bytes; vet exit zero; filtered external Node oracle passed; 157 sources captured from all six consumers.
Mutants: dirty prefix, nonnil empty order, shared values map and shared record all compile and execute cleanly, then fail Go byte comparison on source Node, emitted JavaScript and sanitized native.
Not covered: complete consuming-rule integration, full repository gate, external Tailwind installations/corpora; separate consumer-package gate failed eight placement/live checks.

## Contract and observations

Original Go cohere constructor is NewTheme in collapse/theme.go, pinned commit 715ba94f3608a6500086b1076ce5cb7e51b836db. Every call initializes a fresh values map and fresh record, zero Prefix and deadKeys, and nil keyOrder. The explicit nil bit in ThemeOrder preserves nil-versus-empty state. This base's compiler refused string[] | null, so the representation was adapted inside the owned file without a compiler edit. No later Theme operation is implemented or claimed here.

The zero-argument constructor is run for every one of 157 actual fixture sources from all six consumers, plus one control. Source-derived keys mutate one instance's map, prefix, order and counter while later instances are observed; no other Go helper is approximated. The original Go constructor and private fields are exposed via a temporary owned overlay. Six lines per case yield 13,272 identical bytes on Go, original .a source Node, emitted JavaScript and ASan/UBSan native. All twelve mutant/backend combinations exit zero without stderr before their changed bytes fail comparison. The helper removes a prerequisite from each rule listed below; this is not a claim of complete findings/fixes parity or final helper readiness.

## Rules

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

## Reproduction

```bash
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/from-wave1-06/testdata/regenerate.py > /tmp/helpers06-regenerate.log 2>&1
go test ./stage1/cohere/lint/helpers/from-wave1-06 -run '^TestNewTheme$' -count=1 -v -timeout=20m > /tmp/helpers06-new-theme-view.log 2>&1
go vet ./stage1/cohere/lint/helpers/from-wave1-06 > /tmp/helpers06-vet.log 2>&1
```

Capture is adapted from slot03's published procedure, using runtime hooks before external-engine skips and source-order-independent deduplication. All 157 inputs and the coverage metadata are retained. The unmodified consumer-package gate exits 1 in 29.116s: TestUnknownClassFixturesActuallyRan, TestCanonicalFixturesActuallyRan, TestConflictFixturesActuallyRan and TestClassOrderFixturesActuallyRan cannot find installed tailwindcss under /Users/kirkouimet/Projects/ahra/app/_theme/styles; TestConflictingClassesPlacementIsAccountedFor, TestUnknownClassesPlacesEveryCorpusClass, TestCanonicalClassesPlacementIsAccountedFor and TestClassOrderLiveMatchesTheEngineOverTheCorpus also fail for unavailable external corpus/engine placement. This gate is not counted as passing. Captured sources establish input coverage, not instrumentation of every actual constructor call in the whole rules.

## Environment and superseded attempts

New worktree /workspace/adamic-helpers06, branch codex/lint-helpers-from-codex/lint-wave1-06, based on origin/codex/lint-helpers 95100eb4. bash cloud/setup.sh passed: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 20s, cache warm 490s, done 490s. nproc=5; Go 1.27.1, clang 20.1.8, Node 24.19.0.

An early test raced ahead of capture and failed for missing sources.jsonl.gz. A subsequent metadata adapter omitted the cohere_commit JSON tag and incorrectly reported pin drift; that tag was fixed. The owned main driver was adapted to this base's tagged readTextFile result and string-only console API. Nullable arrays were refused explicitly, then represented by a nil-bit slice view. None of these superseded failures is credited as a mutant or passing run. Only owned helper files, owned test adapters and the claim are authored; no shared generator, shared harness, cohere source or compiler file was changed.
