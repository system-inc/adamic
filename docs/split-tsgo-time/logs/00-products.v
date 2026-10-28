=== RUN   TestProduct_TSGoArchive
=== PAUSE TestProduct_TSGoArchive
=== RUN   TestProduct_SplitTSGoSource
=== PAUSE TestProduct_SplitTSGoSource
=== RUN   TestProduct_SplitTSGoBaseline
=== PAUSE TestProduct_SplitTSGoBaseline
=== CONT  TestProduct_TSGoArchive
=== CONT  TestProduct_SplitTSGoBaseline
=== CONT  TestProduct_SplitTSGoSource
    units_tsgo_test.go:104: build split-tsgo-source 3ecee58127c9 hit 3.95
=== NAME  TestProduct_SplitTSGoBaseline
    units_tsgo_test.go:111: build split-tsgo-source 3ecee58127c9 miss 3.95
--- PASS: TestProduct_SplitTSGoSource (4.22s)
=== NAME  TestProduct_TSGoArchive
    units_tsgo_test.go:31: build go_build_./bridge/tsgo/archive_tsgo.a da9bb6e463a9 miss 121.43
--- PASS: TestProduct_TSGoArchive (121.70s)
=== NAME  TestProduct_SplitTSGoBaseline
    units_tsgo_test.go:111: build go_build_./bridge/tsgo/archive_tsgo.a da9bb6e463a9 hit 117.48
    units_tsgo_test.go:111: build split-tsgo-baseline 9f3bcfb37216 miss 26.72
--- PASS: TestProduct_SplitTSGoBaseline (148.50s)
PASS
ok  	github.com/system-inc/adamic/internal/native	148.524s
