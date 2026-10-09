u045 M08k region-plan replay
Base: bc9edc5560379b63d4688a21006efb6032b60434
Submodule cohere pin: 7945d102a6c18dd36adf9114a758ce646e8b2359
Warm toolchain used: /workspace/adamic-tools/env.sh, Go 1.27.1. Setup skipped.
The only production change during mutant runs inverted plan.feedsRegion(statement) at internal/native/region.go:86. Source was restored and git diff --exit-code passed afterward. Neither tests nor counts.md were changed.
Each invocation used a unique fresh ADAMIC_BUILD_CACHE_DIR and ADAMIC_GATE_UNCACHED=1. Commands, exit statuses and monotonic wall seconds (including compilation and setup) are recorded in runs.json. All output was written to files outside the repository and copied into review only after testing completed.
Counts: base PASS, mutant FAIL only on the counts golden. 33 printed fixture rows moved. Deltas are measured minus recorded: allocations 0, frees +153, retains -1, releases -8, peak 0, in regions -153. All printed pairs are in counts-moved.json and result.json; complete JSON test output and decoded output are preserved. Other unchanged fixture rows have zero delta. This is a deterministic count change, not a demonstrated output or memory-safety disagreement.
Agreement selector: ^(TestNativeAgreesWithNode|TestInputAgreesWithNode)$. Both top-level tests and all their subtests completed without failures. No ASan, LeakSanitizer or UBSan reports were printed. The run contains 946 pass events including the package event.
All nine requested isolated tests passed. JSON also ran TestPortMatchesGoCohereSplit_Setup. No kill, deadline or assertion failure occurred, so the requested two base reruns for failing isolated tests were not needed. YAML's internal setup subprocess retained its own 90-second limit unchanged.
Disk check first: /tmp 6.2 GB free, /workspace 5.2 GB free. Deleted only previous completed unit scratch /tmp/M06i; recheck /tmp 6.5 GB free and /workspace 5.2 GB free. No repository or tools removed. Final disk check: /tmp 6.2 GB free, /workspace 4.8 GB free; no disk errors occurred.
Logs are gzip compressed without changing their contents. No code is committed on this evidence branch.
Counts failing line: counts_test.go:221: the counts changed and counts.md wasn't updated with them (go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts):
