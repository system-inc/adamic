# Stage 3 verdict

| Suite | Pass | Fail | Excluded | Deferred |
|---|---:|---:|---:|---:|
| acceptance | 0 | 301 | 0 | 0 |
| tiny | 0 | 1 | 0 | 0 |
| baselines | 0 | 6262 | 6182 | 0 |

Acceptance includes tiny; the separate tiny row repeats that project.

First difference in acceptance: `001_varianceCantBeStrictWhileStructureIsnt`

```json
{
  "exit": {
    "byte_offset": 0,
    "expected_size": 2,
    "actual_size": 2,
    "expected_context": "b'0\\n'",
    "actual_context": "b'1\\n'"
  }
}
```

First difference in tiny: `tiny`

```json
{
  "stdout": {
    "byte_offset": 0,
    "expected_size": 192,
    "actual_size": 0,
    "expected_context": "b\"argument.ts(2,6): error TS2345: Argument of type 'string' is not assignable to p\"",
    "actual_context": "b''"
  },
  "exit": {
    "byte_offset": 0,
    "expected_size": 2,
    "actual_size": 2,
    "expected_context": "b'2\\n'",
    "actual_context": "b'1\\n'"
  }
}
```

First difference in baselines: `tests/cases/compiler/2dArrays.ts`

```json
{
  "exit": {
    "byte_offset": 0,
    "expected_size": 2,
    "actual_size": 2,
    "expected_context": "b'0\\n'",
    "actual_context": "b'1\\n'"
  }
}
```

