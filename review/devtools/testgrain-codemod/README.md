# Testgrain codemod

Run from the repository root, naming the setup/shard files for one package:

```sh
go run ./review/devtools/testgrain-codemod stage1/cohere/lint/node_table_oracle_floor_test.go
```

The Go AST/printer program recognizes the six initial wave families and the renamed node-table and text families. It mechanically wraps recognized shared preparations in `testgrain.Setup`, replaces recognized command factories with `Command`/`CommandContext`, removes recognized timing verdicts and timers, and places `Unit` at the old shard timer position. It refuses files already importing testgrain. It is a partial transformation: review every edit before running the package.

Hand-finish redundant sync.Once/state, tuple-return preparations, shard identities through Assign, union self-parsing through Union, and planted failures through CaughtByExactly. Move Unit after every shared setup call; setup captures must have no test-side timeout (use `-timeout=0` where appropriate). Preserve current oracle/product behavior and case counts. Remove unused imports and any remaining timer/deadline helpers; commands hidden behind other harnesses need separate review.

After finishing, gofmt the edited files, check `git grep -n 'AfterFunc\|cooked' -- <files>` is empty, compare before/after case counts, and run the package warm with an outer hard timeout and `go test -timeout 90s`. Review the diff for preserved preparation and shard boundaries. The codemod's own tests exercise the AST transformation, including callback preservation, the setup/Unit boundary, and variadic command forwarding.
