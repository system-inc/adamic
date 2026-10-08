Go output mismatches fall from 92 to zero on 27,369 frozen files.
Acceptance disagreements fall from 984 to five, all proved raw-input refusals.
15,647 files match 549,618,641 bytes on Node, sanitized native and emitted JS.
All 11,717 Go refusals reject on native; the full gate and 22 driver mutants pass.
Pushed checkpoints, commands, outputs and gap programs are recorded in FOLLOWUP.md.


# Non-JSON ESTree slice

The claim and origin survey are in this directory's `CLAIM.md`. This unit uses
cohere at `715ba94f3608a6500086b1076ce5cb7e51b836db`; `parse_json.go` is excluded.
The ESTree-local parser adapts the main and recovery branches; the shared scanner and node model remain unedited. The
node table owns its nodes; all child links are numeric indexes, so there are
no owning parent/child cycles. Own fields retain insertion order and absence
is distinct from an explicitly stored null. The visitor table preserves all
337 type entries from Go, including entries not yet produced by this driver.

`main.ts <file.ts>` prints the canonical converted tree to stdout.
`main.ts --manifest <file>` reads one input path per line. This boundary is
an ESTree tree, not formatted JavaScript source. Output includes every own
field and field order, node ranges in UTF-8 bytes, content ends, parenthesis
flags, printer-visible locations, ignored-node semicolon behavior, merged
comments and byte-preserving stripped source. Strings use escaped UTF-16
units. The independent Go driver calls the unmodified public ESTree API
through an overlay, with no oracle implementation copied into the port.

The initial generated cases cover directives, ASI and semicolons, scalar,
regexp and template literals, omitted elements, object properties, spread,
assignment patterns, all expression operator categories, optional chains,
parenthesis boundaries, non-null and type assertions, basic declarations,
if/while/do/with/labels, break/continue/debugger, return/throw, ordinary and
nestled JSDoc comments, hashbangs, Unicode and CRLF, keyword/reference/array/
indexed/tuple/conditional/union/intersection types and type operators.

## Checkpoint validation

`go test -count=1 -v ./stage1/cohere/estree` runs generated agreement and
three successful mutants. Full output is in `validation/step1.log`.
The member mutant reverses computed access; the postprocess mutant removes
logical rebalancing; the JSDoc mutant changes the merged comment value.
Compilation failures, sanitizer crashes and nonzero exits do not qualify.

The focused cohere pass checks all source files and reports 100 percent
Adamic-ready. Original libraries are installed at pinned versions in scratch:
`@typescript-eslint/typescript-estree@8.65.0`, `typescript@6.0.3`, and
`prettier@3.9.6`. Raw conversion matches all 75 cases. Postprocessing matches 72, with the three
exact differences below held by tests.

Setup command: `bash cloud/setup.sh > /tmp/stage1-estree-setup.log 2>&1`, then
`source /workspace/adamic-tools/env.sh`. `nproc`: 5.

