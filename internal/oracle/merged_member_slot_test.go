package oracle

import (
	"path/filepath"
	"testing"
)

func TestMergedMemberSlot(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct {
		name, stdout, stop string
	}{
		{name: "write", stdout: "m/string a,f\n"},
		{name: "optional", stdout: "undefined a\n"},
		{
			name: "required", stdout: "undefined\n",
			stop: "adamic: panic: field read failed: m.g (class Merged, interface merge) is not initialized; expected string, found missing\n",
		},
		{
			name:   "own_keys",
			stdout: "false {\"a\":1} a undefined\ntrue {\"a\":1,\"f\":\"m\"} a,f m\n",
		},
	} {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/merged_member_slot", pin.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			want := run{stdout: []byte(pin.stdout)}
			if diff := disagreement(want, node); diff != "" {
				t.Fatalf("Node: %s, got %s", diff, checkedNamesShown(node))
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			// This ruled stop diverges from Node, which reads undefined.
			if pin.stop != "" {
				want = run{exitCode: 70, stderr: []byte(pin.stop)}
			}
			native, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{
				{"native", native},
				{"JavaScript", onJavaScriptBackend(t, program)},
			} {
				if diff := disagreement(want, backend.got); diff != "" {
					t.Errorf("%s: %s, got %s", backend.name, diff, checkedNamesShown(backend.got))
				} else {
					t.Logf("%s: %s", backend.name, checkedNamesShown(backend.got))
				}
			}
		})
	}
}
