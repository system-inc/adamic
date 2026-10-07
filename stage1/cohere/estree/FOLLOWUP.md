Follow-up on codex/stage1-estree, baseline 80f1ded.
Largest refusal cause first: JSX rescues 523 files without new disagreements.
The source census has 15,222 matches, 92 output mismatches and 461 acceptance disagreements.
UTF-8 loss and parser stalls receive separate checkpoints below.
This remains an incomplete ESTree port; no compiler/runtime files were edited.

Baseline: 27,369 frozen files, 14,699 identical, 861 refused Go answers, 123
accepted Go refusals, 92 output mismatches and 11,594 both refused. Refusal
includes Go's caught canonical-serialization/location panics, not just syntax
errors. Saved inventory and original answers are reused; these counts concern
that frozen corpus, not a new inventory of the added implementation files.

Setup: `bash cloud/setup.sh > /tmp/estree-followup-setup.log 2>&1` passed.
Timing lines: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s;
build cache warm 19s; done 19s on 5 processors. `nproc` = 5. Versions:
Go 1.27.1, clang 20.1.8, Node 24.19.0. Main and Go pins unchanged.

### JSX checkpoint

The local parser is a copy of main 5d4c801's parser, with imports redirected to
its unchanged scanner/statements/grammar/lookahead dependencies. JSX parsing
and conversion live inside the claimed ESTree directory. Shared parser,
compiler and runtime are untouched. Go's parse.go selects JSX for .tsx/.jsx and
TSX for .js/.mjs/.cjs; matching both was necessary to rescue all 523 files.

Source audit command:
`node --disable-warning=ExperimentalWarning testdata/corpus.mjs
/tmp/estree-baseline-go.jsonl /tmp/estree-jsx-all-corpus.jsonl
/tmp/estree-jsx-all-identical > /tmp/estree-jsx-all-audit.log 2>&1`.
It records 15,222 identical, 338 refused Go answers, 123 accepted Go refusals,
92 different outputs and 11,594 both refused: 461 acceptance disagreements.
No old matching file regressed, and no new wrong output/acceptance appeared.
The audit-only progress guard still bounds the unported recovery stalls.

Nine generated JSX files match 11,077 canonical bytes on source Node,
sanitized native and emitted JS, held to Go. Pinned original libraries prove
empty generic argument rejection, exact ampersand spelling differences after
postprocessing and the existing CRLF normalization difference. All other raw
and postprocessed JSX fields are compared. The extra selfClosing mutant
finishes normally on Node/native and is caught at line 23; the three original
port mutants are rerun at this checkpoint.

Retained logs under validation/followup/ provide commands and observed outputs.

Green commands: `ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library
 go test -count=1 -v ./stage1/cohere/estree > /tmp/estree-jsx-green-final.log 2>&1`
passed in 173.050s; all original checks, three original mutants and the extra
JSX mutant pass. `ADAMIC_ESTREE_CORPUS=/tmp/estree-jsx-rescued go test -count=1
-v ./stage1/cohere/estree -run '^TestRepositoryAgreement$' >
/tmp/estree-jsx-all-rescued.log 2>&1` passed in 24.955s: all 523 rescued files,
6,131,940 canonical bytes on source Node, ASan/UBSan native and emitted JS.
The initial library comparison failed on the documented deltas; the final gate
contains their exact proving checks. Cohere checks 21/21 sources, 100% ready.

### Decorated export checkpoint

JSX commit fe07100524fe56da82b15d8d8ff50358ff85c92d was pushed. The next
largest output group had 47 files whose first difference was an export wrapper.
Ignoring decorator children when reading modifiers and starting the wrapper at
the export keyword resolves 46 files; the remaining file has another difference.
The source audit is now 15,268 identical, 338 refused Go answers, 123 accepted
Go refusals, 46 different and 11,594 both refused. No old match regressed.

`go test -count=1 -v ./stage1/cohere/estree -run '^TestDecoratedExport' >
/tmp/estree-exports-gate.log 2>&1`: PASS 39.380s, six generated files and
successful wrapper-range mutant caught on source Node and sanitized native.
`ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library
ADAMIC_ESTREE_CORPUS=/tmp/estree-exports-rescued go test -count=1 -v
./stage1/cohere/estree -run '^TestRepositoryAgreement$|^TestDecoratedExportLibraries$'
> /tmp/estree-exports-rescued.log 2>&1`: PASS 20.277s, all 46 files and 251,078
bytes on all three port builds, six raw and postprocessed library agreements.
The source audit command is the JSX command with exports output paths.
Cohere reports 21/21, 100 percent ready. Logs and full dispositions are retained.

### Parser stall checkpoint

Export checkpoint 068218be92c9caa19e9ce981f178abf599371632 was pushed.
The local Parser.next now bounds repeated identical scanner positions at 32.
The audit injection has been removed entirely, so source census protection is
now the same parser used by the real driver and native build. The census stays
15,268 identical / 338 refused Go answers / 123 accepted Go refusals / 46
mismatches / 11,594 both refused. Exactly 13 real progress refusals remain:
seven Go-accepted and six Go-refused files. This fixes termination, not Go's
recovered AST or diagnostic equivalence for the seven accepted files.

`go test -count=1 -v ./stage1/cohere/estree -run
'^TestBoundedPortParser$|^TestPortStallControl$|^TestGeneratedAgreement$' >
/tmp/estree-stalls-gate.log 2>&1`: PASS 59.127s; 78 generated matches, three
minimal EOF stalls, and a disabled-guard control caught by the 500ms deadline
on source Node and sanitized native. That control is a liveness mutant, not
a successful wrong-output mutant. `go test -count=1 -v ./stage1/cohere/estree
-run '^TestBoundedPortParser$' > /tmp/estree-stalls-recorded.log 2>&1`: PASS
21.871s, all 13 original recorded stalls plus the three minimal cases explicitly
refuse with empty stdout before 2s on source Node, sanitized native and emitted
JS. All 13 inputs are retained under validation/followup/stalls/ as .input files.
`gaps/portParserRecovery.ts` is the smallest local-parser proving program.
The older shared-parser timeout test remains a dependency-gap observation.
