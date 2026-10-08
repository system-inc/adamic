The red run uses compiler/stage3-front-2 80fb9b75ec951e50aee46bfe07d609802d9ae021
and area/stage3 4ad53a4ee46c25e876414391081f51ac4b80103f. The unpushed scratch
merge is e224e66. Compiler build succeeds with its original SDK pins; parser
build stops at six missing Node binding diagnostics. Exact first failure and
minimal reproduction are in report.json. No native parser output exists.

The fresh raw slice is byte-audited before applying validated temporaries 60-64.
Full-tree adaptation pipeline excludes scanner slice-only passes 52-59/80-86
and scanner 50-51 (the compiler front supplies fallthrough/implicit returns).
It includes area/stage3's permanent passes 00-47 and 70. The five parser temps
are applied separately after gathering, with stock emitted-JS identity guards.
Counts distinguish the entry's 1,990 declarations from the driver's 1,991.
Node uses the unchanged /tmp/parser-adapted10 corpus, 81 regular compiler files,
not the adapted/sliced text as input; all three measured outputs match the
required original SHA. User time uses getrusage(RUSAGE_CHILDREN) deltas per
single subprocess; timings include loading, transpilation and printing.

Commands, each with stdout/stderr redirected to a separate log:

```sh
# In detached scratch: merge newest area/stage3, then compiler/stage3-front-2.
# Build exact pinned SDK using scratch absolute module replacements, GOWORK=off.
GOWORK=off go build -mod=mod -o /tmp/parser-front-adamic ./cmd/adamic
SLICE_TYPESCRIPT=STOCK_TS_6_0_3 bash stage3/slice/run.sh TREE ENTRY_SLICE src/compiler/parser.ts:createSourceFile
SLICE_TYPESCRIPT=STOCK_TS_6_0_3 bash stage3/slice/run.sh TREE DRIVER_SLICE src/compiler/parser.ts:createSourceFile src/compiler/parser.ts:forEachChild src/compiler/program.ts:flattenDiagnosticMessageText src/compiler/utilitiesPublic.ts:unescapeLeadingUnderscores src/compiler/types.ts:ScriptKind src/compiler/types.ts:ScriptTarget src/compiler/types.ts:SyntaxKind
node stage3/slice/verify.cjs DRIVER_SLICE
# Apply adapters 60-64 only; each CENSUS_TYPESCRIPT=STOCK_TS_6_0_3 node adapt.cjs DRIVER_SLICE.
PARSER_TYPESCRIPT=STOCK_TS_6_0_3 bash stage3/drivers/parser/run.sh DRIVER_SLICE NODE_OUT --inputs FIXED_CORPUS
PARSER_TYPESCRIPT=STOCK_TS_6_0_3 python3 stage3/drivers/parser/native-proof.py COMPILER SCRATCH DRIVER_SLICE FIXED_CORPUS NODE_OUT NEW_NATIVE_OUT
COMPILER build stage3/drivers/parser/native-node-binding.a -o MINIMAL_BINARY
COMPILER build stage3/drivers/parser/native-output-control.a -o CONTROL_BINARY
```

The native comparator control is a separate tiny executable. One byte of its
stdout is flipped, with length preserved; cmp exits 1 versus the original.
Self comparison exits 0. These facts do not substitute for the unrun parser
output mutant or the unmeasured native parser performance and binary size.
