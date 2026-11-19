# Landing receipts

Tested commit `c902aa42d023026261836a113345c3957c821e46` on `32d927c43c08d2df0cd7f891ffa82942c8fa85ca`. The restoring mutant ran against the already measured entry tree. The census exited 0, the first assertion exited 2 because first-error mode returned the full entry despite its file selector, and the full-stream assertion exited 1 after confirming all 12 previously removed G02 sites returned.

The serial landing gates were `bash stage3/apply.sh`, `go test ./stage3/lane/adaptedtree -run '^TestProduct_Stage3AdaptedTree$' -count=1 -timeout 45m -v`, and the unfiltered `bash stage3/lane/run.sh`. Each exited 0. The lane's Oracle exited 1 with its expected public API baseline difference. The clean verdict needed no alternating main control.
