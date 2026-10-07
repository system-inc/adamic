# Runtime equality experiment, October 7, 2026

Built: a measured parse split and the isolated identical-header equality fast path; no parser/scanner changes.
Commits: 9cc0d58 records the split before fixes; 9fb0b85 imports equality-only 59a0850 and its runtime tests.
Commands and outputs: native/oracle uncached PASS; 77-file AST and full batch 8 byte parity PASS; 9.259G to 9.235G Ir.
Mutants: equality's identical-header branch returning false is caught by Node stdout; summary-only +1 is caught by self reconciliation.
Not covered: historical 8.97G boundary, exact per-token-class allocation Ir, full repository test gate, or either dated performance target.

## Result

| Same 77 files, release clang -O2 -g | Before | After | Saved |
| --- | ---: | ---: | ---: |
| Whole native parse process | 9,259,261,285 | 9,234,837,655 | 24,423,630 |
| Parser.file inclusive | 7,439,580,376 | 7,415,155,903 | 24,424,473 |
| Collect entry edge inclusive | 8,667,357,378 | 8,642,933,598 | 24,423,780 |
| String equality inclusive | 404,753,771 | 380,330,145 | 24,423,626 |
| String equality self | 335,614,089 | 339,705,054 | -4,090,965 |

The reduction is **0.263775%**, or a 1.002645 times instruction improvement.
Go parse alone remains the measured 1,684,637,174 Ir, so native is **5.48179x**.
The requested 1.5x and 1.0x dated limits are not met. No wall-time speedup is
claimed; instruction profiling overlapped tests, and no uncontended timing run
was needed for this instruction-count task.

The literal interning part already exists. The missing pointer shortcut removes
byte comparisons on equal literal/heap headers, but only 5.186% of kind equality
calls have identical headers. Most comparisons stop at differing lengths already.
The extra pointer comparison increases equality self work by 4.09M while reducing
its byte-comparator work enough to save 24.42M overall. Parent-map inclusive Ir
is unchanged. No extra intern table, string representation, token span, emission
hook, scanner change or numeric-kind adaptation is implemented.

`4ffed41` was unavailable after two fetches. The remote branch currently identifies
the equivalent change as 59a085087eefb2c3bb3f0c2cb2874784449d0506. Only that change
was cherry-picked, as 9fb0b85. The test file did not exist on main, giving a
modify/delete conflict; keeping the upstream file preserved its equality test
and its independent release-path test. No release optimization from the other
branch is included. This is an isolated equality experiment, not a measurement
of that branch's aggregate improvements.

## Parity and mutant

Full AST parity with typescript-go, including source positions and node kinds:
**44,766,682 identical bytes, all 77 compiler files**, also compared on Node and
ASan/UBSan native. This is TestWholeCompilerAgrees, not just node counts.
The full batch 8 release output also matches its Go oracle on every input:
**11,442,907 identical bytes**, before and after, with identical SHA-256s in
batch8-parity.log. This includes findings, ranges, edit payloads, suggestion
metadata and converged fixed source. The Go output is retained compressed.

The positive control TestRuntimeStringEquality passed against Node in both
release and sanitized builds. The mutant changes only the new same-header branch
to return false, using a heap string built at runtime. It compiled and ran, then
failed the intended stdout comparison:

```
native: 0 1 0 1 0 0
Node:   1 1 0 1 0 0
```

The run exited 1. It was not killed by a compiler warning or by refusal. The
runtime was restored before all subsequent tests and before the good binary was
built. equality-mutant.log preserves the complete failure; equality-control.log
preserves the passing control. The earlier accounting mutant increases only
the Callgrind summary by 1; the summarizer rejects it because self does not sum
to summary. Both mutants are reported separately from production test passes.

