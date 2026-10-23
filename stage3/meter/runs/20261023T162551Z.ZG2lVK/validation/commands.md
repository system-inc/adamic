# Pinned daily run

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
STAGE3_METER_MAIN_REF=e50e502c21f12f9df64f7e716e980ed84695aa6f STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error timeout --kill-after=60s 7200 bash /workspace/adamic/stage3/meter/twice-daily.sh > /tmp/ck10df5-daily.log 2>&1
node --version > stage3/meter/runs/20261023T162551Z.ZG2lVK/node-version.txt
```

The branch was cut from fetched origin/main e50e502c21f12f9df64f7e716e980ed84695aa6f.
The guarded overlay tooling is c3a7effe64b5aaabc40142d4afcbc87a15ba6f37;
both per-ref compilers and adaptation inputs are the pinned main revision.
The launcher is the unchanged stage3/meter/twice-daily.sh at that main revision.
Setup and fixture/mutant evidence are retained in the tooling commit under
stage3/census/latent/refusalrewrite/evidence/overlay-drift/.
This is a first-error observational census, not a native compilation result.

The original daily command exited 0 in 1353 seconds. Both compressed latent streams were decoded and passed the existing lowering_summary validator against all 79 exact adapted source files; both summaries reproduce report.json. Both copies have equal reason rows after normalizing their distinct adapted-root paths. All gzip artifacts decode successfully. Raw .log and .jsonl files are excluded; compressed bytes preserve their contents. The README and paired report append the supplied Oct 22 comparison.
