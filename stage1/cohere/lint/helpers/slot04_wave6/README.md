# Slot 04 sixth batch

The three helpers live separately in theme_prefix_key.a, variant_registry_has.a and design_system_prefix.a. theme_view.a supplies only the two state-view types; main.a is the differential entry. This package and claims/04.md are the only modified territory.

ThemeView.prefix and PrefixKey's key are immutable raw byte strings, represented by readonly arrays of integer bytes 0..255. Callers must keep backing arrays immutable. ThemeView itself permits prefix replacement, and Prefix reads the current field through DesignSystemView.theme on every call. PrefixKey returns the original immutable key when the prefix is empty; otherwise it concatenates two hyphens, the prefix, a hyphen and the bytes after key offset 2. It does not validate leading hyphens, normalize text or decode UTF-8. A nonempty prefix with a key shorter than two bytes fails explicitly, matching Go's panicking slice.

Has reads presence from a ReadonlyMap<string,boolean> registration view. Marker values are opaque and may be false: presence never depends on their value. The full registration's kind/order fields are not dependencies of this operation. In the arbitrary-byte differential adapter, keys use the injective comma-separated decimal-byte encoding. It preserves Go string equality without silently replacing malformed UTF-8. The helper itself requires no encoding convention, only identical string identity for registry keys and queries.

After sourcing /workspace/adamic-tools/env.sh:

```sh
python3 stage1/cohere/lint/helpers/slot04_wave6/testdata/capture.py > /tmp/slot04-wave6-capture.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave6/testdata/readiness.py > /tmp/slot04-wave6-readiness.log 2>&1
go test ./stage1/cohere/lint/helpers/slot04_wave6 -count=1 -v > /tmp/slot04-wave6-tests-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot04_wave6 > /tmp/slot04-wave6-vet.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot04_wave6/helpers_test.go stage1/cohere/lint/helpers/slot04_wave6/testdata/oracle.go stage1/cohere/lint/helpers/slot04_wave6/testdata/exports.go > /tmp/slot04-wave6-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/slot04-wave6-oracle.log 2>&1
```

Go overlays invoke the actual pinned Cohere methods and capture every prefix-key pair and complete registry key/query state. No shared source is edited. The capture tool installs tailwindcss@4.3.3 in a temporary directory and substitutes the existing two developer-path fixtures. Source Node, sanitized native and emitted JavaScript compare exact output with Go; short-key failures are checked independently rather than counted as matching fabricated values. See REPORT.md and evidence/ for counts and limits.
