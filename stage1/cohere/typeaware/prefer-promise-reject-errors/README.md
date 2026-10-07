This batch ports the core rules `prefer-promise-reject-errors`,
`prefer-regex-literals`, and `prefer-rest-params` as native Adamic `.a` modules.
Each owned directory exports its rule from `rule.a` and has a standalone `main.a`.
`suite.a` here runs all three with one checker program. Existing shared harness
and registration-generator files are untouched.

The Promise rule reproduces the core rule's optimistic syntax predicate, not the
TypeScript-family rule's error-type test. It follows executor rejection bindings
by identity, including duplicate names, nested closures, defaults, and rest
parameters. `--allow-empty` supports its `allowEmptyReject` option.

The regex rule follows the global constructor and global-object member paths
through aliases, assignments, defaults, and object patterns. Pattern scans,
character spans, flag validation, printable filtering, raw templates, comment
checks, preceding-token checks, replacement text, and suggestions are native.
`--redundant` supports `disallowRedundantWrapping`, including both flag readings.
It reproduces pinned Go cohere, including its latest-edition validity behavior.
The rest rule asks whether the original `arguments` symbol has declarations and
exempts dotted property access.

`binding-origin` is a raw checker question implemented in new files on both
sides. It supplies original and read symbol identities and declaration
file/kind/span/ambient metadata. It does not resolve aliases, track globals,
judge Error values, build edits, or call cohere rules. Native code decides all
three policies. The only shared bridge edit is its switch registration.

The compressed fixture tape contains 662 valid upstream projects and 20 further
controls, with both options preserved. Original upstream Go assertions passed.
One additional upstream project writes `Promise.reject(<string>'x');` as TSX;
the independent Go loader rejects it with parse diagnostics, and it is recorded
as excluded rather than counted as a successful comparison.

Reproduce with the configured toolchain and TypeScript v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/unpack.py /tmp/wave20-third-cases
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py \
  --repository "$PWD" --artifacts /tmp/wave20-third-checks \
  --compiler /path/to/TypeScript --cases /tmp/wave20-third-cases \
  > /tmp/wave20-third-checks.log 2>&1
```

The independent Go oracle calls all three unmodified production rules with the
same options. Comparisons include the complete finding protocol and every
suggestion message and edit span/text. The three production rules offer no
automatic fixes; regex suggestions have real repairs, including two suggestions
on some flag-wrapping cases. Empty fixes are verified, not described as nonempty
fix coverage. Mutants must compile and exit normally; only byte differences
count. The regex mutant moves a suggestion's end by one without changing counts.

Normal and ASan/UBSan/LSan builds, the two frozen corpora, released-handle rejection,
registry-retention and checker-answer mutants, and alternating native/Go timings
are part of the isolated gate. See REPORT.md and validation for the evidence.
Shared `.a` harness integration and emitted-JavaScript rule comparisons remain
outside this unit.
