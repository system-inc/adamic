# Latent census full audit, task #3dgsny7

Base `origin/main` was `6128c4e65e209aeef8804c39e66a19133590615c`. The tracing report on `typescript/tracing-sites-8777346a` at `fc2f6182` recorded the original failure.

## Baseline and stop

Built the documented overlay and ran `python3 stage3/census/latent/full_audit.py BINARY EVIDENCE_DIRECTORY`. It exited 1 with the following output.

```text
Traceback (most recent call last):
  File "/workspace/adamic/stage3/census/latent/full_audit.py", line 97, in <module>
    audit(binary, Path(sys.argv[2]).resolve())
  File "/workspace/adamic/stage3/census/latent/full_audit.py", line 78, in audit
    check(observations['full'])
  File "/workspace/adamic/stage3/census/latent/full_audit.py", line 20, in check
    assert actual == {'a string as a condition', 'a number as a condition', 'var'}, actual
           ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
AssertionError: {'var'}
```

The production `cmd/adamic` binary at the base exited 1 on `adamic build probe.a -o OUT` with the byte-identical audit fixture.

```text
probe.a:11:36: error TS2322: Type 'string' is not assignable to type 'number'.
probe.a:19:39: error TS2322: Type 'string' is not assignable to type 'number'.
probe.a:9:11: error TS2322: Type 'string' is not assignable to type 'number'.
```

The overlay recorded the same three checker diagnostics and continued measuring the checker-rejected program. Its lowering output for the two audited units was:

```text
three   Refused            var
three   Boundary           failed statement; state rolled back; continued at next statement or declaration
middle  SkippedDependency  a dependency function whose body has checker diagnostics (measurement skipped)
middle  Boundary           function signature or prologue failed; continued at next declaration
```

Production's whole-file checker gate prevents it from reaching either unit in that fixture. On the extracted `three` declaration, production reached `var` and exited 1. On the extracted `outerOnly` with only the ill-typed `grandBad` declaration removed, production also reached `var` in `middle` and exited 1. The exact overlay fixture yields an empty set of lowering refusals for `middle`. This blocks a production-backed update of the existing assertions on the exact fixture. The harness and mutant were skipped under the task's stop instruction.

## Lowering history and native checks

Production stage 0 was built from `./cmd/adamic` at all five revisions. The full fixture produced the same three TS2322 diagnostics on every revision. To exercise lowering independently of those errors, the exported `three` declaration was compiled on its own, and each condition was also compiled on its own with a branch that printed `taken` or `skipped`.

| Revision | Extracted `three` | String probe | Number probe |
| --- | --- | --- | --- |
| `61e3ce475` | Refused string condition | Refused `a string as a condition` | Refused `a number as a condition` |
| `cdddfee346` | Refused `var` | Native matched Node | Native matched Node |
| `73352e874` | Refused string condition | Refused `a string as a condition` | Refused `a number as a condition` |
| `5a2681b18` | Refused `var` | Native matched Node | Native matched Node |
| `6128c4e65` | Refused `var` | Native matched Node | Native matched Node |

`cdddfee346e70a5cba3e982f808b9ebe426a6dfb` changed both condition types to JavaScript ToBoolean through `censusCondition`. `5a2681b18d677d4c792fa6cc2ad90fd9715fd565` landed it on main. Native and Node both exited 0 and stdout diffs were empty for each accepted probe. The string inputs `""`, `"0"`, and `"hello"` yielded `skipped, taken, taken`. The number inputs `0`, `-0`, `2`, `-2`, `NaN`, and `Infinity` yielded `skipped, skipped, taken, taken, skipped, taken`.

The whole latent census Go suite passed on retry with `go test -p 2 -count=1 ./stage3/census/latent/...`. The first run timed out in `TestPublishedCopyModeIsPrivate` during concurrent compiler builds.
