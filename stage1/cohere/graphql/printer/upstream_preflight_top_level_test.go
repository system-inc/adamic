package printer

import "testing"

// Fixed top-level names are gate units. ADAMIC_TEST_SHARD=i/n remains a local
// selector. Each setup test prepares read-only buildcache products and validates
// the live union. Units prepare products before starting their case timer; every
// caller runs the same recipe, so no unit depends on another having run first.
func TestPrinterUpstreamPreflight_Setup(t *testing.T) { t.Parallel(); printerUpstreamUnit(t, -1) }
func TestPrinterUpstreamPreflight_000(t *testing.T)   { t.Parallel(); printerUpstreamUnit(t, 0) }
func TestPrinterUpstreamPreflight_001(t *testing.T)   { t.Parallel(); printerUpstreamUnit(t, 1) }
func TestPrinterUpstreamPreflight_002(t *testing.T)   { t.Parallel(); printerUpstreamUnit(t, 2) }
func TestPrinterUpstreamPreflight_003(t *testing.T)   { t.Parallel(); printerUpstreamUnit(t, 3) }
