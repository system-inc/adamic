package scanall

import "testing"

func TestScanAllCensusBindStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "bind-stop", true, false, false)
}
func TestScanAllCensusBindClean(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "bind-clean", false, false, true)
}
func TestScanAllCensusRecordStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "record-stop", false, true, false)
}
func TestScanAllCensusRecordClean(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "record-clean", false, false, true)
}
func TestScanAllCensusCombinedBindStopRecordStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "combined-bind-stop-record-stop", false, true, false)
}
func TestScanAllCensusCombinedBindStopRecordClean(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "combined-bind-stop-record-clean", true, false, false)
}
func TestScanAllCensusCombinedBindCleanRecordStop(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "combined-bind-clean-record-stop", false, true, false)
}
func TestScanAllCensusCombinedBindCleanRecordClean(t *testing.T) {
	t.Parallel()
	scanAllControl(t, "combined-bind-clean-record-clean", false, false, true)
}
