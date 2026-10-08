# Certification

The oracle is cohere at `7945d102a6c18dd36adf9114a758ce646e8b2359`.
The descriptor discovers every upstream case and preserves its original filename
and script kind. Owned witnesses run both selected and with all rules.

The additional native mutant test lives in the location rule's
`testdata/certification/`. It mounts a temporary test overlay and calls the
existing harness, including its Go oracle, corpus capture, module copier,
source Node runner, emitted JavaScript runner and sanitized native builder.
It changes no shared test or parser file. Each mutant must differ from Go on
all three runtimes, and all three mutated outputs must agree.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_NEXTJS_RULE=next-no-location-assign-relative-destination python3 stage1/cohere/lint/rules/next-no-location-assign-relative-destination/testdata/certification/run.py > /tmp/next-no-location-assign-relative-destination-certification.log 2>&1
```

The complete merged package also runs with `ADAMIC_TYPESCRIPT_SOURCE` pointing
to TypeScript 6.0.3, `ADAMIC_LINT_BENCH=1`, and both profile artifact and profile
snapshot inputs set. Test output is written to regular log files.

Observed: 99 unique upstream cases and 46 owned witnesses.
The `go-unicode-scheme-fold-omitted` mutant was caught on source Node,
emitted JavaScript and ASan/UBSan native, with identical mutated output.

A final independent probe found Go simple-fold accepting the long s and Kelvin
sign as scheme letters. The port spells those two extra range members explicitly,
because JavaScript legacy regexp ignore-case excludes them. The owned
`testdata/unicode-schemes.ts.txt` witness pins both range positions. The primary
mutant removes those members and must fail against Go on all three runtimes.
The earlier relative-gate inversion also failed on all three runtimes.