Complete native and oracle packages passed uncached:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -count=1 -timeout 30m > /tmp/parse-speed-native-oracle.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/typescript/parser -run '^TestWholeCompilerAgrees$' -count=1 -v -timeout 30m > /tmp/parse-speed-parser-parity.log 2>&1
gofmt -l cmd internal > /tmp/parse-speed-gofmt.log
go vet ./... > /tmp/parse-speed-vet.log 2>&1
```

Native package: `ok` in 127.724s. Oracle package: `ok` in 127.081s.
Parser corpus: `PASS` in 49.101s. Formatting and vet output are empty, exit 0.
A first focused vet invocation found the cohere-only Go profiling harness was
placed in a directory that root-module tools treated as a package. It is now
under testdata, like the existing oracle drivers, and is built only with the
cohere overlay. Repository-wide vet passed after that correction. No module
requirements were changed. gofmt formats that harness too.

## Reproduction

prepare.py reconstructs batch 8 from 4189abd, writes only `.a` Adamic sources,
checks the TypeScript pin and exactly 77 files, writes corpus hashes and creates
the Go overlay files. Its independent replay in a second scratch directory
produces the same corpus.sha256. The Go harness forces the line table; source
parents are already assigned by ParseSourceFile. Profiling flags and the exact
clang line are in SPLIT.md.

```sh
python3 internal/native/performance/parse-speed/prepare.py /workspace/scratch/parse-speed /workspace/scratch/typescript-6.0.3 > /tmp/parse-speed-prepare.log 2>&1
source /workspace/adamic-tools/env.sh
go build -o /workspace/scratch/parse-speed/adamic ./cmd/adamic > /tmp/parse-speed-compiler-build.log 2>&1
/workspace/scratch/parse-speed/adamic c /workspace/scratch/parse-speed/batch8/parse.a > /workspace/scratch/parse-speed/parse.c 2> /tmp/parse-speed-emit.log
```

Build Go binaries from cohere's directory:

```sh
go build -overlay=/workspace/scratch/parse-speed/overlay.json -o /workspace/scratch/parse-speed/go-batch8 /workspace/adamic/cohere/adamic_batch8_oracle.go > /tmp/parse-speed-go-build.log 2>&1
go build -overlay=/workspace/scratch/parse-speed/parse-overlay.json -o /workspace/scratch/parse-speed/go-parse /workspace/adamic/cohere/adamic_parse.go > /tmp/parse-speed-go-parse-build.log 2>&1
```

For exact before/after runtime isolation, compile the identical parse.c with the
runtime tree at 9cc0d58 for before and at 9fb0b85 for after, using SPLIT.md's exact
flags. All raw native/Go profiles and the generated baseline C are preserved
compressed. Counters run separately: instrument.py, token_instrument.py, then
classify_equality.py modify only an isolated scratch runtime and generated C.
Their paths are intentionally this workspace's paths; they do not edit compiler
or parser sources. Classifier-body self reconciliation and its small inclusive
libc relocation difference are explained in SPLIT.md.

## Next decision and limits

The runtime pointer path is correct and a small gain, but it is **not the top
multiplier**. The measured kind equality body is about 0.375G inclusive, while
release/destruction plus retains are 1.906G self and field-write plumbing is
1.317G self. Numeric kinds throughout scanner and parser could also remove kind
return ownership and dispatch work beyond equality; no numeric-kind speedup is
inferred from this runtime measurement. Constructor-only numeric conversion is
not implemented. Existing numeric-kind work and parser-parity work remain untouched.

Token strings are 660,939 allocations grouped by final kind, with speculation
and intermediate literal construction included. Ordinary punctuation already
has static text. The combined identifier/keyword source-slice edge is 175.91M
Ir, so eliminating it alone has a 1.90% whole-process ceiling. A span change
would need a separate semantic/lifetime design and scanner-owner clearance;
no scanner source has been changed. Exact allocation instruction attribution
by final keyword versus identifier category is still unmeasured. The report
contains the observed allocation counts and the measured combined instruction
edge rather than pretending those are per-category instruction counts.

The original historical parse boundary is still unresolved without #93z4yv7's
notes. No future schedule, numeric/scanner rewrite, full repository test pass,
or claim of meeting October 8/9 targets is made. The two requested runtime and
token-copy leads now have measurements for the next decision.
