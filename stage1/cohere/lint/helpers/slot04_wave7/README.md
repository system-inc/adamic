# Slot 04 seventh batch

Three separate helpers: has_variant.a, variant_kind.a, variant_compounds_with.a. variant_view.a defines only their read-only state and ignored-child views. main.a is the differential entry. This package and claims/04.md are the only modified territory.

VariantSystemView projects the actual registry onto its key-to-kind map. HasVariant reads presence, VariantKind returns the stored kind (including empty or unknown kinds) or static for an absent key, and VariantCompoundsWith tests the registered parent kind against compound. Its child argument is deliberately ignored, matching the actual Go implementation's documented approximation. No Tailwind selector or compound bitmask implementation is implied. The methods observe live registry mutations; the read-only view prevents callers from mutating through these helpers.

The differential adapter uses an injective comma-separated decimal-byte encoding for registry roots and queries. This preserves arbitrary Go string equality, including malformed UTF-8. Kind strings are decoded valid UTF-8, including the four actual registration kinds, empty and unknown control strings. The registry's order and other fields do not affect these methods and are omitted from the view; allocation, registration and broader design-system loading remain separate dependencies.

After sourcing /workspace/adamic-tools/env.sh:

```sh
python3 stage1/cohere/lint/helpers/slot04_wave7/testdata/capture.py > /tmp/slot04-wave7-capture.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave7/testdata/readiness.py > /tmp/slot04-wave7-readiness.log 2>&1
go test ./stage1/cohere/lint/helpers/slot04_wave7 -count=1 -v > /tmp/slot04-wave7-tests.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot04_wave7 > /tmp/slot04-wave7-vet.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot04_wave7/helpers_test.go stage1/cohere/lint/helpers/slot04_wave7/testdata/oracle.go stage1/cohere/lint/helpers/slot04_wave7/testdata/exports.go > /tmp/slot04-wave7-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/slot04-wave7-oracle.log 2>&1
```

Temporary Go overlays invoke the actual pinned LoadedDesignSystem methods and capture complete registry kind/query states for every call. No shared source is edited. Capture installs tailwindcss@4.3.3 in a temporary directory and substitutes two existing developer-path fixtures. The coverage assertion derives consumers from the frozen readiness ledger. Exact output from source Node, sanitized native and emitted JavaScript is compared with actual Go, and each semantic mutant must compile and execute before being caught. See REPORT.md and evidence/ for scope and measurements.
