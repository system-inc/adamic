Adamic static-walk proof — 2026-11-10

1. Checkout and setup: PASS.

Fetched `origin compiler/static-walk-per-edge-witness`, checked out detached HEAD `e8c4014fb8d1807880306603285783cb9ab7265b`, and completed `git submodule update --init --recursive`. HEAD’s parent is `6a13892f9b3113bb036e4d023f6f1aca57817a17`.

`gofmt -l internal/lower` exited 0 with empty stdout after adding the provisioned Go tools to PATH; the initial invocation could not find gofmt.

Created `/tmp/adamic-gate` outside the repository. Provisioned upstream TypeScript at `/tmp/adamic-proof-typescript`, verified commit `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE` to that path. Installed stage3/api’s unchanged pinned manifest and lockfile in `/tmp/adamic-proof-api`; its bin directory was on PATH and `tsc --version` printed `Version 6.0.3`. Go was 1.27.1; Node was v24.19.0.

2. Baseline: PASS, exit 0.

Command: `go test ./internal/lower -run 'TestStaticWalkPerEdgeWitness|TestStaticWalkReachAdmissions' -count=1 -v`.

| Suite | Subtest | Result |
|---|---|---|
| TestStaticWalkPerEdgeWitness | a declared function reading a later static through a getter | PASS |
| TestStaticWalkPerEdgeWitness | the same with the static declared first | PASS |
| TestStaticWalkPerEdgeWitness | an arrow in the static initializer handed to the constructor | PASS |
| TestStaticWalkReachAdmissions | assert reached with no verboseDebugInfo | PASS |
| TestStaticWalkReachAdmissions | assert reached with an outer function reading the class | PASS |
| TestStaticWalkReachAdmissions | assert reached with a static arrow field reading a later static | PASS |
| TestStaticWalkReachAdmissions | assert reached with a closure reading a later static | PASS |
| TestStaticWalkReachAdmissions | Node's Error.captureStackTrace | PASS |
| TestStaticWalkReachAdmissions | a program object's captureStackTrace | PASS |

No baseline failures.

3. “No per-edge check” mutant: detected.

Asserted that the staticParameterCalls declaration insertion anchor occurred exactly once, then inserted `return nil` as its first body statement, bypassing its per-edge checks.

Command: `go test ./internal/lower -run 'TestStaticWalkPerEdgeWitness' -count=1 -v`. Exit: 1.

| Subtest | Result |
|---|---|
| a declared function reading a later static through a getter | FAIL |
| the same with the static declared first | PASS |
| an arrow in the static initializer handed to the constructor | PASS |

The only red subtest’s complete diagnostic:

```text
got <nil>, want main.a:11:10: Adamic 0.1 refuses a static field read before its initializer has run: Version.later; declare the field earlier, or move the read after that field's initialization
```

Restored class_static.go byte-for-byte. `git diff --exit-code` exited 0 with empty stdout.

4. Third-row program: rejected at HEAD.

The passing baseline subtest verifies this refusal:

```text
main.a:2:35: Adamic 0.1 refuses a static field read before its initializer has run: Box.later; declare the field earlier, or move the read after that field's initialization
```

Ran the exact third-row source, stored outside the repository, with `node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/adamic-proof-third.a`.

Stdout:

```text
undefined
object
```

Exit: 0. Stderr: empty.

Because the program does not lower at HEAD, the conditional compiler build, native release, native sanitizer, and generated-JavaScript runs were not performed. There is no native stdout to report; no native result printing anything other than undefined was observed.

Publication: saved this report as `proof/lower-ssw/report.md`, committed only that file on new branch `proof/lower-ssw-report` based directly on the specified HEAD, and pushed only that branch to origin.
