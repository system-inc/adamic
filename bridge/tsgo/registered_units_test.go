// Each original bridge observation has one independently selectable root.
package tsgo_test

import "testing"

var bridgeCases = []bridgeCase{
	{"TestBridgeABI", "abi", "", 0, bridgeABI},
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
func TestBridgeABI(t *testing.T) { runBridgeCase(t, 0) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeInputLength(t *testing.T) { runBridgeCase(t, 1) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeUnlinkedBuild(t *testing.T) { runBridgeCase(t, 2) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeUnlinkedC(t *testing.T) { runBridgeCase(t, 3) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeUnlinkedJavaScript(t *testing.T) { runBridgeCase(t, 4) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOutputLength(t *testing.T) { runBridgeCase(t, 5) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeStaleHandle(t *testing.T) { runBridgeCase(t, 6) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeLinkage(t *testing.T) { runBridgeCase(t, 7) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOutputFree(t *testing.T) { runBridgeCase(t, 8) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeRegion(t *testing.T) { runBridgeCase(t, 9) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeRegionOwnership(t *testing.T) { runBridgeCase(t, 10) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOracleSample(t *testing.T) { runBridgeCase(t, 11) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound1Sample(t *testing.T) { runBridgeCase(t, 12) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound2Sample(t *testing.T) { runBridgeCase(t, 13) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound3Sample(t *testing.T) { runBridgeCase(t, 14) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeWrongPositionSample(t *testing.T) { runBridgeCase(t, 15) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOracleChecker(t *testing.T) { runBridgeCase(t, 16) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound1Checker(t *testing.T) { runBridgeCase(t, 17) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound2Checker(t *testing.T) { runBridgeCase(t, 18) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound3Checker(t *testing.T) { runBridgeCase(t, 19) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeWrongPositionChecker(t *testing.T) { runBridgeCase(t, 20) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOracleParser(t *testing.T) { runBridgeCase(t, 21) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound1Parser(t *testing.T) { runBridgeCase(t, 22) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound2Parser(t *testing.T) { runBridgeCase(t, 23) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound3Parser(t *testing.T) { runBridgeCase(t, 24) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeWrongPositionParser(t *testing.T) { runBridgeCase(t, 25) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOracleTypes(t *testing.T) { runBridgeCase(t, 26) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound1Types(t *testing.T) { runBridgeCase(t, 27) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound2Types(t *testing.T) { runBridgeCase(t, 28) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound3Types(t *testing.T) { runBridgeCase(t, 29) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeWrongPositionTypes(t *testing.T) { runBridgeCase(t, 30) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeOracleUtilities(t *testing.T) { runBridgeCase(t, 31) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound1Utilities(t *testing.T) { runBridgeCase(t, 32) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound2Utilities(t *testing.T) { runBridgeCase(t, 33) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeTimingRound3Utilities(t *testing.T) { runBridgeCase(t, 34) }

// Not parallel: sanitizer subprocesses and baseline timings share the instance.
func TestBridgeWrongPositionUtilities(t *testing.T) { runBridgeCase(t, 35) }
