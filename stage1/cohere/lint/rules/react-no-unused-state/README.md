# react/no-unused-state

Independent branch from lint-helpers/rules-react at fbaffdc53. A fresh audit of all 1,418 origin refs, every rule descriptor and claim-containing repository path found no existing rule directory or rule ownership claim. Historical helper-consumer inventory mentions do not reserve the rule.

The port preserves the pinned Go traversal, written-field insertion order, per-method aliases, lifecycle parameters, factory getInitialState last-statement rule, TypeScript unwrapping and whole-component give-up conditions. Messages and report ranges match Go. No fixes or suggestions are emitted.

161 unique upstream source/file/options cases plus the owned witness in selected and all-rule modes matched Go on source Node, emitted JavaScript and ASan/UBSan native: 80,877 identical output bytes. The compiling unused_state_read_inverted mutant is caught on all three backends. The initial inferred-never array failure was resolved with plain explicitly typed alias arrays.

Run, after sourcing the setup environment:

```sh
python3 stage1/cohere/lint/rules/react-no-unused-state/testdata/run_selected.py
```

The runner injects only this directory's selected test via a temporary Go overlay; shared harness and registry sources remain unchanged.
