=== RUN   TestSplitTSGoAgreesUnit00
    units_tsgo_test.go:165: build go_build_./bridge/tsgo/archive_tsgo.a da9bb6e463a9 hit 0.48
    units_tsgo_test.go:165: build split-tsgo-source 3ecee58127c9 hit 0.01
    units_tsgo_test.go:165: build split-tsgo-baseline 9f3bcfb37216 hit 0.01
    units_tsgo_test.go:165: product fetch phase 0.906s
    units_tsgo_test.go:165: whole checker run phase 0.088s
=== RUN   TestSplitTSGoAgreesUnit00/cache_0
native: unit main.c compiles uncached: main.c holds a directive past its first line
    units_tsgo_test.go:158: checker uncached=0: 92 identical bytes: file	/tmp/adamic-gate/TestSplitTSGoAgreesUnit001540431116/001/sample.a
        findings 0 queries 1
--- PASS: TestSplitTSGoAgreesUnit00 (3.59s)
    --- PASS: TestSplitTSGoAgreesUnit00/cache_0 (2.59s)
PASS
ok  	github.com/system-inc/adamic/internal/native	3.616s
