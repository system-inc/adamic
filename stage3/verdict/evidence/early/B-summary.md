# Stage 3 verdict

| Suite | Pass | Fail | Excluded | Deferred |
|---|---:|---:|---:|---:|
| acceptance | 300 | 1 | 0 | 0 |
| tiny | 0 | 1 | 0 | 0 |
| baselines | 39 | 1 | 6539 | 5865 |

Acceptance includes tiny; the separate tiny row repeats that project.

First difference in acceptance: `tiny`

```json
{
  "stdout": {
    "byte_offset": 18,
    "expected_size": 192,
    "actual_size": 192,
    "expected_context": "b\"argument.ts(2,6): error TS2345: Argument of type 'string' is not assignable to parameter of type '\"",
    "actual_context": "b\"argument.ts(2,6): Error TS2345: Argument of type 'string' is not assignable to parameter of type '\""
  }
}
```

First difference in tiny: `tiny`

```json
{
  "stdout": {
    "byte_offset": 18,
    "expected_size": 192,
    "actual_size": 192,
    "expected_context": "b\"argument.ts(2,6): error TS2345: Argument of type 'string' is not assignable to parameter of type '\"",
    "actual_context": "b\"argument.ts(2,6): Error TS2345: Argument of type 'string' is not assignable to parameter of type '\""
  }
}
```

First difference in baselines: `tests/cases/compiler/ArrowFunctionExpression1.ts`

```json
{
  "stdout": {
    "byte_offset": 35,
    "expected_size": 119,
    "actual_size": 119,
    "expected_context": "b'ArrowFunctionExpression1.ts(1,10): error TS2369: A parameter property is only allowed in a constructor implementati'",
    "actual_context": "b'ArrowFunctionExpression1.ts(1,10): Error TS2369: A parameter property is only allowed in a constructor implementati'"
  }
}
```

