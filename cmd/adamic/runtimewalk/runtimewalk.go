// Package runtimewalk holds which packages build the native-runtime archives that runtimeproducts declares, and
// proves it by walking the repository's Go sources. The walk lives in a package of its own so that what it reads,
// every Go file in the tree, keys only this package's test. Run from cmd/adamic, it needed a reads line for every Go
// file, and that line moved the keys of cmd/adamic's runtime product tests on any Go edit anywhere, so about 109 units
// that hold those products reran on changes that never reached them (#hbq28nr).
package runtimewalk

// Builders are the packages whose sources call runtimeproducts.Build or Supply, sorted. cmd/adamic counts each command
// runtime archive's builders from it, and TestBuildersAreWhatTheWalkFinds holds it to the source tree, so a new
// producer is reviewed here.
var Builders = []string{"internal/oracle"}
