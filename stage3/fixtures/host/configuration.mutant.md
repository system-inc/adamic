Mutant evidence for host fixture configuration selection (#cn4yapb).

This backs the optional `configuration` key and its readers. The real
`status.json` is unchanged. The proof row is
`internal/hostconfiguration/testdata/status.json`; its program is
`configuration.a` beside it.

Mutant: in `internal/hostconfiguration/configuration.go`, `Resolve` treats every
configuration as absent (`if configuration == nil` becomes `if true`). The
reader therefore compares the configured row to stock Node.

Command: `go test ./internal/hostconfiguration -count=1 -run '^TestHostConfigurationReaderUsesDeclaredNode$' -v`

Result with the mutant: exit 1. Restored, the same command passes. The run used
a Go overlay to leave the working tree unchanged. Its real failing lines were:

```text
configuration_test.go:48: configured Node mismatch: <nil>: got "stock\n" want "configured\n"
configuration_test.go:90: configured control: exit status 1
--- FAIL: TestHostConfigurationReaderUsesDeclaredNode (0.08s)
FAIL
```

The same test also deletes the configuration field from a scratch row. That row
runs on stock Node and fails the byte comparison, with the same mismatch. The
configured control uses a stock-Node launcher that labels the proof input; it
proves binary selection, not an inspector implementation or a real
`--without-inspector` build. A separate control names stock Node directly and fails its capability probe.
The launcher explicitly stubs the expected probe line before delegating to Node;
it is mechanism evidence, not a real without-inspector build. An unset variable produces the named skip; an unknown name fails.

`TestHostConfigurationScriptReaders` runs the Python and CJS reader proofs from
Go. Each execution reader runs its real fixture Node command using a selected
probe-verified test stub and checks the recorded bytes, or reports its named
named skip before any configured Node starts. Native and stage0 checks against
recorded bytes continue when the configured binary is unavailable. Generated
JavaScript uses that same configured binary and skips only its own execution. The stock-check reader runs its checker worker under
the selected binary. Adaptation 47's fixture loop is isolated from its unrelated
before/after source proofs. The four glob/manifest readers keep every row in
their census, validate the configuration, and report the named requirement.

The status-diff auditor observes a newly declared profile with its selected
binary, or reports it as skipped. Ordinary records retain the original byte
guard. Same-profile golden changes and admission regressions still fail.


Capability-probe mutant: in `internal/hostconfiguration/configuration.go`, change
`if err != nil || actual != expected` to `if err != nil`. This accepts stock
Node for the configured build. Command:

`go test ./internal/hostconfiguration -count=1 -run '^TestHostConfigurationRejectsStockNode$' -v`

With the Go overlay mutant, exit 1; restored, exit 0. Actual failing lines:

```text
configuration_test.go:192: stock Node must fail its capability probe: <nil>
--- FAIL: TestHostConfigurationRejectsStockNode (0.03s)
FAIL
```

`TestHostConfigurationScriptProbes` independently checks the same mismatch and
one-probe caching (including failures) in both Python and CJS. All resolvers read
`probe` and `expect` from the shared registry. Stock Node prints
`v24.19.0 true`; the registered build must print `v24.19.0 false`.

Recorded-bytes pin: `TestHostConfigurationFixtureReaders/execute/wrong-recorded`
leaves the configured binary unset and plants wrong recorded stdout. The real
native comparison still runs and fails:

```text
native versus Node byte comparison failed: got exit=0 stdout="stock\n" stderr=""; want exit=0 stdout="wrong recorded bytes\n" stderr=""
--- FAIL: TestFixturesHost/host/configuration.a/native
```

The stage0 comparison still runs. This Compiles row has no JavaScript backend
check in the fixture harness: its original scope remains CheckedStop rows only.
`TestHostConfigurationOracleReaders` repeats the same unset-reference and
wrong-recorded-byte proof for both oracle reader paths.

Reference-scope mutant: move the `hostconfiguration.Skip` guard from the `node`
subtest into the row body before `node := entry.Node` in
`stage3/fixtures/fixtures_test.go`. Require itself no longer skips. Command:

`go test ./stage3/fixtures -count=1 -run '^TestHostConfigurationFixtureReaders$' -v`

The Go overlay mutant exits 1. The restored test passes. Actual failing lines:

```text
host_configuration_test.go:133: native comparison didn't run: === RUN   TestFixturesHost
--- FAIL: TestHostConfigurationFixtureReaders/execute/unset (1.32s)
host_configuration_test.go:117: recorded bytes mismatch wasn't caught: <nil>
--- FAIL: TestHostConfigurationFixtureReaders/execute/wrong-recorded (1.24s)
--- FAIL: TestHostConfigurationFixtureReaders (9.46s)
```

The missing native pass and the surviving wrong record independently catch a
row-wide skip. `/native` and `/stage0` continue against the pinned record.
Reference Node and the existing JavaScript checks each require the row's verified
Node binary. Each Go call site owns its skip in its test file; Require never skips.

JavaScript binary-selection mutant: in `internal/oracle/host_status_test.go`,
replace `hostNodeWith(t, inputRun{directory: repository}, nodeBinary, js)` with
`hostNodeWith(t, inputRun{directory: repository}, "node", js)`, retaining the
Require call. Command:

`go test ./internal/oracle -count=1 -run '^TestHostConfigurationOracleReaders$' -v`

The overlay mutant exits 1; the restored reader family passes. Actual lines:

```text
host_configuration_test.go:109: source and generated JavaScript must both use the configured binary: <nil> --disable-warning=ExperimentalWarning
--- FAIL: TestHostConfigurationOracleReaders/status/selected (2.07s)
```

The verified test launcher records every non-probe invocation. The source and
generated JavaScript must both pass through it, independently of whether their
stdout happens to match on stock Node. With no configured binary, native still
compares the recorded bytes and the existing JavaScript execution skips by name.
`test_fs_javascript_uses_configured_node` checks the Python reader's same rule.
The fixture harness additionally asserts that ordinary Compiles rows do not
acquire JavaScript comparisons: only its original CheckedStop path runs them.


Call-site classification pin: each Go caller tests `hostconfiguration.Skip`
in its own `_test.go` file, then calls Require for a verified binary. A missing
binary is the only skip. An unknown name, invalid metadata, or stock binary
returns no skip reason, and Require fails loudly. The manifest logs the missing
configuration and keeps its row.

Error-to-skip mutant: in `hostconfiguration.Skip`, return `err.Error()` instead
of `""` when Resolve fails. Command:

`go test ./internal/hostconfiguration -count=1 -run '^TestHostConfigurationReaderUsesDeclaredNode$' -v`

The overlay exits 1. The restored helper suite passes. Actual failing lines:

```text
configuration_test.go:108: unknown recipe silently accepted: <nil>
--- FAIL: TestHostConfigurationReaderUsesDeclaredNode (0.26s)
```

The helper also checks invalid metadata with no configured binary and stock Node
with valid metadata: both must fail, never skip. The current-head rerun catches
all five source mutants: ignored configuration, accepted stock probe, error as a
skip, row-wide skip, and stock Node for generated JavaScript. The wrong-recorded
bytes and deleted-field controls remain caught by the unmutated reader tests.


Literal-message classification: the eight Go sites use site-unique literal
Skipf messages, with exactly one `%d` for the row. Their Skip conditions are
unchanged. The Go assertions distinguish the fixture's execution skip from the
manifest's log; Python and CJS continue printing their original named messages.
A quick table check validates all eight literals and 13,750 same-package
substring comparisons for rows 1 through 25, with no overlaps. The helpers,
fixture suite and oracle readers pass with these messages; all five source
mutants above still exit 1 at their intended assertions.


Fresh-checkout dependency pin (40848821 follow-up)
------------------------------------------------

This harness also requires setup.sh's pinned TypeScript 6.0.3 API and the
existing TestProduct_FixtureOracleHook. No stage3/api/node_modules installation
is assumed. The script proof passes CENSUS_TYPESCRIPT from the resolved tsc
installation, as adaptation 47's verifier already accepts; stock-check uses the
loader's bundled Node declarations with that explicit API.

In a source-only checkout of 40848821, with an empty environment except runner
and toolchain settings, TestHostConfigurationScriptReaders/node exits 1:

    Error: Cannot find module '/tmp/host-key-2-proof/stage3/api/node_modules/typescript'

The prebuilt fixture proof with empty Go and product caches also exits 1: all
four execute subtests (unset, selected, unknown and wrong-recorded) hit their
30-second child deadline while compiling fixtureOracleHook. The parent now
fetches that declared product before launching bounded child proofs. The
standalone, prebuilt, wholly empty-cache fallback can still exceed the existing
hook recipe's three-minute build bound on this box; the fresh-cache passing
proof prepares products first, as the gate does.

Mutant: remove the Go caller's CENSUS_TYPESCRIPT binding (replace the cmd.Env
assignment with `_ = api`). Run:

    go test ./internal/hostconfiguration -run '^TestHostConfigurationScriptReaders$' -count=1

With no stage3/api/node_modules, this exits 1 at
TestHostConfigurationScriptReaders/node:

    Error: Cannot find module 'typescript'
    --- FAIL: TestHostConfigurationScriptReaders (0.66s)

Restored, the whole helper package passes (6 top-level tests), loading and
checking the real pinned API. The fixture proof uses the same existing product
recipe and preserves its wrong-recorded-bytes negative control. No skip is
added for either missing input.
