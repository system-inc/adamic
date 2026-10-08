# react-hooks/void-use-memo

Port of `cohere/internal/lint/rules/react/void_use_memo.go`. Preserves lexical shadow checks, component/hook recognition, callback return detection and discarded comma-expression results. Uses the shared rules/react helpers.

Ownership checked across all 1,480 origin branches before implementation; no rule directory or ownership reservation existed. Helper consumer lists and historical candidate inventories do not reserve rules.

All 77 upstream cases and the owned witness, including the `all` input set, match unchanged Go on Node, emitted JavaScript and ASan/UBSan native: 38,803 bytes per backend. The return-detection mutant is caught on all three backends.

Run `python3 stage1/cohere/lint/rules/react-hooks-void-use-memo/testdata/run_selected.py` with the project toolchain environment loaded.
