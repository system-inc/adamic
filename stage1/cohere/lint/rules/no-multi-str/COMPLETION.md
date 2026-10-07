Built three claimed core literal rules in .a, with full immutable proposals and explicit unsupported-shape refusals.
Commits: claim 4a0352de, multiline 01ca8a69, decimal escapes 12b298cb, octal ac92122c; all on codex/lint-wave1-15.
Validation: 74 fixtures plus 705 file/rule comparisons, Go/Node/emitted JavaScript/sanitized native byte-identical; external sentinel passed.
Mutants: remove U+2029, force preceding-null flag, restrict octal digits to 0..7; all compiled and differed only in comparison output on all three backends.
Not covered: nine shared-parser inputs, decimal-escape shared finding serialization, default registry/harness integration, converging fixer and full gate.

The claim was pushed before any new source, after fetching 341 origin refs and
checking 52 distinct claim Markdown blobs. The original helper-ready list was
exhausted; these were the first eligible inventory syntax-ready entries.
`selection.json` records the origin snapshot. Prior owned implementations and
evidence were already pushed at 59274126. No shared generator, harness, compiler
or cohere submodule file was edited. All new Adamic source files are .a.

Implementation and observations

`no-multi-str` reads raw StringLiteral text for all four line terminators,
exempts every JSX parent kind and supplies no repair. Its constructor refuses
.tsx files containing `<` when the parser supplied no JSX nodes. This conservative
refusal also rejects otherwise valid generic-only .tsx files. It prevents the
measured JSX-expression false positive from the shared parser's recovered tree.
The gate can be narrowed after the shared parser produces correct JSX nodes.

`no-nonoctal-decimal-escape` scans raw strings in escape-sized steps, reports each
real backslash-8/9 separately and records exact independent finding ranges.
It offers both readings, with all three exact suggestions after a real null
escape and two otherwise. No automatic repair is supplied. The owned result
model retains every description, interval and edit; the current shared bridge
explicitly refuses these proposals instead of discarding suggestions or widening
an escape range to the enclosing literal. Future integration needs the shared
suggestion model and a finding API that accepts independent diagnostic ranges.

`no-octal` matches a raw leading zero followed by any decimal digit, including
8 and 9. It excludes strings, templates, modern prefixes, fractions and a
zero followed by an underscore. It supplies no repair.

Validation

Run from the repository root, with source /workspace/adamic-tools/env.sh:

    python3 stage1/cohere/lint/rules/no-multi-str/validate.py --scratch /tmp/wave15-fourth-final --typescript /tmp/lint-wave1-15-typescript --mutants --throughput > /tmp/wave15-fourth-final-validation.log 2>&1
    go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 > /tmp/wave15-fourth-oracle-sentinel.log 2>&1
    bash cloud/setup.sh > /tmp/wave15-fourth-setup.log 2>&1
    go run ./cmd/lint-registry > /tmp/wave15-fourth-registry.log 2>&1

The owned comparison builds virtual standalone Go entries through an overlay;
no shared source is replaced. The Go entry invokes the unchanged cohere rules,
walker and parser. The owned serialization compares complete finding IDs,
descriptions, UTF-8 byte ranges, suggestion arrays/edits, each individually
applied suggestion and automatic fixed source. It tests source Node, emitted
JavaScript and ASan/UBSan native generated from the same .a graph. Corpus output
omits repeated per-suggestion applied source but retains every suggestion edit
and the automatic fixed source. It does not certify the default CLI formatter,
all-rule ordering or converging fixer.

Unmodified Go test families TestNoMultiStr, TestNoNonoctalDecimalEscape and
TestNoOctal pass. Their prefix selection also runs TestNoOctalEscape, which is
not claimed here. The capture yielded 79 distinct upstream cases: 21 multiline,
27 decimal escapes and 31 octal. Nine excluded inputs leave 70 comparable cases;
three semantic witnesses and a Unicode/multiple-escape probe bring the fixture
manifest to 74. The final fixture output is 26,425 identical bytes on all four
sides. Original capture records and the full Go log are retained.

The corpus is TypeScript v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8:
77 src/compiler files and 158 current stage1 .ts/.a files, 235 files total,
705 file/rule pairs and 37,166,890 identical output bytes. Output hashes for all
sides and gzipped Go corpus output are committed in evidence/. No stderr is
accepted for successful comparisons or mutants.

Each mutant is compiled anew for emitted JavaScript and sanitized native and
run on all its applicable fixtures. All nine mutated backend executions exit
successfully with empty stderr and differing output. The null-flag mutant leaves
finding counts alone and changes only suggestions, establishing that a count/id
check would miss it. The external one-byte sentinel passed (0.311s).

Nine exclusions are reproduced individually against Go and every backend.
Go succeeds on all. The six JSX inputs either fail in the shared parser or hit
the new owned refusal; the three recovered numbers 01.5, 0777.5 and 0755n fail
with expected-semicolon parser errors before dispatch. Before the explicit gate,
a JSX element-content expression parsed with the wrong parent and produced a
false positive; the final gate catches it on all three backends. The full Go
output and each backend refusal are preserved. JSX-parent behavior and recovered
numeric-node behavior remain blocked on shared parser support.

All three backend bridge probes succeed for multiline and octal. Decimal escape
bridge probes exit 70 with NotYet: shared lint repair serialization. Directory
registration is described by rule.json; the current shared generator still
requires rule.ts. Its unchanged default command failed on .a discovery; registry
output is preserved. The fetched harness branch f4d98cab50048692781da3599131317dc569d466
has newer models, but is not silently substituted into this branch. Integration
and independent finding-range support remain unverified.

Setup failed while warming tests, specifically profile_test.go:32 cannot range
over portFiles, a function returning []string. Go 1.27.1, clang 20.1.8, Node
24.19.0 and submodules each reported ready in 0s; setup printed no successful
warm/done timing. nproc = 5. The cached toolchain was used for all successful
comparisons. No full gate was claimed.

Throughput

Best of three interleaved complete-process timings over the 77 compiler files
plus one 1,000-finding synthetic file per rule (78 files). The count includes
parser and rule evaluation and skips formatting, fixes and suggestions output.
Each backend returned exactly 1,000 findings. Native throughput uses an
unsanitized optimized binary; correctness and mutants use sanitized native.
Build time is excluded; process startup and file reads are included. Source
Node is the reported Node rate. These are observations on this five-core host.

no-multi-str: 78 files, 1000 findings, best of 3 Go 6058.03 findings/s (0.165070s), native 930.98 findings/s (1.074141s), Node 1252.80 findings/s (0.798212s)
no-nonoctal-decimal-escape: 78 files, 1000 findings, best of 3 Go 5758.54 findings/s (0.173655s), native 953.40 findings/s (1.048872s), Node 1186.46 findings/s (0.842844s)
no-octal: 78 files, 1000 findings, best of 3 Go 5648.09 findings/s (0.177051s), native 883.50 findings/s (1.131863s), Node 1259.61 findings/s (0.793898s)
