// Each original bridge observation has one independently selectable root.
package tsgo_test

import "testing"

var bridgeCases = []bridgeCase{
	{"TestBridgeABI", "abi", "", 0, bridgeAbi},
	{"TestBridgeInputLength", "input-length", "", 0, bridgeInputLength},
	{"TestBridgeUnlinkedBuild", "unlinked-build", "", 0, bridgeUnlinkedBuild},
	{"TestBridgeUnlinkedC", "unlinked-c", "", 0, bridgeUnlinkedC},
	{"TestBridgeUnlinkedJavaScript", "unlinked-js", "", 0, bridgeUnlinkedJavaScript},
	{"TestBridgeOutputLength", "output-length", "", 0, bridgeOutputLength},
	{"TestBridgeStaleHandle", "stale-handle", "", 0, bridgeStale},
	{"TestBridgeLinkage", "linkage", "", 0, bridgeLinkage},
	{"TestBridgeOutputFree", "output-free", "", 0, bridgeOutputFree},
	{"TestBridgeRegion", "region", "", 0, bridgeRegion},
	{"TestBridgeRegionOwnership", "region-ownership", "", 0, bridgeRegionOwnership},
	{"TestBridgeOracleSample", "oracle", "sample.ts", 0, bridgeOracle},
	{"TestBridgeTimingRound1Sample", "timing-1", "sample.ts", 1, bridgeTiming},
	{"TestBridgeTimingRound2Sample", "timing-2", "sample.ts", 2, bridgeTiming},
	{"TestBridgeTimingRound3Sample", "timing-3", "sample.ts", 3, bridgeTiming},
	{"TestBridgeWrongPositionSample", "wrong-position", "sample.ts", 0, bridgeWrongPosition},
	{"TestBridgeOracleChecker", "oracle", "checker.ts", 0, bridgeOracle},
	{"TestBridgeTimingRound1Checker", "timing-1", "checker.ts", 1, bridgeTiming},
	{"TestBridgeTimingRound2Checker", "timing-2", "checker.ts", 2, bridgeTiming},
	{"TestBridgeTimingRound3Checker", "timing-3", "checker.ts", 3, bridgeTiming},
	{"TestBridgeWrongPositionChecker", "wrong-position", "checker.ts", 0, bridgeWrongPosition},
	{"TestBridgeOracleParser", "oracle", "parser.ts", 0, bridgeOracle},
	{"TestBridgeTimingRound1Parser", "timing-1", "parser.ts", 1, bridgeTiming},
	{"TestBridgeTimingRound2Parser", "timing-2", "parser.ts", 2, bridgeTiming},
	{"TestBridgeTimingRound3Parser", "timing-3", "parser.ts", 3, bridgeTiming},
	{"TestBridgeWrongPositionParser", "wrong-position", "parser.ts", 0, bridgeWrongPosition},
	{"TestBridgeOracleTypes", "oracle", "types.ts", 0, bridgeOracle},
	{"TestBridgeTimingRound1Types", "timing-1", "types.ts", 1, bridgeTiming},
	{"TestBridgeTimingRound2Types", "timing-2", "types.ts", 2, bridgeTiming},
	{"TestBridgeTimingRound3Types", "timing-3", "types.ts", 3, bridgeTiming},
	{"TestBridgeWrongPositionTypes", "wrong-position", "types.ts", 0, bridgeWrongPosition},
	{"TestBridgeOracleUtilities", "oracle", "utilities.ts", 0, bridgeOracle},
	{"TestBridgeTimingRound1Utilities", "timing-1", "utilities.ts", 1, bridgeTiming},
	{"TestBridgeTimingRound2Utilities", "timing-2", "utilities.ts", 2, bridgeTiming},
	{"TestBridgeTimingRound3Utilities", "timing-3", "utilities.ts", 3, bridgeTiming},
	{"TestBridgeWrongPositionUtilities", "wrong-position", "utilities.ts", 0, bridgeWrongPosition},
}

