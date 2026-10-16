// Package compilercensus holds the census of C compiler calls, TestEveryCCompilerRunsThroughNative, in a package of its
// own. It reads every Go file in the repository, so Loom keys its unit on all of them, and a change to any one runs it
// again. In internal/native it was keyed on native's own reads, so a change that added a compiler call elsewhere reused
// native's earlier green and put the census red on main (#fwms5s4). Here a rerun is about one second, not native's
// whole package.
package compilercensus
