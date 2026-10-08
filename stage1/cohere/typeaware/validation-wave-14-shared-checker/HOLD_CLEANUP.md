# Hold cleanup checks

423 archived fixtures renamed to .a.txt; every SHA-256 matches its pre-rename bytes and every old bare .a path is absent. No bare .a archive remains under validation-wave-14*.

The namespace descriptor directory is removed. `go run ./cmd/lint-registry` exits 0 and its emitted inventory contains no nexus/correctness-no-mock-on-module-namespace. `gofmt` was run on the edited owned overlay control; `git diff --check` passes. Logs are preserved beside this file.

No area merge, full lint run or new semantic mutant run was performed. Await the parser and JSX fixes as instructed. Earlier gate logs/results remain historical observations, not certification of this cleanup.
