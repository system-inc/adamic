# Entry latent checkpoint

First-error mode, LATENT_ASSERT_NO_OUTPUT=1; LATENT_TIMES and LATENT_PREPASS_TIMES set. Head has only the saved scratch seam. Findings are buffered until the entire pass finishes; a timed-out or running pass cannot supply a census of unit stops. Prepass totals for such a pass are partial and are not whole-pass comparison numbers.

## parent

```json
{
  "state": "bounded stop",
  "status": {
    "command": [
      "/workspace/scratch/k031ey0-5/parent-entry-census",
      "/workspace/scratch/k031ey0-5/adapted/src/tsc/tsc.ts",
      "/workspace/scratch/k031ey0-5/evidence/parent-latent.jsonl"
    ],
    "exit_code": -15,
    "seconds": 1780.220593903,
    "stopping_reason": "elapsed bound",
    "before": {
      "utc": "2026-11-13T10:23:53.608849+00:00",
      "hostname": "d312b35f425b",
      "cpus": 5,
      "load": [
        1.00439453125,
        1.20556640625,
        1.2373046875
      ],
      "cpu_max": "400000 100000"
    },
    "after": {
      "utc": "2026-11-13T10:53:33.829671+00:00",
      "hostname": "d312b35f425b",
      "cpus": 5,
      "load": [
        1.537109375,
        1.45263671875,
        1.33544921875
      ],
      "cpu_max": "400000 100000"
    }
  },
  "files_emitted": 0,
  "units_emitted": 0,
  "findings_emitted": 0,
  "new_reason_findings": [],
  "new_reason_units": [],
  "unit_timers": 3588,
  "latest_unit_timer": {
    "file": "/workspace/scratch/k031ey0-5/adapted/src/compiler/emitter.ts",
    "where": "/workspace/scratch/k031ey0-5/adapted/src/compiler/emitter.ts:6370:1",
    "wall_seconds": 0.196510353,
    "cpu_seconds": 0.19717000000036933
  },
  "prepass_calls": 3588,
  "prepass_seconds": 0.14272401,
  "stderr_prepass_errors": {}
}
```

## head

```json
{
  "state": "running snapshot",
  "status": null,
  "files_emitted": 0,
  "units_emitted": 0,
  "findings_emitted": 0,
  "new_reason_findings": [],
  "new_reason_units": [],
  "unit_timers": 1090,
  "latest_unit_timer": {
    "file": "/workspace/scratch/k031ey0-5/adapted/src/compiler/sys.ts",
    "where": "/workspace/scratch/k031ey0-5/adapted/src/compiler/sys.ts:1975:1",
    "wall_seconds": 0.159874094,
    "cpu_seconds": 0.16033699999996998
  },
  "prepass_calls": 1090,
  "prepass_seconds": 34.429085193,
  "stderr_prepass_errors": {}
}
```
