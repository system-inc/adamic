# Skip census

The declaration is the one-line `// census:` comment immediately above each
skip. The AST scan supplies classes directly to both TestCensus and CheckLog;
there is no central declaration table. The normative grammar and migration
proof are in [internal/skipcensus/README.md](../internal/skipcensus/README.md). The AST scan
includes all Adamic-owned `_test.go` files, including overlay files in testdata,
and ignores build tags. The pinned cohere submodule is a separate repository;
its own tests are not enumerated by Adamic's gate census.

```sh
go test -count=1 ./internal/skipcensus/... > /tmp/skip-census-tests.log 2>&1
go run ./internal/skipcensus/cmd -scan > /tmp/skip-census-source.json
go run ./internal/skipcensus/cmd gate-out/test.jsonl > /tmp/skip-census-log.txt 2>&1
```

The command audits every source annotation before reading the log. The function `skipcensus.CheckLog` can also be called by a gate.
Unknown or ambiguous named skips fail closed. Package-level `[no test files]`
events do not count as test skips. This checker judges skips, not test failures,
coverage, or whether a complete gate finished; retain the ordinary gate verdict.

Each row records its source file, lexical enclosing function, diagnostic line,
all enclosing if guards, read calls, reachable top-level test callers, skip
message, class and provision. Helper sites list all callers. The identity is
the lexical function name followed by SHA-256 of go/printer's condition text;
file scopes it. Discovering a new helper caller does not change the identity.
If initializers are present, their printed text is included so two `err != nil`
guards for different tools remain distinct. Else guards are negated. An
unguarded skip has condition `true`. Lines are ignored by validation; moving a
site does not require a declaration update. Identical conditions in one test
produce an identity collision and fail closed rather than silently sharing a row.

The scanner resolves direct helper calls in the same test package and testing
receiver parameter declarations, import aliases, local receiver aliases and nested callbacks. It is a source
census, not whole-program execution analysis: indirect function-value calls,
arbitrary external helper bodies are not resolved. Such
runtime skips fail as unknown unless they resolve uniquely to a declared site.
Read calls conservatively cover a whole function and its reachable test helpers;
read expressions retain symbolic paths rather than inventing concrete filenames.

`required-input` is the default for verification, debug bypasses and harness
inputs. `measurement` covers opt-in throughput and profile comparisons, never a
replacement for correctness. `not-applicable` states a platform or CPU witness
that cannot run on this host. `opt-in-lane` declares a verification lane selected
on its own shard while its opt-in is off; it is not a benchmark. GCC uses
ADAMIC_GCC_LANE=1 and the shipping-build lane uses ADAMIC_RELEASE_LANE=1.
After opt-in, missing prerequisites and fixture-scope exclusions fail the test,
preserving the original diagnostic. A lane that cannot verify a selected fixture
must report failure rather than a green result with a skip.

The AST scanner records `opt_in_off` for environment guards that disable a lane
and `opt_in_on` for skips reached after such a guard, including nested callbacks
and direct helpers. It resolves constant keys and local Getenv aliases, and
recognizes `Getenv(key) != "1"` and `Getenv(key) == ""`. TestCensus requires every
row reached after opt-in to be `required-input`, regardless of the variable name
or whether the caller is a benchmark. Source-derived annotations are compared
with the table, so removing an annotation cannot bypass this rule. Platform
skips before opt-in remain classified by their actual applicability.
Provisioning details are in [gate-inputs.md](gate-inputs.md).
No result cache is added: ADAMIC_GATE_UNCACHED=1 and ordinary mode compute the
same answers directly from source and log.

Add or update the annotation beside the call when changing a skip. Missing or
invalid annotations fail with the file and line. There is no remote table row
to update when another caller is added or a test is moved.
`ADAMIC_SKIP_CENSUS_ROOT` lets TestCensus inspect a scratch source copy for
mutation proofs. `go run ./internal/skipcensus/cmd -audit` audits without a log.
