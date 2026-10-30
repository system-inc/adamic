// Package treewalk holds cmd/adamic's checks that read every Go file in the repository. They live in a package of
// their own so that what they read keys only this package's tests. Run from cmd/adamic, the runtime census needed a
// reads line for every Go file, and that line moved the keys of cmd/adamic's runtime product tests on any Go edit
// anywhere, so about 109 units that hold those products reran on changes that never reached them (#hbq28nr). The Go
// half of docs/cli.md's citations is read here for the same reason: the page cites Go outside cmd/adamic's closure,
// and a reads line for each such file would be a promise nothing keeps.
package treewalk

import "regexp"

// RuntimeBuilders are the packages whose sources call runtimeproducts.Build or Supply, sorted. cmd/adamic counts each
// command runtime archive's builders from it, and TestRuntimeBuildersAreWhatTheWalkFinds holds it to the source tree,
// so a new producer is reviewed here.
var RuntimeBuilders = []string{"internal/oracle"}

// Citation is a backticked repository path in docs/cli.md, with an optional anchor after a colon: a symbol list for
// Go, or lines for any other file. cmd/adamic checks the citations of other files, and this package's test the Go ones.
var Citation = regexp.MustCompile("`((?:[A-Za-z0-9_.-]+/)+[A-Za-z0-9_.-]+\\.[a-z]+)(?::([A-Za-z0-9_.,-]+))?`")
