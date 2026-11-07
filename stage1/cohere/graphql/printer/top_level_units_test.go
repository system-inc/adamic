package printer

import "testing"

// Fixed top-level names are gate units. ADAMIC_TEST_SHARD=i/n remains a local
// selector. Each setup test prepares read-only buildcache products and validates
// the live union. Units prepare products before starting their case timer; every
// caller runs the same recipe, so no unit depends on another having run first.
func TestPrinterUpstreamPreflight_Setup(test *testing.T) {
	test.Parallel()
	printerUpstreamUnit(test, -1)
}
func TestPrinterUpstreamPreflight_000(test *testing.T) { test.Parallel(); printerUpstreamUnit(test, 0) }
func TestPrinterUpstreamPreflight_001(test *testing.T) { test.Parallel(); printerUpstreamUnit(test, 1) }
func TestPrinterUpstreamPreflight_002(test *testing.T) { test.Parallel(); printerUpstreamUnit(test, 2) }
func TestPrinterUpstreamPreflight_003(test *testing.T) { test.Parallel(); printerUpstreamUnit(test, 3) }
