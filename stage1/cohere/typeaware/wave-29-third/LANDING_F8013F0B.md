Built: rebased wave 29 onto current main f8013f0b; owned checks now accept an explicit fresh compiler.
Commits: rebased source b29e4ac3; final evidence commit reported in handoff; only codex/typeaware-wave-29 is pushed.
Commands: setup 127s, nproc 5; six full rule comparisons, two graph kernels and numeric metadata pass against fresh builds.
Mutants: six rule verdicts, provenance, released handle, four graph kernel mutations and eighteen metadata mutations caught by their independent checks.
Uncovered: full React source pipelines and numeric per-node handlers remain blocked; no full repository gate or new claims.

All origin heads were fetched explicitly. Main advanced from e8ba3d5d to
f8013f0baac41ddc340d76f83bddde38536a8f07. Rebase completed without conflicts.
The previously pushed own tip was 027916315d9c44b2e87d630ac6425aefb6800658;
the push uses an exact lease on that own branch to publish the requested rebase.
No push to main or area branches is performed. No Diagnostic landing sha was
provided in the instruction, and the shared finding model was not changed here.

The continuation comparison initially used its old default compiler artifact.
That preliminary run is excluded from landing evidence. Three owned scripts
now accept ADAMIC_COMPILER, preserving the existing default. A separate
`go build -o /workspace/wave29-latest-kernel-adamic ./cmd/adamic` produced the
current compiler. The continuation, metadata, gap and graph-kernel checks use
that explicit artifact. Configured naming and the Go default test build their
own fresh current compiler. No shared harness or registration generator edits.

Observed results:

- Configured naming: 272 cases, 209 findings, 70068 bytes identical to Go;
  denylist and match mutants compile and exit zero, caught only by bytes.
  Instrumented native and bridge agree without sanitizer diagnostics; the
  RegExp control agrees with Node/native/emitted JavaScript/sanitized native.
- Second batch: 400 cases in seven profile groups, 252 findings. Globals,
  setter and shadow mutants compile and exit zero with empty stderr and differ
  from production Go. Compiler 77 roots/5241 bytes and repository 287 roots/
  18485 bytes agree, both zero findings, including full bridge/native sanitizers.
  Native/Go process observations: compiler 2.086631035s/0.415756321s (5.02x),
  repository 0.309890236s/0.156803926s (1.98x).
- Original batch: TestWave29AgreementAndMutants PASS 195.039s. Default controls
  have 15 findings/9922 bytes; header paths account for the changed byte count.
  Controls and both frozen corpora agree with Go under sanitizers. Denylist,
  match, race end offset and provenance mutants compile/exit zero and differ;
  the retained/released handle mutant is caught by the required panic check.
  Native/Go observations: compiler 1.866598144s/0.323183084s (5.78x), repository
  0.283165951s/0.145810496s (1.94x). Concurrent validation is not a quiet benchmark.
- Static-components kernel: 34 supplied graphs, 27 findings, 14419 identical
  bytes. Store-binding and phi-creator mutants each compile/exit zero and change
  one compared finding; sanitizer output agrees. This is not the full source rule.
- Control kernel: 17796 graphs, including 28 production source graphs; 243119
  bytes match Go/native/Node/emitted JavaScript/sanitized native. Throw-as-exit
  mutant changes 3143 results; dropped switch-case tests change 4575 results.
  Both compile/exit zero with empty stderr; only independent comparison catches them.
- Nine manifests and native listener declarations: 315 bytes match actual Go
  registrations, Node and sanitized native. Nine valid JSON wrong-kind mutants
  and nine compiling native declaration mutants are caught only by comparison.

The fresh React blocker probes remain positive on Go and fail to supply a native
pipeline. Go emits one finding for each of three JSX inputs. Native parses the
effect/render JSX as TypeAssertionExpression with no JSX nodes; static-components
self-closing JSX panics at slash offset 83 with exit 70. Two hook-only witnesses
also produce one Go finding each and parse natively, but there is no native React
SSA lowering/checker/capture/analysis entry. The shared-ssa branch still changes
only docs/shared-flow.md, a design document. The published harness ParseNode
still exposes string kind only. Existing exported listener arrays and numeric
rule.json declarations do not make legacy whole-file run methods per-node
handlers; the incoming driver must not invoke them per node.

No new claims are taken while these three source ports and numeric handler
migration remain incomplete. Full findings/fixes/suggestions, full-rule mutants
and native-versus-Go lint timing remain unavailable for the three React rules.
No full repository gate, inherited ten-rule repeat or standalone raw-fact/lifetime
suite was rerun. The original rule gate's released-handle mutant did rerun.

Reproduction (all output goes to log files):

```sh
source /workspace/adamic-tools/env.sh
python3 -u stage1/cohere/typeaware/wave-29-configured/check.py /workspace/wave29-regex-controls > /tmp/wave29-latest-configured.log 2>&1
ADAMIC_COMPILER=/workspace/wave29-latest-kernel-adamic python3 -u stage1/cohere/typeaware/wave-29-next/check.py /workspace/wave29-latest-next-fresh > /tmp/wave29-latest-next-fresh.log 2>&1
ADAMIC_WAVE29_ARTIFACTS=/workspace/wave29-latest-default ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript go test ./stage1/cohere/typeaware -run '^TestWave29AgreementAndMutants$' -count=1 -timeout=30m -v > /tmp/wave29-latest-default.log 2>&1
ADAMIC_COMPILER=/workspace/wave29-latest-kernel-adamic python3 -u stage1/cohere/typeaware/wave-29-third/check_static_core.py /workspace/wave29-latest-static > /tmp/wave29-latest-static.log 2>&1
ADAMIC_COMPILER=/workspace/wave29-latest-kernel-adamic python3 -u stage1/cohere/typeaware/wave-29-third/check_control_core.py /workspace/wave29-latest-control > /tmp/wave29-latest-control.log 2>&1
ADAMIC_COMPILER=/workspace/wave29-latest-kernel-adamic python3 -u stage1/cohere/typeaware/wave-29-third/check_listeners.py /workspace/wave29-latest-listeners > /tmp/wave29-latest-listeners.log 2>&1
ADAMIC_COMPILER=/workspace/wave29-latest-kernel-adamic python3 -u stage1/cohere/typeaware/wave-29-third/prove_gaps.py /workspace/wave29-latest-gaps > /tmp/wave29-latest-gaps.log 2>&1
```

Setup: Go 1.27.1 ready 0s, clang 20.1.8 ready 1s, Node 24.19.0 ready 1s,
submodules ready 1s, cache warm 127s, total 127s; nproc 5, CPU quota four cores,
17.6 GB. Evidence is in validation/landing-f8013f0b, including command records,
rebase/fetch logs and compared graph/metadata streams.
