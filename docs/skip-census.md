# Skip census

The checked-in table is JSON, inside the new package's testdata directory, so
both the scanning test and log checker read exactly the same declarations.
There is no independently maintained Markdown twin to drift. The AST scan
includes all Adamic-owned `_test.go` files, including overlay files in testdata,
and ignores build tags. The pinned cohere submodule is a separate repository;
its own tests are not enumerated by Adamic's gate census.

```sh
go test -count=1 ./internal/skipcensus/... > /tmp/skip-census-tests.log 2>&1
go run ./internal/skipcensus/cmd -scan > /tmp/skip-census-source.json
go run ./internal/skipcensus/cmd gate-out/test.jsonl > /tmp/skip-census-log.txt 2>&1
```

The command checks the current source against the declaration table before
reading the log. The function `skipcensus.CheckLog` can also be called by a gate.
Unknown or ambiguous named skips fail closed. Package-level `[no test files]`
events do not count as test skips. This checker judges skips, not test failures,
coverage, or whether a complete gate finished; retain the ordinary gate verdict.

Each row records its source file, lexical enclosing function, diagnostic line,
all enclosing if guards, read calls, reachable top-level test callers, skip
message, class and provision. Helper sites list all callers. The identity is
those test names followed by SHA-256 of go/printer's condition text; file scopes
the identity. For helpers with no known test caller the lexical name is used.
If initializers are present, their printed text is included so two `err != nil`
guards for different tools remain distinct. Else guards are negated. An
unguarded skip has condition `true`. Lines are ignored by validation; moving a
site does not require an inventory update. Identical conditions in one test
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
on its own shard, or its explicit fixture scope; it is not a benchmark. GCC uses
ADAMIC_GCC_LANE=1 and the shipping-build lane uses ADAMIC_RELEASE_LANE=1.
Their missing tools are failures after opt-in. Release runs restrict their scope
to fixtures that finish on Node; ordinary oracle verification covers the others.
Provisioning details are in [gate-inputs.md](gate-inputs.md).
No result cache is added: ADAMIC_GATE_UNCACHED=1 and ordinary mode compute the
same answers directly from source and log.

To update the table, scan and review changes, retaining explicit classifications
and providers. Do not blindly classify new skips as measurement. Run the census
test after editing. `ADAMIC_SKIP_CENSUS_ROOT` lets the scanning test inspect a
scratch source copy for mutation proofs while retaining the checked-in table.
