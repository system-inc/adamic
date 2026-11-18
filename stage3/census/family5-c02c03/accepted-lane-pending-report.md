# Family 5 C02/C03: accepted alternating lane evidence

The owner accepts C02's lane as green: both new C02 runs in the alternating sequence passed at 106,366 passing and one sanctioned failure, real lane exit 0. The two earlier reds are accepted as nondeterminism.

[Complete alternating-run report](alternating-runs/REPORT.md), [difference hunks](alternating-runs/artifact-hunks.md), and all evidence are now committed here. The large tarball is excluded; its SHA-256 is in [TARBALL.sha256](alternating-runs/TARBALL.sha256). The original report's no-push statement refers to its measurement turn.

C02 measured candidate: `28a2e22015d0cf51667fcefa7acb246cab553817`. Its README will be amended to record the accepted green runs and earlier nondeterministic reds, retaining one landing commit on main `fbe5b72ba115a8f3060d4b5229fd578574f4505f`.

Next: product test with -timeout 45m, paired full G02 censuses, then the mutant, one proof at a time. No landing is pushed until those gates pass. C03 will be directly stacked on accepted C02 and proved sequentially.
