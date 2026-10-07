Built: CFG enter in one .a helper file, preserving incoming/reachable state and current-block identity.
Commits: claim d6fbc1351 pushed before implementation; rule branch parked at 850fcdd0f.
Checks: real Go on all four consumer suites plus controls; source Node, emitted JavaScript and ASan/UBSan native agree; base package PASS 66.867s.
Mutants: retained prior reachability and retained prior current block compile/run; comparison alone catches both on all three backends.
Not covered: complete CFG/rule integration, other CFG helper functions, invalid nil-pointer panic bytes, full repository gate or speed claims.

Observed calls: array-callback-return 79 (632 bytes), consistent-return 123 (984 bytes), no-unreachable-loop 8,861 (70,936 bytes), react-hooks/rules-of-hooks 495 (4,309 bytes), controls 120 (960 bytes), identical on each backend. Four helper prerequisite edges removed across these exact four rules; zero complete rule blocker sets removed. The helper takes a handed block and reads no syntax kinds.

Base setup timings: Go/clang/Node/submodules ready 0s; cache warm and total 85s; nproc 5. Test output retained in evidence/enter-base.log, setup in evidence/setup.log. Capture runner uses actual Go private function bodies with post-call observation only, validates that each consumer suite ran and produced calls, and fails any Go test failure. No canned finding table or Go expected result is used by the runtime helper.
