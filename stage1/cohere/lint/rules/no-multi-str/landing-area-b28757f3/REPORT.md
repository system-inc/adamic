Retained current main 71d7e491 and cleanly merged lint area b28757f3; no owned rule source changes.
TestOwnedWitnesses PASS: 127500 bytes match Go, Node, emitted JavaScript and sanitized native.
Eight owned TestMutants subtests PASS 348.565s: semantic mismatches caught on all three Adamic paths after successful compile and execution.
TestRulesAgree and TestNodeTableIsLinkOnly explicitly refuse case-936/Octal.ts at semicolon offset 10 after recovery classification; this is the recorded parser-recovery gap.
Full gate, 17 required external checks and throughput remain unverified; no new rule claimed.

Commands: source /workspace/adamic-tools/env.sh, then go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestNodeTable.*)$' -count=1 -v -timeout=20m; and TestMutants filtered to the same eight names as the preceding landing report. Complete raw logs are beside this report. The combined parity command exits 1 in 208.179s: TestRulesAgree FAIL 115.18s, TestNodeTableIsLinkOnly FAIL 51.20s on the same explicit recovery refusal, TestOwnedWitnesses PASS 41.78s. No guard was relaxed or bypassed. Ledger unchanged. New driver link guard, fixed-source path handling and fix-pass behavior were adopted unchanged from upstream.
