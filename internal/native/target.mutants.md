# Mutants for the native triple (#5xsd2ha, wave 1)

Each mutant is one exact line edit, run on a Linux box against the test it names, then put back from
a saved copy. The tests are in internal/native/target_test.go. Run each with
`go test ./internal/native/ -run '^(TestAProductKeysEachArchApart|TestACrossBuildKeysTheBuildsArch|TestAnEmptyTargetIsTheBuildsTriple)$' -count=1`.

Not yet run: the change was written before Golem's box rule allowed new work, so the Red column is
what each mutant is built to turn red, and it is replaced by what the run printed once it has run.

| Mutant | Edit | Red |
|---|---|---|
| A: Flags drops the triple | native.go, `Flags`: `flags = append(flags, "--target="+options.Target, "-pthread")` made `flags = append(flags, "-pthread")` | TestAProductKeysEachArchApart: the product keyed by native.Flags for linux/amd64 and for linux/arm64 holds one key, since nothing else of its key names the arch. TestACrossBuildKeysTheBuildsArch goes red with it: the cross build's flags hold no `--target` |
| B: the build's platform is the process's | target.go, `buildPlatform`: `goos, goarch = os.Getenv("GOOS"), os.Getenv("GOARCH")` made `goos, goarch = os.Getenv("GOOS"), goruntime.GOARCH` | TestACrossBuildKeysTheBuildsArch: with `GOARCH` set to the arch the host isn't, the flags name the host's triple and the archive's key holds the host's arch, so an arm64 build on an amd64 box keys as amd64. TestAProductKeysEachArchApart and TestAnEmptyTargetIsTheBuildsTriple go red with it |
