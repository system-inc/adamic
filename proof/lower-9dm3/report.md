Proof for Adamic task #9dmh4em, run 2026-11-10 UTC. **The proof is not fully green:** step 3 exited 1, and step 5's restored rerun exited 1. The lower regression tests and admission delta passed; both mutants produced the expected red tests.

HEAD: `747fde5150b74b15a51aa30e7ce03dbe848e5cdb` (`compiler/method-field-widening-3`), one commit above main `37a549e53ea10621ed79ca2387ab2eb6b3109440`.

Setup: loaded `/workspace/adamic-tools/env.sh` first. Created its missing work directory `/tmp/adamic-gate` outside the repository. Provisioned upstream `https://github.com/microsoft/TypeScript.git` at exactly `050880ce59e30b356b686bd3144efe24f875ebc8` in `/tmp/adamic-proof-9dm3/typescript` and exported `ADAMIC_TYPESCRIPT_SOURCE` to that checkout. Installed `stage3/api`'s lockfile packages outside the repository in `/tmp/adamic-proof-9dm3/api`; `tsc --version` on PATH returned `Version 6.0.3`. Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`. An external npm launcher re-exported these settings for gate-started npm after the gate's environment filtering. Ran one heavy command at a time. All prerequisite assertions passed.

**1. Checkout, formatting, vet — PASS.**

Ran `git fetch origin compiler/method-field-widening-3`; asserted FETCH_HEAD matched the requested HEAD; checked it out detached; ran `git submodule update --init --recursive`. All succeeded. Cohere: `5360753e19e4417675a90e92a1823bea66b13918`; nested TypeScript: `d92d9bfee114c80be2c375d72edae966176e3a4f`.

`gofmt -l internal/lower` printed nothing; the empty-output assertion passed. `go vet ./internal/lower` exited 0. The worktree was clean.

**2. Targeted lower and uncached oracle tests — PASS.**

`go test ./internal/lower -run 'TestMethodSeenAsWritableFieldIsRefused' -count=1 -v` exited 0. Under `TestMethodSeenAsWritableFieldIsRefused`, every subtest passed:

| Subtest suffix | `.a` | `.ts` |
| --- | --- | --- |
| `an_object_literal's_method` | PASS | PASS |
| `a_class's_method` | PASS | PASS |
| `a_readonly_field` | PASS | PASS |
| `a_method_seen_as_a_method` | PASS | PASS |
| `the_same_type` | PASS | PASS |

`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'StdoutHandleView' -count=1 -v -timeout 30m` exited 0. Under `TestStdoutHandleViews`, every subtest passed:

| Subtest basename | `.a` | `.ts` |
| --- | --- | --- |
| `n1_fd` | PASS | PASS |
| `n2_cast_chain` | PASS | PASS |
| `n3_annotated` | PASS | PASS |
| `n4_any_out` | PASS | PASS |

No failures or first errors. The uncached oracle reported Node cache hits=0, misses=8.

**3. Assertions, enums, objects fixtures — FAIL, exit 1.**

Command: `go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixtures(Assertions|Enums|Objects)$' -v`.

`TestFixturesEnums` passed. `TestFixturesAssertions` failed during oracle-hook preparation, before its fixture rows ran. Its whole first error:

```text
    products_test.go:145: python3 timed out: context deadline exceeded
```

`TestFixturesObjects`, `TestFixturesObjects/objects`, `TestFixturesObjects/objects/19_identifier_multimap.a`, and its `/stage0` subtest failed. These ancestor failures share this leaf's whole first error:

```text
    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/objects/19_identifier_multimap.a:87:33: Adamic 0.1 refuses a method read as a value (trackSymbol would lose its object, and this with it); call it in an arrow that keeps the object: (value) => its object.trackSymbol(value) (unbound-method) Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/objects/19_identifier_multimap.a:79:82: stage 0 can't lower a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) | undefined yet Stdout:}
```

Every other executed enum/object fixture subtest passed. No fixture status was updated.

**4. Lower mutant — detected, exit 1; restored cleanly.**

Read `internal/lower/method-field-widening.mutants.md`. Asserted its old text occurred exactly once in `internal/lower/invariance.go`; count was 1. Applied its `if false && ...` replacement and ran its command: `go test ./internal/lower -count=1 -run 'TestMethodSeenAsWritableFieldIsRefused' -v`.