```text
setup: go ready (1s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (72s)
setup: done in 72s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

## Limits

Source extensions other than TypeScript are refused at this checkpoint.
Unrepresented conversion kinds fail explicitly. Generic wrapper positions are
reconstructed from parser child spans and source tokens, then checked against Go.
The native parser does not promise diagnostic parity, JSX or full JSDoc
parsing. Parser list ranges are not part of its current node interface.
No compiler or runtime files were edited. No complete repository gate or
throughput measurement has been run yet.

## Checkpoint 2

Expanded to declarations, function/arrow parameters, destructuring and defaults,
annotations and generic wrappers; class/interface members, heritage, constructors
and static blocks; imports, exports, enums, modules, mapped and predicate types,
loops, switch and try/catch. No parser files were changed.

`ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library go test -count=1 -v
./stage1/cohere/estree > /tmp/stage1-estree-step2.log 2>&1` passed in 68.324s.
All 75 generated files matched 112,528 canonical bytes in four implementations.
All three port mutants finished normally on Node and sanitized native and were
caught again. `TestOriginalLibraries` checks all 75 against the pinned original
converter and Prettier parser; only the exact three documented deltas are allowed.
`TestPostfixValueGap` holds the successful Node answer and stage 0's NotYet.
Full output: `validation/step2.log`; cohere: `validation/step2-cohere.log`.

A complete checkout census found 27,359 JS/TS/Adamic source files: 478 Adamic,
472 cohere and 26,409 TypeScript, totaling 63,017,023 source bytes. Nine files
are not UTF-8. This is an inventory, not a claim of successful conversion.
Whole-checkout agreement and throughput are still pending.

## Checkpoint 3

Added import-type options, `typeof this` qualifiers, tagged invalid escapes,
Go numeric parsing and arbitrary precision bigint spelling; JavaScript extensions
use the Babel boundary and closure type-cast comments. Added the pure `answer`
pipeline, parent indexes and cached UTF-8 offsets for stripped ContentEnd lookups.
Literal conversion, simple node mappings and template escape validation have
separate homes. No parser, compiler or runtime files were changed.

The generated gate passed in 101.899s: 78 files, 120,355 canonical bytes, source
Node, ASan/UBSan native and emitted JS all match Go. Three qualifying mutants
finished normally and were caught on both source Node and sanitized native.
The pinned raw original library matches 78/78; the postprocessed wrapper matches
75/78 with only the three previously proved deltas. Gap tests hold information
loss in the input API, parser recovery stalls and explicit refusals for malformed
recovered identifiers, unpaired cooked surrogates and JSX.

The final snapshot audit examined 27,366 files. Source Node matched Go on 14,697;
860 Go-accepted inputs were refused, 123 Go-refused inputs were accepted, 92
outputs differed, and both refused 11,594. The audit alone instruments the parser
with a repeated-scan guard to classify 13 stalls without exhausting the Node heap.
The actual source driver imports the unchanged parser and can still stall. This
is an incomplete port and does not meet the whole-checkout validation contract.
The frozen inventory and every audit disposition are retained under `validation/`.
The optional repository test independently checks the identical subset with the
uninstrumented driver, sanitized native and emitted JS; it does not turn exclusions
or mismatches into passes.

A failed interface-based converter refactor also exposed a compiler gap. Its
minimal program is `gaps/interfaceDefault.ts`: source Node and emitted JS print 5;
ASan reports a stack-buffer-overflow in native. The release build happened to print
4, an observation of undefined behavior, not a portable expectation. The port
uses concrete model classes and simple tables instead. See `GAPS.md` and the
retained gap-test log. This gap is not one of the three qualifying port mutants.

## Final audit and throughput

The inventory was extended with the three final unit source files: 27,369 source
files total, with nine malformed UTF-8 files. Of these, 14,699 match Go on the
uninstrumented source driver, sanitized native and emitted JS, totaling
504,506,334 checked bytes. Both sides refuse 11,594; the port refuses 861
Go-accepted files, accepts 123 Go-refused files, and has 92 output mismatches.
There are 984 acceptance disagreements. These are failures of the whole-checkout
contract, not successful cases. Audit instrumentation detected 13 recovery stalls.
The two final matching files passed separately; the other final file was refused.

Eight additional scalar edge files add 5,969 Go-identical bytes, for 86 generated
files and 126,324 canonical bytes across Go, source Node, sanitized native and
emitted JS. The original-library gate checks 78 general files and five finite
scalar/bigint files: 83 raw agreements, 80 postprocessed agreements and the exact
three known postprocessed deltas. Two numeric range differences from the original
library are separately proved and documented in `GAPS.md`.

Release throughput medians: native 5,709, source Node 4,601, Go 11,345 texts/s.
Each run processes the 78-case manifest 20 times: 1,560 texts and 2,407,100 identical
output bytes. Includes startup, input reads, full canonical serialization,
file-backed stdout and the verification read. One warm-up and five rotated runs
per implementation. This measures the implemented ESTree tree driver, not whole
JavaScript formatting or parsing alone. See `validation/throughput.log` for every
sample, and `REPORT.md` for commands, commits and the precise limits.

## Follow-up: JSX

The largest baseline cause was the JSX extension refusal. Go also selects TSX
mode for .js/.mjs/.cjs; matching that dispatch rescues 523 frozen repository files.
The current source census is 15,222 identical, 338 refused Go answers, 123 accepted
Go refusals, 92 different, 11,594 both refused. The ESTree-local sourceParser.ts
copies the parser from main 5d4c801 and adds JSX through jsxParser.ts; shared
parser files remain unchanged. Conversion preserves fragments, names, attributes,
spread, comments, HTML entities, ranges and empty type argument lists. This does
not complete diagnostic parity. See FOLLOWUP.md for current checks and counts.

## Follow-up recovery

All 92 baseline output mismatches are resolved in the frozen source census.
The 190 newly matching files pass 15,512,882 bytes on source Node, sanitized
native and emitted JS. Generated recovery cases and three additional successful
wrong-output mutants cover the new parser/converter behavior. Thirteen baseline
stalls now terminate and follow Go's acceptance outcome. Source diagnostics and
remaining recovery cases still prevent full acceptance parity; raw input bytes
remain unavailable. Historical checkpoint claims above describe their own
revisions. FOLLOWUP.md is the current report.
