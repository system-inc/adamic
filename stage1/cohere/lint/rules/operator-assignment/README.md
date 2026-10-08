# operator-assignment

Ports the pinned Go cohere core rule without editing shared registration or harness files. `rule.json` discovers the listener and the Go oracle adapter. The adapter returns the unmodified upstream rule and preserves its typed default and positional string options.

The default requires compound assignment; `never` expands it. Reference comparison, precedence, source slices, comment checks, ranges, and fix declines follow Go. Shared `RuleContext` operations, `comments.ForFile`, and `FileCommentCache` are reused. Each listener owns one cache for its immutable parsed file. Comments have byte ranges, so the rule compares byte offsets without slicing at those offsets.

Go-specific cases retained include `obj.a = obj?.a + b` (same reference, finding without a fix), `x = y * x` (finding without a fix), and dynamic computed keys (finding without a fix). Dot and bracket spellings remain distinct. Logical assignments are excluded.

Validation captures all 126 upstream source/rule/options combinations: 69 reporting cases, 50 clean cases, and seven TypeScript cases. `validate.py` selects this rule's upstream tests and witnesses using temporary Go overlays; its selected capture guard requires exactly 126 cases. It never edits shared files. The normal whole-package test uses no overlay.

With the setup environment and inputs in `evidence/inputs.sh` sourced, run:

```sh
go run ./cmd/lint-registry
python3 stage1/cohere/lint/rules/operator-assignment/validate.py
go test ./stage1/cohere/lint -count=1 -v -json -timeout=90m
```

`evidence/selected.log` records byte-identical Go, source Node, emitted JavaScript, and sanitized native outputs on the captured upstream cases and inherited corner corpus, plus both witnesses. The `operator-assignment-reversed-fix` mutant compiles and executes, then disagrees with Go on all three runtimes. `testdata/never.options.json` enables the expansion witness.

`evidence/whole-summary.json` records the whole-package result, counts including subtests, skip and failure names, wall time, nproc, and load. `evidence/whole.log` is its complete readable output. The initial selected run's all-rule minimum-count guard rejected the correct 126-case subset; `evidence/selected-initial.log` preserves that failure, and the final selected run uses the exact rule count.

The initial whole-package attempt exhausted its 30-minute limit during registered mutant builds, with no semantic failure reported. `evidence/whole-initial.log` and `whole-initial-summary.json` preserve that incomplete attempt. The complete run uses a 90-minute limit.

The second whole-package attempt reached the sanitized corpus command limit. `evidence/whole-before-cache-fix.log` preserves it. The final rule uses the existing per-file comment cache instead of rescanning the source for every proposed fix; final validation runs after that change.

The final whole-package run passed: 114 tests including subtests, zero failures, and one inherited skip, `TestCheckerBridgeRefusalPending` (the checker error bridge is pending). Wall time was 2979.335 seconds on nproc 5; initial load was 0.01/0.97/1.72 and final load was 2.31/2.10/2.59. All requested input variables were supplied. No rule helper or language blocker remains. Raw Go JSON event streams are retained as `.jsonl.gz` beside the readable logs.