The four red subtests were:

```text
    --- FAIL: TestMethodSeenAsWritableFieldIsRefused/.a/an_object_literal's_method (0.19s)
    --- FAIL: TestMethodSeenAsWritableFieldIsRefused/.ts/an_object_literal's_method (0.24s)
    --- FAIL: TestMethodSeenAsWritableFieldIsRefused/.ts/a_class's_method (0.30s)
    --- FAIL: TestMethodSeenAsWritableFieldIsRefused/.a/a_class's_method (0.31s)
```

Their errors, verbatim:

```text
    method_field_widening_test.go:38: got <nil>, want main.a:6:33: ...whose method fail becomes a writable field: a write of () => void would replace a method of type () => never
    method_field_widening_test.go:38: got <nil>, want main.ts:6:33: ...whose method fail becomes a writable field: a write of () => void would replace a method of type () => never
    method_field_widening_test.go:38: got <nil>, want main.ts:7:33: ...whose method fail becomes a writable field: a write of () => void would replace a method of type () => never
    method_field_widening_test.go:38: got <nil>, want main.a:7:33: ...whose method fail becomes a writable field: a write of () => void would replace a method of type () => never
```

All six controls passed. Restored `invariance.go` byte-for-byte; `git diff --exit-code` exited 0.

**5. Fixture 19 mutant — detected, exit 1; restored rerun also exited 1.**

Read `stage3/fixtures/objects/19_identifier_multimap.mutant.md`. Asserted the current line 58 was exactly:

```typescript
    trackSymbol?(symbol: Symbol, enclosingDeclaration: Node | undefined, meaning: SymbolFlags): boolean;
```

The assertion passed. Replaced line 58 with exactly:

```typescript
    trackSymbol?: (symbol: Symbol, node: Node | undefined, meaning: SymbolFlags) => boolean;
```

Verified the replacement, left `status.json` unchanged, and ran the file's command:

```sh
go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesObjects$' -v
```

Mutant exit and every red subtest, with its `fixtures_test.go` error line verbatim (the ancestors have no separate error):

```text
    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/objects/19_identifier_multimap.a:181:56: Adamic 0.1 refuses a value of type SymbolTrackerImpl seen as SymbolTracker | undefined, whose method trackSymbol becomes a writable field: a write of ((symbol: Symbol, node: Node | undefined, meaning: number) => boolean) | undefined would replace a method of type (symbol: Symbol, enclosingDeclaration: Node | undefined, meaning: number) => boolean; make trackSymbol readonly in the type it's seen as (readonly trackSymbol: ...), or declare it there as a method, which isn't written (adamic/invariant-mutable) Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/objects/19_identifier_multimap.a:79:82: stage 0 can't lower a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) | undefined yet Stdout:}
--- FAIL: TestFixturesObjects (3.24s)
    --- FAIL: TestFixturesObjects/objects (0.00s)
        --- FAIL: TestFixturesObjects/objects/19_identifier_multimap.a (0.36s)
            --- FAIL: TestFixturesObjects/objects/19_identifier_multimap.a/stage0 (0.00s)
MUTANT_EXIT=1
```

Restored the fixture byte-for-byte; `git diff --exit-code` exited 0. Reran the same command. Restored exit and error, verbatim:

```text
    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/objects/19_identifier_multimap.a:87:33: Adamic 0.1 refuses a method read as a value (trackSymbol would lose its object, and this with it); call it in an arrow that keeps the object: (value) => its object.trackSymbol(value) (unbound-method) Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/objects/19_identifier_multimap.a:79:82: stage 0 can't lower a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) | undefined yet Stdout:}
--- FAIL: TestFixturesObjects (3.39s)
    --- FAIL: TestFixturesObjects/objects (0.00s)
        --- FAIL: TestFixturesObjects/objects/19_identifier_multimap.a (0.21s)
            --- FAIL: TestFixturesObjects/objects/19_identifier_multimap.a/stage0 (0.00s)
RESTORED_EXIT=1
```

All other object fixture subtests passed on both runs. Restoration removes the mutant's `adamic/invariant-mutable` stop at 181:56, but HEAD still fails the recorded fixture status with `unbound-method` at 87:33.

