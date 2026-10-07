Rebased and rechecked the two retained CSS helpers on main e8ba3d5d81de4d3773c723914fccd4c76248b965.

Commands (source /workspace/adamic-tools/env.sh), all successful:

- python3 stage1/cohere/lint/helpers/wave15/css_string/validate.py
- python3 stage1/cohere/lint/helpers/wave15/css_declaration/validate.py
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/... -count=1 -v -timeout=20m
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v

parseCSSString: 198,584 queries, 5,648,824 byte-identical output bytes against unchanged Go on source Node, emitted JavaScript and sanitized native. Its escape mutant compiled and finished successfully with empty stderr on all three backends; only the byte comparison caught it.

parseCSSDeclaration: 8,836 queries, 619,181 byte-identical bytes on the same four sides. Its important-range mutant was likewise caught only by comparison on all three backends.

Each helper covers literals from every original fixture file of six consuming rules: better-tailwindcss/enforce-canonical-classes, better-tailwindcss/enforce-consistent-class-order, better-tailwindcss/enforce-consistent-variant-order, better-tailwindcss/enforce-shorthand-classes, better-tailwindcss/no-conflicting-classes, better-tailwindcss/no-unknown-classes. Their unchanged upstream Go tests pass. These helper implementations unblock those dependencies; they do not port the six consumers.

The foundation helper packages pass in 48.570s and 102.987s. The comment package reruns all five semantic mutants and the explicit JSX refusal guard mutant, caught by their independent comparisons. The external input oracle passes in 1.381s with six probe misses and zero hits.

This rerun supersedes the earlier e011f8f6 landing checkpoint, which remains recorded in landing-evidence. No new helper was claimed. The rule branch still has shared harness blockers, so the landing cap remains active. Full repository gate, default lint integration and unsupported JSX parsing are not certified by these results.
