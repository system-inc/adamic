# Wave 14 surrogate flag completion

Removed the lone-surrogate flag refusal from no-invalid-regexp.a. The three
third-batch rules remain partial; this change does not claim pattern or
constructor parity. No additional rules are claimed.

Fetched all origin branches before resuming. The branch was clean at
6f236b82a6a9eaab8a85e2f4be0635916f105dd6 and already pushed. Current origin/main
is e011f8f60899586d6373a5ccb07335ad82cfbf3c; the checker branch remains
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6. The full fetched ref snapshot is in
validation-wave-14-surrogates/origin-refs.txt.

## Behavior and evidence

The pinned typescript-go scanner combines adjacent high/low surrogate escapes,
including braced forms, but preserves an unpaired unit as three CESU-8 bytes.
Go's range-over-string and the canonical output decoder read each invalid byte
as one replacement character. The native rule now normalizes each unpaired
UTF-16 unit to three replacement characters before flag validation. Valid pairs
stay intact. This normalization applies to constructor flags only.

The independent unchanged Go production rule supplied the observed behavior.
Three new .a controls cover high and low unpaired units, paired escapes, braced
astral escapes, repeated and reversed units, literal astral and replacement
characters, and duplicate/u-v precedence after an unpaired unit. The former
Go-positive refusal witness is now included in normal byte agreement.
No checker question or shared registration, parser, harness or generator changed.

All 54 controls agree on 60 complete findings and 19,352 bytes, including every
fix and suggestion field. Normal and ASan/UBSan/LeakSanitizer streams are
identical with empty native stderr. The frozen 77 compiler and 287 repository
roots also match in both builds: zero findings, 5,318 and 18,485 bytes respectively.
Source hashes, complete compressed output, stderr, input controls, agreement
hashes and test logs are retained in validation-wave-14-surrogates.

Every comparison mutant builds, exits 0 and produces empty stderr; only the full
independent Go byte comparison catches it:

| Mutant | First differing byte |
| --- | ---: |
| inverted label membership | 51 |
| lost duplicate-flag precedence | 5098 |
| inverted Unicode printable quoting | 5692 |
| one replacement per surrogate instead of three | 17957 |
| inverted combining-mark detector | 6839 |
| scope meaning Value changed to Variable | 1344 |

The released-handle probe panics 70; retaining the handle in the registry exits 0
and fails that expectation. This seventh mutant is an ownership assertion
mutant rather than a finding comparison mutant.

## Commands and output

Toolchain setup succeeded: Go ready 0 s; clang, Node and submodules ready 1 s;
build cache warm 21 s; done 21 s. nproc is 5, with a four-core cgroup quota.
Go 1.27.1, clang 20.1.8 and Node v24.19.0. Shell commands source
/workspace/adamic-tools/env.sh.

```sh
bash cloud/setup.sh > /tmp/wave-14-resume-setup.log 2>&1

ADAMIC_WAVE14_THIRD_ARTIFACTS=/workspace/wave-14-resume/final \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript \
ADAMIC_WAVE14_NEXT_COMPILER_MANIFEST=/workspace/wave-14-artifacts/compiler.manifest \
ADAMIC_WAVE14_NEXT_REPOSITORY_MANIFEST=/workspace/wave-14-artifacts/repository.manifest \
go test ./stage1/cohere/typeaware \
  -run '^TestWave14ThirdAgreementAndMutants$' -count=1 -v -timeout=30m \
  > /tmp/wave-14-resume-final.log 2>&1

go vet ./... > /tmp/wave-14-resume-vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware \
  > /tmp/wave-14-resume-gofmt.log 2>&1
git diff --check > /tmp/wave-14-resume-diffcheck.log 2>&1
```

Rule suite PASS, 99.844 s. Vet, gofmt and whitespace checks pass with empty
logs. The pinned Go native formatter also confirms the changed .a source is
already formatted, using a scratch overlay and virtual .ts filename without
writing a .ts Adamic file or changing the source.

The first run stopped after the first three mutants because the new mutant's Go
search string decoded its Unicode escapes instead of matching the escaped .a
source. Raw Go strings fixed that test construction. The final run exercises
all six comparison mutants and ownership mutant. first.log preserves that failed
attempt; it was not a production-rule discrepancy.

Three quiet alternating rounds measure whole-process --count runs after complete
byte validation. Raw output, timing records, medians and measure.py are retained.
Compilation and sanitizer runs are outside these timings.

| Population | Native median | Go median | Native / Go | Findings |
| --- | ---: | ---: | ---: | ---: |
| compiler | 1.521785 s | 0.320026 s | 4.76x | 0 |
| repository | 0.200480 s | 0.113855 s | 1.76x | 0 |
| controls | 0.020593 s | 0.028872 s | 0.71x | 60 |

## Remaining boundaries

The final test still demonstrates independent Go-positive witnesses and native
exit-70 refusals for:

- undefined labels: the shared native statement parser accepts only Identifier
  tokens before a colon. origin/codex/stage1-estree at 4973271a still has this
  gate. Its parser must change before this case can run end to end.
- string-pattern validation: native ECMAScript rewrite and regexp2-compatible
  syntax validation and error messages are not implemented in this rule slice.
- constructor character classes: native global reference tracking, aliases,
  constant-flow evaluation and cooked-to-raw source mapping are incomplete.

The preceding leaked-number-render rule also still needs real JSX parsing.
origin/codex/lint-harness-dot-a is at f4d98cab; module-loading and shared profile
work does not supply the demonstrated parser or regex dependencies. None of
those shared files was changed or merged into this branch.

The native regex dependencies are substantive missing implementation, not a
shared-harness problem or passing zero-finding comparison. They remain explicit
refusals. Because the claimed rules are not complete, this resume does not select
or claim another batch. The full repository test gate, full option matrices,
new-rule JavaScript comparison and refused paths are not counted as covered.
Earlier full bridge and filtered Node evidence remains in WAVE_14_THIRD_REPORT.md;
no Go checker or bridge code changed in this resume.
