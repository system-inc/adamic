# react/no-set-state

Port of `cohere/internal/lint/rules/react/no_set_state.go`. Uses shared `enclosingComponentOf` and `skipParenthesesOptional`. Preserves parenthesized callees and receivers, private-name hash trimming, rejection of computed member access, React component membership and the exact callee finding range.

Ownership checked across every origin branch before implementation; no rule directory or ownership reservation existed. Helper consumer lists, archived helper claims and historical candidate inventories do not reserve rules.

All 38 upstream cases and the owned witness, including the `all` input set, match unchanged Go on Node, emitted JavaScript and ASan/UBSan native: 10,475 bytes per backend. The setState-name mutant is caught on all three backends.

Run `python3 stage1/cohere/lint/rules/react-no-set-state/testdata/run_selected.py` with the project toolchain environment loaded.
