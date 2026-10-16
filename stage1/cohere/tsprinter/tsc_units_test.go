package tsprinter

import "testing"

func tscOracle(t *testing.T, root string) string {
	t.Helper()
	return printerGateGoOracle(t) + "/oracle"
}

type tsPrinterProducts struct{ source, backend, sanitized, release string }
