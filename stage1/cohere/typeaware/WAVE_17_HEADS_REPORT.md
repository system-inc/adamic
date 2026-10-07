Built: native decisions for both remaining Next Head rules, in `.a` files.<br>
Commits: this continuation follows `1330dbc2`; the commit containing this report holds the Head implementation.<br>
Commands/output: projected Head differential test PASS 54.198s; 40 controls, 21 findings, 10,044 identical bytes.<br>
Mutants: duplicate count threshold and Script spelling changes caught only by Go byte comparison, at bytes 46 and 456.<br>
Not covered: production JSX parsing still refuses; projected controls do not establish end-to-end JSX parity.

# Head decisions and the shared parser boundary

`no_duplicate_head.a` follows the production import and checker-declaration
identity rules, including foreign modules, aliases, namespace bindings, type-only
imports, shadowing and member tags. `no_script_component_in_head.a` resolves a
next/head default import and reports the Head tag for each direct literal Script
child. Neither production rule has fixes or suggestions. The existing native
suite now registers both rules locally; no shared parser, registration generator
or test harness was edited. No new checker question was necessary.

`projected_head_tree.a` and `wave_17_projected_suite.a` are validation-only files.
The independent Go oracle exports raw AST records (kind, source offsets, literal
text, import phase and children), not lint decisions. Adamic reconstructs these
nodes and makes every rule decision natively, using the actual checker program
and binding-declarations question. All 40 accepted upstream and targeted JSX
controls match the unchanged production Go rules normally and under ASan/UBSan.
Unicode and CRLF positions, direct versus nested children and multiple findings
are covered. The complete stream is 10,044 bytes and contains 21 findings.

The duplicate threshold mutant changes `occurrences > 1` to `> 2`; the script
mutant changes the literal Script tag test to S. Both sanitized binaries finish
with exit 0 and empty stderr. Only comparison with Go diagnostics catches their
wrong output (byte 46 and byte 456 respectively).

The production suite continues to use the shared native parser. A positive JSX
source still exits 70 with the parser slice refusal. The projection is deliberately
absent from that suite. Until the shared parser provides JSX nodes, these two
rules cannot meet the requested end-to-end parity bar. No performance claim for
positive JSX sources is made. The previously reported compiler/repository zero
finding measurements apply to the former suite; the regression run recorded
separately exercises the updated production suite on those frozen manifests.

Command (after sourcing /workspace/adamic-tools/env.sh):

```sh
ADAMIC_WAVE17_HEAD_ARTIFACTS=/workspace/wave-17-heads \
TMPDIR=/workspace/wave-17-artifacts \
go test ./stage1/cohere/typeaware \
  -run '^TestWave17HeadJudgmentsAndParserBoundary$' \
  -count=1 -v -timeout=20m > /workspace/wave-17-heads.log 2>&1
```

Setup on this continuation: Go, clang, Node and submodules ready 0s; build cache
warm 25s; done 25s. nproc: 5. Full repository gate was not run. Prior Nexus
released-handle and ownership tests remain recorded in WAVE_17_NEXT_REPORT.md.
The pre-existing request-refusal mutation anchor also remains an integration gap;
this continuation does not change that shared harness or the bridge dispatcher.
Compressed raw comparison and mutant outputs are in validation-wave-17-heads.

Updated production regression: PASS 70.166s. Compiler: 77 roots, 0 findings,
5,318 bytes; repository: 287 roots, 0 findings, 18,485 bytes. Both streams agree
under sanitizers as well. Single observed whole-process compiler time: native
1.857003s, Go 0.357492s; repository: native 0.233692s, Go 0.159871s. These
observations occurred during the validation work, not quiet median measurements.
The async end+1 mutant was caught at byte 52 and the retained released-handle
mutant was caught by the required panic 70. Vet exits 0 with an empty log.
