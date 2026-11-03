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
configuration_test.go:45: configured Node mismatch: <nil>: got "stock\n" want "configured\n"
configuration_test.go:90: configured control: exit status 1
--- FAIL: TestHostConfigurationReaderUsesDeclaredNode (0.08s)
FAIL
```

The same test also deletes the configuration field from a scratch row. That row
runs on stock Node and fails the byte comparison, with the same mismatch. The
configured control uses a stock-Node launcher that labels the proof input; it
proves binary selection, not an inspector implementation or a real
`--without-inspector` build. A separate control names stock Node directly and
passes. An unset variable produces the named skip; an unknown name fails.

`TestHostConfigurationScriptReaders` runs the Python and CJS reader proofs from
Go. Each execution reader runs its real fixture Node command using a selected
stock-Node path and checks the recorded bytes, or reports its named skip before
any reference Node starts. The stock-check reader runs its checker worker under
the selected binary. Adaptation 47's fixture loop is isolated from its unrelated
before/after source proofs. The four glob/manifest readers keep every row in
their census, validate the configuration, and report the named requirement.

The status-diff auditor observes a newly declared profile with its selected
binary, or reports it as skipped. Ordinary records retain the original byte
guard. Same-profile golden changes and admission regressions still fail.
