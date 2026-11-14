# B full sys.ts changes

## lowered

```json
[
  {
    "identity": [
      "src/compiler/sys.ts",
      "309ba1c45fee15a5bed3e275e908fddcbd7c7caf0a32519b43c0fe9d3bd57573",
      68165
    ],
    "base": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1467:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1517:9",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1588:191: stage 0 can't lower process.recordreplay as a value (use its supported process operation directly) yet",
      "start": 68155,
      "end": 72577
    },
    "head": []
  },
  {
    "identity": [
      "src/compiler/sys.ts",
      "309ba1c45fee15a5bed3e275e908fddcbd7c7caf0a32519b43c0fe9d3bd57573",
      71057
    ],
    "base": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1467:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1588:191",
      "reason": "process.recordreplay as a value (use its supported process operation directly)",
      "text": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1588:191: stage 0 can't lower process.recordreplay as a value (use its supported process operation directly) yet"
    },
    "head": []
  }
]
```

## moved

```json
[]
```

## new_findings

```json
[
  {
    "identity": [
      "src/compiler/sys.ts",
      "309ba1c45fee15a5bed3e275e908fddcbd7c7caf0a32519b43c0fe9d3bd57573",
      72238
    ],
    "finding": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1467:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1612:21",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1517:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet",
      "start": 72216,
      "end": 72306
    }
  },
  {
    "identity": [
      "src/compiler/sys.ts",
      "309ba1c45fee15a5bed3e275e908fddcbd7c7caf0a32519b43c0fe9d3bd57573",
      72353
    ],
    "finding": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1467:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1613:46",
      "reason": "reading modulePath",
      "text": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1613:46: stage 0 can't lower reading modulePath yet"
    }
  },
  {
    "identity": [
      "src/compiler/sys.ts",
      "309ba1c45fee15a5bed3e275e908fddcbd7c7caf0a32519b43c0fe9d3bd57573",
      72328
    ],
    "finding": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1467:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1613:21",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1613:46: stage 0 can't lower reading modulePath yet",
      "start": 72306,
      "end": 72397
    }
  }
]
```

## unit_first_stop_delta

```json
[]
```
