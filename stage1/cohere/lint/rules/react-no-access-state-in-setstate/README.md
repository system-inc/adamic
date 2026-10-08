# react/no-access-state-in-setstate

Port of `cohere/internal/lint/rules/react/no_access_state_in_setstate.go`. Uses shared `enclosingComponentOf`. Preserves the upstream preorder traversal, accumulating method and variable records, first-argument membership, bare-name propagation, exact block scope identity, destructuring keys and duplicate finding behavior.

Ownership checked across every origin branch before implementation; no rule directory or ownership reservation existed. Mentions in helper consumer lists, archived helper claims and historical candidate inventories do not reserve this rule.

All 55 upstream cases and the owned witness, including the `all` input set, match unchanged Go on Node, emitted JavaScript and ASan/UBSan native: 26,482 bytes per backend. The receiver-detection mutant is caught on all three backends.

Run `python3 stage1/cohere/lint/rules/react-no-access-state-in-setstate/testdata/run_selected.py` with the project toolchain environment loaded.
