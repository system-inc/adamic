# YAML native speed work

## Shared schema checkpoint

The composer compiled both fixed schema tables on every formatting call.
It now shares their completed pattern trees for the module's lifetime. Matching
only reads those trees; document state and diagnostics remain per composer.
An initial release comparison measured 3.555806s before and 1.267965s after
(2,819.61 to 7,907.16 texts/s), with all 992,282 saved Go answer bytes unchanged.
These single observations are not the final five-round speed comparison.

The composer/schema/formatter/driver/mutant suite passed in 156.388s, including
sanitized native, source Node, emitted JavaScript and independent pinned libraries.
All 36 repository files and 10,026 formatter cases still match Go; the same 42
independently proved published-Prettier differences remain. Nine existing mutants
(composer, schema and printer) compile and finish with empty stderr; byte
comparisons catch each. Cohere's 276-rule check reported 42 files and 100%
Adamic-ready, and formatting passed.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=20m ./stage1/cohere/yaml -run 'TestCompose|TestSchema|TestFormatter|TestFileDriver|TestBundledParser' > /tmp/stage1-yaml-speed/shared-schema-suite.log 2>&1
```

Logs: [suite](audit/speed-shared-schema-suite.log),
[lint](audit/speed-shared-schema-lint.log),
[format](audit/speed-shared-schema-format.log).

## Runtime/compiler cost proving program

[gaps/stringUnitScan.ts](gaps/stringUnitScan.ts) computes the same UTF-16 checksum
using either one-unit slices or direct numeric unit reads. It consumes a file,
so the scanned strings are built at runtime. On `('abc中😀' repeated 4000)` and
100 rounds, both native variants and source Node print the same checksum.
The counted native build reports:

```text
slice:   allocations 2400007 frees 2400007 retains 105 releases 2400110 peak 7 regions 0
numeric: allocations       7 frees       7 retains 105 releases     110 peak 7 regions 0
```

Observed: 2.4 million unit slices make 2.4 million allocations and releases;
the numeric reads avoid them. Runtime `string.c` clamps and maps both UTF-16
slice boundaries; `string_share.c` copies slices shorter than 64 bytes, except
whole-string slices. This explains this probe's allocation cost. It is a
performance cost, not a demonstrated correctness bug. Compiler/runtime files
are outside this unit's scope and remain unchanged.

```sh
go run ./cmd/adamic build stage1/cohere/yaml/gaps/stringUnitScan.ts -o /tmp/stage1-yaml-speed/unit-scan --count
/tmp/stage1-yaml-speed/unit-scan /tmp/stage1-yaml-speed/unit-scan.txt slice 100
/tmp/stage1-yaml-speed/unit-scan /tmp/stage1-yaml-speed/unit-scan.txt numeric 100
node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/cohere/yaml/gaps/stringUnitScan.ts /tmp/stage1-yaml-speed/unit-scan.txt slice 100
```

The existing [shared slice append correctness gap](gaps/sharedSliceAppend.ts)
remains separately held by its existing test. This optimization does not use
string append or alter that runtime behavior.
