=== RUN   TestSplitTSGoAgreesUnit01
    units_tsgo_test.go:170: build go_build_./bridge/tsgo/archive_tsgo.a da9bb6e463a9 hit 0.86
    units_tsgo_test.go:170: build split-tsgo-source 3ecee58127c9 hit 0.03
    units_tsgo_test.go:170: build split-tsgo-baseline 9f3bcfb37216 hit 0.01
    units_tsgo_test.go:170: product fetch phase 2.107s
    units_tsgo_test.go:170: whole checker run phase 0.200s
=== RUN   TestSplitTSGoAgreesUnit01/cache_1
    units_tsgo_test.go:158: checker uncached=1: 92 identical bytes: file	/tmp/adamic-gate/TestSplitTSGoAgreesUnit013095320154/001/sample.a
        findings 0 queries 1
--- PASS: TestSplitTSGoAgreesUnit01 (27.01s)
    --- PASS: TestSplitTSGoAgreesUnit01/cache_1 (24.70s)
PASS
ok  	github.com/system-inc/adamic/internal/native	27.052s
