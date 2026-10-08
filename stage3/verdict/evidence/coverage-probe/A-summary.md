# Stage 3 verdict

| Suite | Pass | Fail | Excluded | Deferred |
|---|---:|---:|---:|---:|
| acceptance | 301 | 0 | 0 | 0 |
| tiny | 1 | 0 | 0 | 0 |
| baselines | 6262 | 2 | 6180 | 0 |

Acceptance includes tiny; the separate tiny row repeats that project.

First difference in baselines: `tests/cases/conformance/types/members/duplicateNumericIndexers.ts`

```json
{
  "stdout": {
    "byte_offset": 0,
    "expected_size": 1285,
    "actual_size": 1380,
    "expected_context": "b'duplicateNumericIndexers.ts(4,5): error TS2374: Duplicate index signature for ty'",
    "actual_context": "b'../../../../stage3-verdict-adapted/built/local/lib.es5.d.ts(529,5): error TS2374'"
  }
}
```

