# Stage 3 meter 20261103T200643Z

The area label repeats the pinned main tree for reporter compatibility.

| Run | Main SHA | Whole program | Own file | Lowering | tsc entry diagnostics |
| --- | --- | ---: | ---: | ---: | ---: |
| Current | a0ba9dd52b23326880b16a7898d446597cfb5a88 | 79/79 | 79/79 | 2/79 | 0 |
| 20261103T131714Z.YWfwfQ | 0080ee95cf1c71dfadd636fef55189467033dfa0 | 2/79 | 78/79 | 2/79 | 2 |

Lowering counts ordinary successful IR lowering. No native output is claimed.

Area reuses main's measurement, not a second measurement. main/ was copied to area/ before reporting and gzip; both labels use main-adapted. See compiler-mode.json measurement_sources.

Elapsed: 69m 3s.

## Measurement record

Measured origin/main: `a0ba9dd52b23326880b16a7898d446597cfb5a88`. It contains `fa47f0d8`.
Launcher code commit: `48a7287651d8ab7cb835e969ddedacf8a9270db0`.
Daily command used the README environment and timeout, with STAGE3_METER_MAIN_REF set to the recorded origin/main SHA.
Exit: 0. Wall: 4143 seconds (69m 3s). Cgroup limit: 17179869184 bytes (16 GiB).
Cgroup memory.peak: 15942144000 bytes (14.847279 GiB). No OOM or OOM kills.
One-second sampled peak: 15923269632 bytes. The environment reconnect interrupted the outer sampler; the original timeout and meter continued. The resumed observer attached strace to that timeout and recorded exit_group(0). Sampling has an explicitly recorded gap; the cgroup peak counter continued throughout.

| Measure | Main and copied area |
| --- | ---: |
| Whole-program checker diagnostics | 0 |
| tsc entry checker diagnostics | 0 |
| Own-file diagnostics summed | 0 |
| Files past checker, own file | 79/79 |
| Files past checker, whole program | 79/79 |
| Files lowering | 2/79 |
| Compiler latent NotYet distinct sites | 560 |
| Compiler latent Refused distinct sites | 4294 |
| Compiler latent panics / errors | 0 / 0 |
| Entry latent NotYet distinct sites | 207 |
| Entry latent Refused distinct sites | 4240 |
| Entry latent panics / errors | 2279 / 0 |
| Entry resolved sources | 81 |

Both latent streams have 0 skipped dependencies. Counts deduplicate (kind, where, reason, text).
The separate 79-root ordinary aggregate reports `lower: stage 0 compiles a program from one entry file, got 79`; its checker diagnostic count is still zero. Three JSON inputs are non-source input errors and excluded from the 79-file denominator.

### First lowering stops for all 77 non-lowering source files

| Kind | Reason | Count | Owner from owners.json |
| --- | --- | ---: | --- |
| Refused | a generator function | 73 | OWNER BLANK |
| NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | 2 | OWNER BLANK |
| NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | 1 | codex/namespaces-tsc |
| Refused | a cast the runtime can't check | 1 | 01a11410 |

Every affected file, exact first stop, and fix is listed in [measurement-summary.json](measurement-summary.json).
Every panic and error reason/text is retained verbatim in [failures-verbatim.json](failures-verbatim.json), including all 2279 entry panic sites and the ordinary aggregate/non-source errors.
The main and area report trees are identical; all five copied streams match byte for byte after decompression. See [alias verification](validation/alias-verification.json).

### Validation

Compiler tests: 9 passed. Reporter tests: 25 ran, 2 optional binary probes skipped. Entry tests: 11 ran, 2 optional binary probes skipped.
The second-measurement mutant is caught by `8 != 4 : main must be measured exactly once`.
The full Python discovery ran 46 tests, with 4 skips and one pre-existing generator probe failure: its temporary probe cannot import github.com/system-inc/adamic/internal/ir. The broader daily mutant runner also stops at this unrelated generator failure after catching the launcher mutants.
The exact requested `go test -count=1 ./stage3/meter/...` exits 1 with `go: warning: "./stage3/meter/..." matched no packages` and `no packages to test`. There are no Go tests reading twice-daily.sh or report.py.
All command logs and exit codes are under [validation/](validation/). Setup exited 0. report.py is unchanged.
