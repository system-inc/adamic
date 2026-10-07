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

To update the table, scan and review changes, retaining explicit classifications
and providers. Do not blindly classify new skips as measurement. Run the census
test after editing. `ADAMIC_SKIP_CENSUS_ROOT` lets the scanning test inspect a
scratch source copy for mutation proofs while retaining the checked-in table.

## Degraded inputs under a passing test

The same AST inventory records `kind: "degraded-input"` for testing Log/Logf
and fmt Print/Printf/Println or Fprint variants directed to stdout, stderr or
`t.Output()`. It resolves constant text, concatenation and local aliases.
A diagnostic matches when a line names an ADAMIC_ variable and contains
`set ADAMIC_`, `not checked`, or `absent`. `variables` records the named inputs.
Only required-input and not-applicable are valid classifications for this kind;
each requires a provision or an explicit applicability reason.

Skip identities retain their original condition hash. Diagnostic identities also
hash the kind and message, so two logs under the same guard remain distinct.
Lines and columns do not participate in identity. Internal AST columns keep
same-line calls distinct while deriving opt-in annotations.

CheckLog counts these diagnostics when their emitting test passes. It prints
class, package, test, site identity, kind and named variables. Repeated output
from one site in one execution is counted once. Unknown and ambiguous diagnostics
fail closed. A known testdata overlay message can be attributed to the parent
package test that forwards its output; rewritten overlay filenames do not matter.
Go formatting substitutions are supported for diagnostic matching.
Package output from fmt is associated conservatively with active tests; paused
tests are excluded. Output chunks are reassembled before matching. The ordinary
gate verdict remains responsible for tests that fail or runs that never finish.
This is a declared diagnostic vocabulary, not arbitrary inference from prose.
