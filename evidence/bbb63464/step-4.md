# bbb63464: step 4, whole lower suite

Exact tested revision: bbb63464cd039548dacaaa7ccbde06915c291b8a. Go 1.27.1, clang 20.1.8, Node 24.19.0. ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source, pinned upstream TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. Compiler source restored exactly after steps 2 and 3 before this run.

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source go test -count=1 -timeout 40m -skip '^TestProduct_' ./internal/lower/
```

Exit 0. Test execution 1619.275 seconds. No --- FAIL lines, no failures to quote, no timeout.

Complete stdout:

```text
ok  	github.com/system-inc/adamic/internal/lower	1619.275s
```

Complete stderr (empty):

```text
```

[Steps 1–3](steps-1-3.md) contain every new-test stop and complete comparison/mutant output. Gofmt, vet, focused suite, main copied-test comparison, unguarded e5531c8 copied-test comparison and guard mutant all exit 0. The new pin does not go red with the guard removed: both unguarded and mutant cases still refuse at 13:14 with a member-typed-any diagnostic. Thus the pin does not demonstrate guard sensitivity, despite every test passing. Timer control stays green.

Final source diff against bbb63464 exits 0 and final source status is empty. All copied tests and temporary compiler edits removed.
