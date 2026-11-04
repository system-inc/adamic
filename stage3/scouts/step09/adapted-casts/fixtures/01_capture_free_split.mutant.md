Mutant evidence for adapted-casts fixture 01's stage-0 record in status.json.

What moved: castProof admits `<string>.split(<RegExp literal>[, limit]) as string[]` when the literal has no capturing group, named or unnamed (#b5w3ycg slice 1, internal/lower/cast_split.go). Fixture 01 (tsc's emitter.ts:4105) splits on `/\r\n?|\n/`, which has none, so its stage-0 record moves from Refused at the cast to Compiles, and its first line, the refusal's a-check header, goes. The body after the old header is unchanged, so body_sha256 stays 37faa7c3.

Natively built with `adamic build` and run, its stdout and stderr are byte for byte Node's (oracle/node.mjs), the record's node stdout, exit 0 on both.

Mutant: captureFreeSplit returns false before reading the call, the change undone; the header's removal and status.json stay.
Command: go test -p 1 -count=1 -timeout 30m ./internal/acheck -v
Result with the mutant: exit 1, red on exactly 01_capture_free_split.a. Restored, checked with cmp, the same command passes (linux/amd64, Node v24.19.0).

    acheck_test.go:91: stage3/scouts/step09/adapted-casts/fixtures/01_capture_free_split.a breaks its first line's promise: refused (adamic: stage3/scouts/step09/adapted-casts/fixtures/01_capture_free_split.a:4:19: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast))
--- FAIL: TestEveryProgramOutsideAPackageKeepsItsPromise (15.02s)
