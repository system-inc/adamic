The one-unit eligibility probe confirms first-error-stops-the-unit.
Main: 74fb6490, with meter hook 855e1585 merged at 2596bcba.
Only checker.ts:1486:1 has its body-diagnostic lookup replaced with none.
Lowering records one Refused, var at checker.ts:1490:5; no crash.
Separately, the AST refusal scanner records 4,912 Refused findings in that unit.

createTypeChecker spans bytes 46,466 through 3,144,000: 3,097,534 bytes.
Its baseline status is skipped_checker_body. The probe status is attempted,
while the program remains checker-rejected. LATENT_ASSERT_NO_OUTPUT=1 passed;
production Load and Lower stay disabled and no backend is called.

The syntax scanner's many findings do not mean the lowering attempt continues:
its sole lowering finding is the first var declaration, four lines into the body.
This is the continuation wall for part 2.

The baseline and probe use the same compiler sources, adaptations and dependency
pins. A scratch Go workspace reuses byte-identical, clean dependency checkouts
already warmed on this box; shared.work.txt lists every path override.
No production compiler file is edited. The probe is produced by probe_unit.py.

Commands:
```sh
python3 stage3/census/latent/make_overlay.py "$PWD" BASELINE_OVERLAY > overlay.log 2>&1
python3 stage3/census/latent/probe_unit.py BASELINE_OVERLAY PROBE_OVERLAY
go build -buildvcs=false -overlay=PROBE_OVERLAY/overlay.json -o PROBE ./stage3/census/latent/tool > build.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 LATENT_PROBE_UNIT=/workspace/stage3-latent-full-adapted/src/compiler/checker.ts:1486:1 PROBE /workspace/stage3-latent-full-adapted/src/compiler/checker.ts probe.jsonl > probe.log 2>&1
```

The production compiler gate and native execution are outside this measurement.
Toolchain setup is still warming the full checkout; its completed timings will
be included with the final run. Node 24.19.0 is first on PATH; nproc=5.