var bridgeBindings = []func(*testing.T){
	TestBridgeABI,
	TestBridgeInputLength,
	TestBridgeUnlinkedBuild,
	TestBridgeUnlinkedC,
	TestBridgeUnlinkedJavaScript,
	TestBridgeOutputLength,
	TestBridgeStaleHandle,
	TestBridgeLinkage,
	TestBridgeOutputFree,
	TestBridgeRegion,
	TestBridgeRegionOwnership,
	TestBridgeOracleSample,
	TestBridgeTimingRound1Sample,
	TestBridgeTimingRound2Sample,
	TestBridgeTimingRound3Sample,
	TestBridgeWrongPositionSample,
	TestBridgeOracleChecker,
	TestBridgeTimingRound1Checker,
	TestBridgeTimingRound2Checker,
	TestBridgeTimingRound3Checker,
	TestBridgeWrongPositionChecker,
	TestBridgeOracleParser,
	TestBridgeTimingRound1Parser,
	TestBridgeTimingRound2Parser,
	TestBridgeTimingRound3Parser,
	TestBridgeWrongPositionParser,
	TestBridgeOracleTypes,
	TestBridgeTimingRound1Types,
	TestBridgeTimingRound2Types,
	TestBridgeTimingRound3Types,
	TestBridgeWrongPositionTypes,
	TestBridgeOracleUtilities,
	TestBridgeTimingRound1Utilities,
	TestBridgeTimingRound2Utilities,
	TestBridgeTimingRound3Utilities,
	TestBridgeWrongPositionUtilities,
}

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeABI(test *testing.T) { runBridgeCase(test, 0) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeInputLength(test *testing.T) { runBridgeCase(test, 1) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeUnlinkedBuild(test *testing.T) { runBridgeCase(test, 2) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeUnlinkedC(test *testing.T) { runBridgeCase(test, 3) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeUnlinkedJavaScript(test *testing.T) { runBridgeCase(test, 4) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOutputLength(test *testing.T) { runBridgeCase(test, 5) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeStaleHandle(test *testing.T) { runBridgeCase(test, 6) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeLinkage(test *testing.T) { runBridgeCase(test, 7) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOutputFree(test *testing.T) { runBridgeCase(test, 8) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeRegion(test *testing.T) { runBridgeCase(test, 9) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeRegionOwnership(test *testing.T) { runBridgeCase(test, 10) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOracleSample(test *testing.T) { runBridgeCase(test, 11) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound1Sample(test *testing.T) { runBridgeCase(test, 12) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound2Sample(test *testing.T) { runBridgeCase(test, 13) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound3Sample(test *testing.T) { runBridgeCase(test, 14) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeWrongPositionSample(test *testing.T) { runBridgeCase(test, 15) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOracleChecker(test *testing.T) { runBridgeCase(test, 16) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound1Checker(test *testing.T) { runBridgeCase(test, 17) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound2Checker(test *testing.T) { runBridgeCase(test, 18) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound3Checker(test *testing.T) { runBridgeCase(test, 19) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeWrongPositionChecker(test *testing.T) { runBridgeCase(test, 20) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOracleParser(test *testing.T) { runBridgeCase(test, 21) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound1Parser(test *testing.T) { runBridgeCase(test, 22) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound2Parser(test *testing.T) { runBridgeCase(test, 23) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound3Parser(test *testing.T) { runBridgeCase(test, 24) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeWrongPositionParser(test *testing.T) { runBridgeCase(test, 25) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOracleTypes(test *testing.T) { runBridgeCase(test, 26) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound1Types(test *testing.T) { runBridgeCase(test, 27) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound2Types(test *testing.T) { runBridgeCase(test, 28) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound3Types(test *testing.T) { runBridgeCase(test, 29) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeWrongPositionTypes(test *testing.T) { runBridgeCase(test, 30) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOracleUtilities(test *testing.T) { runBridgeCase(test, 31) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound1Utilities(test *testing.T) { runBridgeCase(test, 32) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound2Utilities(test *testing.T) { runBridgeCase(test, 33) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound3Utilities(test *testing.T) { runBridgeCase(test, 34) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeWrongPositionUtilities(test *testing.T) { runBridgeCase(test, 35) }
