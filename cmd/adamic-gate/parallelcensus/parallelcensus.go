// Package parallelcensus holds cmd/adamic-gate's checks that read every _test.go in the repository: every test runs
// in parallel or says why, every test has one descriptive name, and every duplicated product declaration agrees. They
// live in a package of their own so that what they read keys only this package's tests. Run from cmd/adamic-gate,
// their reads line named every _test.go, and that line moved the keys of cmd/adamic-gate's product tests on any test
// edit anywhere: one comment in internal/gatesample/sample_test.go rebuilt them (#se8mh14). cmd/adamic/treewalk is the
// same shape for cmd/adamic's walk of every Go file (#hbq28nr).
package parallelcensus
