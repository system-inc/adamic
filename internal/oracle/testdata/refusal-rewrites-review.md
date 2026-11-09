# Two measured witness rewrites for Lower's review

Base: origin/main b19e8fc0b09cd10d73d2fb9da55673d0103deeaf.
This second head starts directly from main, not the first refusal-pin branch.
No new refusal pins or diagnostic assertions are included. Lower reviews
these sources and measurements before any pin is added.

## Sources and classification

### Computed parameter default

New source: [census_small_default.a](census_small_default.a).
It was moved from census_small_stopped/census_small_default_stopped.a after
lowering and execution agreed with Node.

The parameter is now a plain name, with the same default:

```typescript
selectedMatcher: WildcardMatcher = wildcardMatchers[usage],
```

The body begins with:

```typescript
const { singleAsteriskRegexFragment, doubleAsteriskRegexFragment, replaceWildcardCharacter } = selectedMatcher;
```

The remainder of the body and its five calls are unchanged. The header names
the computed default read as an agreement witness. Its registration is now
lowers: true. The old refusal-only Node subcase was removed; the normal
agreement family checks the promoted fixture. The existing count updater
records the runtime counts required by that new agreement registration.

### Nullable logical value

Source:
[logical_and_reference_null.a](logical_and_reference_refused/logical_and_reference_null.a).

The class and find function are unchanged. The final statements are:

```typescript
const none = find(false);
const value = none && none.ready;
if (value === null) {
    console.log('null');
} else {
    console.log(value ? 'true' : 'false');
}
```

The value retains the boolean | null construct. Printing uses a null test,
with string-valued branches as .a's console.log requires. There is no template
or String conversion. Its header states the measured native disagreement.

Lowering returns a program with no error: there is no NotYet or Refused kind,
reason, or Where to pin. Native execution does not agree with Node, so this
fixture is not promoted to agreement. Its existing lowers: false registration
is left unchanged pending Lower's review. Consequently the generic native
refusal check fails on this measured rewrite because lowering returns nil
error. That failure is disclosed, not hidden by reclassifying it as agreement
or pinning an unrelated refusal.

## Final measurements

The temporary observation test calls the existing lowered helper:
load.Load followed by lower.Lower in process. When lowering succeeds, it runs
onNode, nativelyCheckingLeaks and onJavaScriptBackend, comparing all process
results. Both temporary subtests call t.Parallel. The observation test makes
no diagnostic-reason or location assertions and is removed before commit.

| Fixture | Lowering | Node | Native | JavaScript |
| --- | --- | --- | --- | --- |
| census_small_default.a | Succeeds; no NotYet or Refused | Five lines below, exit 0, empty stderr | Same stdout, exit 0, empty stderr; leak check passes | Same stdout, exit 0, empty stderr |
| logical_and_reference_refused/logical_and_reference_null.a | Succeeds; no NotYet or Refused | null plus newline, exit 0, empty stderr | Empty stdout, exit 70, panic below | null plus newline, exit 0, empty stderr |

Exact native stderr for the logical fixture, including its final newline:

```text
adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back
```

This is a runtime disagreement, not a compile-time refusal. There is no
notYet.Where location for it. No compiler or runtime code was changed.

## Node output

Both originals were extracted from main and run alongside their rewrites using:

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs SOURCE
```

Node v24.19.0 ran all four sources with exit 0 and empty stderr. Each rewritten
stdout was compared byte for byte with its original and the expected output.

Census stdout, original and rewritten:

```text
one/two
missing
one/[^/]*
(?:one(?:/[^./])?)?
one/(?:[^./][^/]*)?
```

Logical stdout, original and rewritten:

```text
null
```

There were no neighbouring expected-output files to update. The outputs are
unchanged; the promoted census source is now checked by ordinary agreement.

## Measurement notes and logs

The first logical printing attempt used String(value). It stopped at a
different conversion refusal, so the final source uses the user's permitted
null-test form instead. The final form exposes the native disagreement above.
No template stop or conversion stop is pinned.

The initial main base was c3b7baf0b3438bf5ac7aa9fcf6eb2ba385e2f9bd.
Main advanced during measurement, so the patch was moved onto b19e8fc0 and the
measurements repeated. The initial observation log is retained separately.

- Final in-process observations: /tmp/refusal-pins-2-measure.log.
- Initial-main observations: /tmp/refusal-pins-2-measure-initial-main.log.
- String-conversion observation: /tmp/refusal-pins-2-measure-string-attempt.log.
- Direct Node output:
  /tmp/refusal-pins-2-{census,logical}-{original,rewritten}-node.{stdout,stderr}.
- Setup: /tmp/refusal-pins-2-setup.log.

All Go checks source /workspace/adamic-tools/env.sh. Setup uses bash
cloud/setup.sh. Every wrapper reports its child's exit code. Logs go to files,
never through head or tail.

## Validation results

| Check | Child exit code | Evidence |
| --- | --- | --- |
| Setup on final main base | 0 | /tmp/refusal-pins-2-setup.log |
| Original and rewritten direct Node runs | 0 each; expected stdout matches, stderr empty | Node output files above |
| Temporary backend observation | 1, solely logical native disagreement; census passes | /tmp/refusal-pins-2-measure.log |
| Scoped census count update | 0 | /tmp/refusal-pins-2-census-counts-update.log |
| Existing census source Node test, family wrapper currency, count file census | 0 | /tmp/refusal-pins-2-source-and-registration-checks.log |
| Native agreement family, census fixture only | 0 | /tmp/refusal-pins-2-native-census.log |
| Generic native refusal family, logical fixture only | 1, because Lower returns nil error | /tmp/refusal-pins-2-native-logical.log |
| go vet ./internal/oracle | 0, empty log | /tmp/refusal-pins-2-vet.log |
| gofmt -l internal/oracle/*.go | 0, empty log | /tmp/refusal-pins-2-gofmt.log |

The logical generic-refusal check's output line is:

```text
oracle_test.go:791: want stage 0 to refuse with where and what, got <nil>
```

The failed temporary observation and failed generic-refusal check are known
results of the measured logical rewrite. This review head is not an all-green
agreement promotion for that fixture. No new pin, skip, harness adjustment,
or compiler fix is used to make those failures disappear.

The census count update uses the existing TestCountsAreRecorded numbered
family, filtered to internal/oracle/testdata/census_small_default.a, with
-count=1 -v -args -update-counts. Native agreement checks likewise filter the
existing TestNativeAgreesWithNode numbered family to each fixture's path.
The temporary observation source is absent from the committed tree.
