package printer

import "testing"

// Fixed top-level names are gate units. ADAMIC_TEST_SHARD=i/n remains a local
// selector. Each setup test prepares read-only buildcache products and validates
// the live union. Units prepare private paths before starting their case timer;
// no sync.Once or shared mutable state makes the first unit build for the rest.
func TestPrinterAsGoCohere_Setup(t *testing.T)        { t.Parallel(); printerAsGoUnit(t, -1) }
func TestPrinterAsGoCohere_000(t *testing.T)          { t.Parallel(); printerAsGoUnit(t, 0) }
func TestPrinterAsGoCohere_001(t *testing.T)          { t.Parallel(); printerAsGoUnit(t, 1) }
func TestPrinterAsGoCohere_002(t *testing.T)          { t.Parallel(); printerAsGoUnit(t, 2) }
func TestPrinterAsGoCohere_003(t *testing.T)          { t.Parallel(); printerAsGoUnit(t, 3) }
func TestPrinterUpstreamPreflight_Setup(t *testing.T) { t.Parallel(); printerUpstreamUnit(t, -1) }
func TestPrinterUpstreamPreflight_000(t *testing.T)   { t.Parallel(); printerUpstreamUnit(t, 0) }
func TestPrinterUpstreamPreflight_001(t *testing.T)   { t.Parallel(); printerUpstreamUnit(t, 1) }
func TestPrinterUpstreamPreflight_002(t *testing.T)   { t.Parallel(); printerUpstreamUnit(t, 2) }
func TestPrinterUpstreamPreflight_003(t *testing.T)   { t.Parallel(); printerUpstreamUnit(t, 3) }
func TestPrinterWhitespaceGap_Setup(t *testing.T)     { t.Parallel(); printerWhitespaceUnit(t, -1) }
func TestPrinterWhitespaceGap_000(t *testing.T)       { t.Parallel(); printerWhitespaceUnit(t, 0) }
func TestPrinterWhitespaceGap_001(t *testing.T)       { t.Parallel(); printerWhitespaceUnit(t, 1) }
func TestPrinterWhitespaceGap_002(t *testing.T)       { t.Parallel(); printerWhitespaceUnit(t, 2) }
func TestPrinterWhitespaceGap_003(t *testing.T)       { t.Parallel(); printerWhitespaceUnit(t, 3) }
func TestPrinterWhitespaceGap_004(t *testing.T)       { t.Parallel(); printerWhitespaceUnit(t, 4) }