**6. Admission delta — PASS, exit 0.**

Generated HEAD's complete manifest with `python3 cloud/admission-corpus/manifest.py --sha 747fde5150b74b15a51aa30e7ce03dbe848e5cdb`, saving it outside the repository. Ran alone:

```sh
go run ./cmd/adamic-admission-delta --base 37a549e53ea10621ed79ca2387ab2eb6b3109440 --head 747fde5150b74b15a51aa30e7ce03dbe848e5cdb --manifest /tmp/adamic-proof-9dm3/manifest.json --workers 2 --json
```

| Result | Value |
| --- | --- |
| Exit | 0 |
| Verdict | `pass` |
| Void | `false` |
| Void reasons | None; no report-level or per-program reason |
| `diff_count` | 1 |
| Admitted (newly accepted) | 0 |
| Newly refused | 0 |
| Programs | 5,191 |
| Accepted by both | 3,244 |
| Refused by both | 1,947 |
| Complete corpus | `true` |
| Runtime programs selected / omitted | 115 / 0 |
| Compile timeout limit | 45 seconds per command |
| Compile timeouts | 0 |
| Runtime timeouts | 0 |

The changed program was `stage3/fixtures/objects/19_identifier_multimap.a`. The gate reported 50 accepted negative witnesses, 52 ruled runtime stops, 0 ruled backend divergences, and 13 Node agreements. No compile timeout occurred at either SHA, so no standalone timeout rerun was required.

Exactly two programs had different classification stderr. Both remained `refused-by-both`. The following are the complete original stderr texts, including the gate's temporary checkout path and terminal newline:

`internal/oracle/testdata/stdout_handle_view/n3_annotated.a`

Main `37a549e5`:

```text
adamic: /tmp/adamic-gate/admission-delta-3979218897/head/internal/oracle/testdata/stdout_handle_view/n3_annotated.a:3:94: Adamic 0.1 refuses optional property fd in { setBlocking?: (value: boolean) => void; readonly fd?: number; } absent from structural source { setBlocking(blocking: boolean): void; }, which can hide fields; declare fd on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

HEAD `747fde51`:

```text
adamic: /tmp/adamic-gate/admission-delta-3979218897/head/internal/oracle/testdata/stdout_handle_view/n3_annotated.a:3:94: Adamic 0.1 refuses a value of type { setBlocking(blocking: boolean): void; } | undefined seen as { setBlocking?: (value: boolean) => void; readonly fd?: number; } | undefined, whose method setBlocking becomes a writable field: a write of ((value: boolean) => void) | undefined would replace a method of type (blocking: boolean) => void; make setBlocking readonly in the type it's seen as (readonly setBlocking: ...), or declare it there as a method, which isn't written (adamic/invariant-mutable)
```

`internal/oracle/testdata/stdout_handle_view/n3_annotated.ts`

Main `37a549e5`:

```text
adamic: /tmp/adamic-gate/admission-delta-3979218897/head/internal/oracle/testdata/stdout_handle_view/n3_annotated.ts:3:94: Adamic 0.1 refuses optional property fd in { setBlocking?: (value: boolean) => void; readonly fd?: number; } absent from structural source { setBlocking(blocking: boolean): void; }, which can hide fields; declare fd on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

HEAD `747fde51`:

```text
adamic: /tmp/adamic-gate/admission-delta-3979218897/head/internal/oracle/testdata/stdout_handle_view/n3_annotated.ts:3:94: Adamic 0.1 refuses a value of type { setBlocking(blocking: boolean): void; } | undefined seen as { setBlocking?: (value: boolean) => void; readonly fd?: number; } | undefined, whose method setBlocking becomes a writable field: a write of ((value: boolean) => void) | undefined would replace a method of type (blocking: boolean) => void; make setBlocking readonly in the type it's seen as (readonly setBlocking: ...), or declare it there as a method, which isn't written (adamic/invariant-mutable)
```

After both mutants and the admission gate, `git diff --exit-code` and `git diff --cached --exit-code` both exited 0. Only `proof/lower-9dm3/report.md` was then committed on the new branch `proof/lower-9dm3-report`, created from the tested HEAD, and pushed to `origin/proof/lower-9dm3-report`. No other file was committed and no other branch was pushed.
