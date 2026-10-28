=== RUN   TestSplitTSGoAgreesUnit00
    units_tsgo_test.go:165: build go_build_./bridge/tsgo/archive_tsgo.a da9bb6e463a9 hit 1.09
    units_tsgo_test.go:165: build split-tsgo-source 3ecee58127c9 hit 0.02
    units_tsgo_test.go:165: build split-tsgo-baseline 9f3bcfb37216 hit 0.02
    units_tsgo_test.go:165: product fetch phase 2.401s
    units_tsgo_test.go:165: whole checker run phase 0.283s
=== RUN   TestSplitTSGoAgreesUnit00/cache_0
native: unit main.c compiles uncached: main.c holds a directive past its first line
    units_tsgo_test.go:158: checker uncached=0: 92 identical bytes: file	/tmp/adamic-gate/TestSplitTSGoAgreesUnit003311150695/001/sample.a
        findings 0 queries 1
--- PASS: TestSplitTSGoAgreesUnit00 (28.94s)
    --- PASS: TestSplitTSGoAgreesUnit00/cache_0 (26.25s)
PASS
ok  	github.com/system-inc/adamic/internal/native	29.057s
