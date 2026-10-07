Rebased eight surviving owned rules onto lint area b84a9d9314b65d3d0261ee017e233287b4f071da, containing main c7991b90.
Unchanged ledger applied; four losing copies remain retired, eight owned active rules retained.
Fresh unified TestOwnedWitnesses passes in 28.074s with 59204 identical bytes across Go, Node, emitted JavaScript and sanitized native.
All eight rule mutants are caught only by output comparison after successful execution on all three Adamic paths; gate passes in 154.777s.
No new claim: broad parity retains the Octal.ts parser blocker; full gate, required 17 checks and throughput remain unverified.

Commands with /workspace/adamic-tools/env.sh sourced: go test ./stage1/cohere/lint -run ^TestOwnedWitnesses$ -count=1 -v -timeout=20m and the eight named TestMutants subtests with identical flags. Logs retained beside this report. Owned source files unchanged. Shared merge conflicts took exact area files. The area delta changes proven relations and predicates in the compiler but not the lint parser, harness or ledger. The last broad failure remains recorded at landing-area-7481e032: case-906/Octal.ts, expected semicolon at 10. No broad green is claimed; no check was removed or relaxed.
