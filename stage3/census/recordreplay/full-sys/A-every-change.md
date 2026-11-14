# A full sys.ts changes

## lowered

```json
[]
```

## moved

```json
[
  {
    "identity": [
      "src/compiler/sys.ts",
      "30876cbae0e2cb19eae2525cbed919abdba86fac921ecc22fa6aaad0441a3398",
      68124
    ],
    "base": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1467:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1517:9",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191: stage 0 can't lower process.recordreplay as a value (use its supported process operation directly) yet",
      "start": 68114,
      "end": 72545
    },
    "head": [
      {
        "root_kind": "",
        "depth": 0,
        "measurement": "measured on a checker-rejected program",
        "unit": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1467:1",
        "phase": "lowering",
        "kind": "Boundary",
        "where": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1517:9",
        "reason": "failed statement; state rolled back; continued at next statement or declaration",
        "text": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191: stage 0 can't lower process.recordreplay with an unrepresented type yet",
        "start": 68114,
        "end": 72545
      }
    ]
  },
  {
    "identity": [
      "src/compiler/sys.ts",
      "30876cbae0e2cb19eae2525cbed919abdba86fac921ecc22fa6aaad0441a3398",
      71016
    ],
    "base": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1467:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191",
      "reason": "process.recordreplay as a value (use its supported process operation directly)",
      "text": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191: stage 0 can't lower process.recordreplay as a value (use its supported process operation directly) yet"
    },
    "head": [
      {
        "root_kind": "",
        "depth": 0,
        "measurement": "measured on a checker-rejected program",
        "unit": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1467:1",
        "phase": "lowering",
        "kind": "NotYet",
        "where": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191",
        "reason": "process.recordreplay with an unrepresented type",
        "text": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191: stage 0 can't lower process.recordreplay with an unrepresented type yet"
      }
    ]
  }
]
```

## new_findings

```json
[
  {
    "identity": [
      "src/compiler/sys.ts",
      "30876cbae0e2cb19eae2525cbed919abdba86fac921ecc22fa6aaad0441a3398",
      71016
    ],
    "finding": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1467:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191",
      "reason": "process.recordreplay with an unrepresented type",
      "text": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191: stage 0 can't lower process.recordreplay with an unrepresented type yet"
    }
  },
  {
    "identity": [
      "src/compiler/sys.ts",
      "30876cbae0e2cb19eae2525cbed919abdba86fac921ecc22fa6aaad0441a3398",
      68124
    ],
    "finding": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1467:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1517:9",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191: stage 0 can't lower process.recordreplay with an unrepresented type yet",
      "start": 68114,
      "end": 72545
    }
  }
]
```

## unit_first_stop_delta

```json
[]
```
