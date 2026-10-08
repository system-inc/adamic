package load

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptionLedgerContinuesAfterOrdinaryErrors(t *testing.T) {
	t.Parallel()
	paths := projectProgram(t, `{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2024"],"types":[]}`, `const items: number[] = [];
const first: number = items[0];
const point: { x?: number } = { x: undefined };
try {} catch (error) { error.message; }
const serialized: string = JSON.stringify(undefined);
const wrong: number = 'wrong';`)
	manifest := []LedgerRow{
		{ID: "D099", File: "main.ts", Line: 2, Column: 7, Code: 2322},
		{ID: "D017", File: "main.ts", Line: 3, Column: 7, Code: 2375},
		{ID: "D128", File: "main.ts", Line: 4, Column: 24, Code: 18046},
		{ID: "D230", File: "main.ts", Line: 5, Column: 7, Code: 2322},
	}
	program, report, err := LoadOptionLedger(paths, filepath.Dir(paths[0]), manifest)
	var rejected *CheckError
	if program != nil || !errors.As(err, &rejected) {
		t.Fatalf("rejected census exposed a compilable program: %v", err)
	}
	if len(report.Rows) != 4 || report.Counts[OptionCheckScheduled] != 4 || report.Counts[OptionRemainingError] != 0 {
		t.Fatalf("incomplete census: %+v; %v", report, err)
	}
	if report.Rows["D128"].State != OptionCheckScheduled || report.Rows["D128"].Kind != "caught-type" {
		t.Fatalf("catch contract lost: %+v", report.Rows["D128"])
	}
	if len(report.OrdinaryErrors) != 1 || !strings.Contains(report.OrdinaryErrors[0], "Type 'string' is not assignable") {
		t.Fatalf("ordinary error was attributed or suppressed: %+v", report.OrdinaryErrors)
	}
	for _, row := range report.Rows {
		if row.State == "inserted-check" {
			t.Fatal("loader counted a pending contract as an emitted check")
		}
	}
}

func TestOptionLedgerRejectsIdentityDrift(t *testing.T) {
	t.Parallel()
	paths := projectProgram(t, `{"strict":true,"lib":["es2024"],"types":[]}`, `const first: number = [1][0];`)
	row := LedgerRow{ID: "D099", File: "main.ts", Line: 1, Column: 7, Code: 2322}
	for _, test := range []struct {
		name    string
		rows    []LedgerRow
		message string
	}{
		{"missing manifest", nil, "unlisted production site"},
		{"duplicate id", []LedgerRow{row, row}, "duplicate ledger identity"},
		{"missing site", []LedgerRow{{ID: "D099", File: "main.ts", Line: 1, Column: 8, Code: 2322}}, "unlisted production site"},
		{"extra ledger row", []LedgerRow{row, {ID: "D100", File: "main.ts", Line: 2, Column: 7, Code: 2322}}, "missing production site"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := LoadOptionLedger(paths, filepath.Dir(paths[0]), test.rows)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("identity drift accepted: %v", err)
			}
		})
	}
}

func TestReadOptionLedgerMultiline(t *testing.T) {
	t.Parallel()
	rows, err := ReadOptionLedger(strings.NewReader("id,file,line,column,code,message\nD099,main.ts,1,7,TS2322,\"first line\nsecond line\"\n"))
	if err != nil || len(rows) != 1 || rows[0].ID != "D099" || rows[0].Code != 2322 {
		t.Fatalf("ledger identity corrupted: %v %v", rows, err)
	}
}

func TestUnsupportedOptionContractRemainsError(t *testing.T) {
	site := OptionSite{Options: []string{"strictBindCallApply"}, Message: "unsupported stricter option diagnostic"}
	row := (&Program{}).scheduleOptionSite(site)
	if row.State != OptionRemainingError || !strings.Contains(row.Reason, site.Message) {
		t.Fatalf("unsupported option lost its named remaining error: %+v", row)
	}
}
