These four `.go.txt` files are byte-for-byte copies of the rule overlays carried
by scout `0d078a8da915d349e3fc70831fd213287afcf017`. The scout attributes
`function_bind.go` to Language's `ba619e28`, and `cycles.go`, `graph_records.go`
and `records.go` to Memory's `b529db90`. They are independent rule heads, not a
claim that either is merged into main. They are generated into scratch only.

Main's orchestration and other production files remain at the worktree's base.
The lower.go copy also receives the `graphRecords` field those heads need. Main
at the example base does not contain the scout's b01 declared-parameter refusal;
the pinned bind overlay is necessary for the exact requested controls.
