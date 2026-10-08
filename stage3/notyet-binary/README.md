# Binary expressions across mixed representations

[BREAKDOWN.md](BREAKDOWN.md) groups the five original lowering rows by operator and by the checker's types of each operand and the whole expression. [sites.csv](sites.csv) preserves the exact failing node and source expression; [selected-raw.csv](selected-raw.csv) preserves all original attempting contexts.

The original compiler and this unit both pin cohere to `7945d102a6c18dd36adf9114a758ce646e8b2359`. The working branch starts at the resolved `origin/area/compiler` SHA `8cadb46576e2de70791bc60c1e48694cd70e5d87`. Replay `9a1f14c5d994aa855625e7cfa295677060348fec` was merged cleanly in `5c66582926617416cd11b9b63d9981e4e1608d11`.

The scratch instrumentation runs on table pin `dc6b1529ae9d2a2210672e105c8bb6374619a59d`, at combine's actual failing node. Merely parsing every binary at a printed position overcounts: nested binary expressions can share their start. No production source is changed by the analysis. Targeting the original CSV's 1,123 attempting units retains the full checker project, ancestor bindings, signature preparation and statement rollback. The join rejects missing or ambiguous sites and requires the original five row totals exactly.

Reproduction, from this branch with the setup environment sourced:

```sh
git show dc6b1529:stage3/notyet-table/raw.csv > /tmp/notyet-binary-raw.csv
python3 stage3/census/latent/make_overlay.py /tmp/notyet-binary-census-base /tmp/notyet-binary-exact-overlay > /tmp/notyet-binary-exact-prepare.log 2>&1
python3 stage3/notyet-binary/instrument.py /tmp/notyet-binary-census-base /tmp/notyet-binary-exact-overlay /tmp/notyet-binary-raw.csv
```

The isolated base is a detached checkout of `dc6b1529`, using the matching cohere checkout and the lockfile-installed `stage3/api/node_modules`. In that checkout:

```sh
python3 stage3/meter/entry_overlay.py /tmp/notyet-binary-exact-overlay /tmp/notyet-binary-target-entry > /tmp/notyet-binary-target-entry.log 2>&1
go build -buildvcs=false -overlay=/tmp/notyet-binary-target-entry/overlay.json -o /tmp/notyet-binary-target-census ./stage3/census/latent/tool > /tmp/notyet-binary-target-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/notyet-binary-target-census /tmp/notyet-binary-adapted/src/tsc/tsc.ts /tmp/notyet-binary-target-full.jsonl > /tmp/notyet-binary-target-run.log 2> /tmp/notyet-binary-target-types.jsonl
```

Back on this branch:

```sh
python3 stage3/notyet-binary/build-breakdown.py /tmp/notyet-binary-raw.csv /tmp/notyet-binary-target-types.jsonl stage3/notyet-binary > /tmp/notyet-binary-breakdown.log 2>&1
```

The adapted tree comes from `bash stage3/apply.sh /tmp/notyet-binary-adapted`. All 81 source hashes match the table's saved source manifest. The first unfiltered pass was superseded by targeted measurement; its partial output is not used for coverage claims.

Before any production change, both original table examples reproduced on this branch using the required command:

```sh
go run ./stage3/census/latent/replay -project /tmp/notyet-binary-adapted/src/tsc/tsc.ts -where /tmp/notyet-binary-adapted/src/compiler/binder.ts:423:18 -kind NotYet -reason 'a BinaryExpression with a value and a value' > /tmp/notyet-binary-before-423.json 2> /tmp/notyet-binary-before-423.log
go run ./stage3/census/latent/replay -project /tmp/notyet-binary-adapted/src/tsc/tsc.ts -where /tmp/notyet-binary-adapted/src/compiler/binder.ts:1698:21 -kind NotYet -reason 'a BinaryExpression with a value and a value' > /tmp/notyet-binary-before-1698.json 2> /tmp/notyet-binary-before-1698.log
```

Both exit 0, reproducing the exact signature. The first reports load/register/lower 5.288456805s and total 38.889615157s; the second 4.003819753s and 5.766485521s. These are measurement results, not backend validation.

Setup command: `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh > /tmp/notyet-binary-setup.log 2>&1`, then `source /workspace/adamic-tools/env.sh`. Setup exits 0: Go ready 0.064s, Node ready 0.066s, clang ready 0.499s, markdown dependencies ready 1.226s, submodules ready 16.766s, build cache warm 221.332s, done 221.364s. `nproc=5`, CPU quota `400000 100000`; Go 1.27.1, clang 20.1.8, Node v24.19.0.

Conservative interpretation: the current documented October 7 ruling admits non-boolean conditions, superseding the original boolean-only rule. This unit uses existing ToBoolean support and does not change condition admission. Source-checker errors and existing representation/refusal boundaries remain in force. Counts describe isolated sites on a checker-rejected project, not complete programs or every child behind a failed boundary.

Current continuation and delivery status, including the c41c0e06 merge, exact per-group coverage limits, all mutants and pending conditional ownership, are in [FINAL.md](FINAL.md). Older delivery/deferred-work statements describe their respective commits.
