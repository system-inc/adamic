# Full sys.ts census delta

New: 0; gone: 1; changed: 0. Main 59 unique; candidate 58; restoring mutant 59.

## Gone

```json
{
  "finding": {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "src/compiler/sys.ts:1383:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "src/compiler/sys.ts:1385:31",
    "reason": "a method read as a value (writeFile would lose its object, and this with it)",
    "text": "src/compiler/sys.ts:1385:31: Adamic 0.1 refuses a method read as a value (writeFile would lose its object, and this with it); call it in an arrow that keeps the object: (path, data, writeByteOrderMark) => sys.writeFile(path, data, writeByteOrderMark) (unbound-method)"
  },
  "phases": [
    "refusal_scan"
  ],
  "attempts": [
    {
      "unit": "src/compiler/sys.ts:1383:1",
      "phase": "refusal_scan",
      "actualWhere": "/workspace/scratch/write-file-view-1385/main/src/compiler/sys.ts:1385:31"
    }
  ]
}
```
