package load

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

// LedgerRow supplies identity, never an admission override. Row numbers are
// read from the manifest rather than inferred from checker traversal order.
type LedgerRow struct {
	ID     string `json:"id"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Code   int    `json:"code"`
}

type LedgerDisposition struct {
	ID string `json:"id"`
	OptionDisposition
}

type OptionLedgerReport struct {
	Rows           map[string]LedgerDisposition `json:"rows"`
	Counts         map[string]int               `json:"counts"`
	OrdinaryErrors []string                     `json:"ordinary_errors"`
}

// ReadOptionLedger reads quoted multiline checker messages correctly. Extra
// columns are evidence owned by the ledger and do not affect guard selection.
func ReadOptionLedger(input io.Reader) ([]LedgerRow, error) {
	reader := csv.NewReader(input)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	columns := map[string]int{}
	for i, name := range header {
		columns[name] = i
	}
	for _, name := range []string{"id", "file", "line", "column", "code"} {
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("load: ledger lacks %s", name)
		}
	}
	rows := []LedgerRow{}
	for {
		fields, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		row := LedgerRow{ID: fields[columns["id"]], File: fields[columns["file"]]}
		for _, field := range []struct {
			name  string
			value *int
		}{{"line", &row.Line}, {"column", &row.Column}, {"code", &row.Code}} {
			*field.value, err = strconv.Atoi(strings.TrimPrefix(fields[columns[field.name]], "TS"))
			if err != nil || *field.value <= 0 {
				return nil, fmt.Errorf("load: ledger row %s has invalid %s", row.ID, field.name)
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

type ledgerSiteKey struct {
	file               string
	line, column, code int
}

// LoadOptionLedger makes one whole-program production load and reports every
// manifest row even when remaining errors prevent emission. The report does not
// expose a rejected Program. Its scheduled checks are not --explain-checks counts.
// Missing, additional, or duplicate identities fail closed instead of inventing
// a row or accepting a manifest as permission to weaken the checker.
func LoadOptionLedger(paths []string, sourceRoot string, manifest []LedgerRow) (*Program, *OptionLedgerReport, error) {
	program, loadErr := Load(paths)
	var dispositions []OptionDisposition
	report := &OptionLedgerReport{Rows: map[string]LedgerDisposition{}, Counts: map[string]int{}, OrdinaryErrors: []string{}}
	if program != nil {
		dispositions = program.OptionDispositions()
	} else {
		var rejected *CheckError
		if !errors.As(loadErr, &rejected) {
			return nil, report, loadErr
		}
		dispositions = rejected.OptionDispositions
		report.OrdinaryErrors = append(report.OrdinaryErrors, rejected.OrdinaryDiagnostics...)
	}
	root, err := filepath.Abs(sourceRoot)
	if err != nil {
		return nil, report, err
	}
	keys := map[ledgerSiteKey]string{}
	ids := map[string]bool{}
	for _, row := range manifest {
		path := filepath.Clean(row.File)
		if row.ID == "" || ids[row.ID] || filepath.IsAbs(path) || path == ".." || len(path) > 3 && path[:3] == "../" {
			return nil, report, fmt.Errorf("load: invalid or duplicate ledger identity %q", row.ID)
		}
		ids[row.ID] = true
		key := ledgerSiteKey{filepath.Join(root, path), row.Line, row.Column, row.Code}
		if _, exists := keys[key]; exists {
			return nil, report, fmt.Errorf("load: duplicate ledger site for %s", row.ID)
		}
		keys[key] = row.ID
	}
	for _, disposition := range dispositions {
		site := disposition.Site
		id, ok := keys[ledgerSiteKey{filepath.Clean(site.File), site.Line, site.Column, site.Code}]
		if !ok {
			return nil, report, fmt.Errorf("load: unlisted production site %s:%d:%d TS%d", site.File, site.Line, site.Column, site.Code)
		}
		if _, exists := report.Rows[id]; exists {
			return nil, report, fmt.Errorf("load: duplicate production site for %s", id)
		}
		report.Rows[id] = LedgerDisposition{ID: id, OptionDisposition: disposition}
		report.Counts[disposition.State]++
		if disposition.Kind != "" {
			report.Counts[disposition.State+":"+disposition.Kind]++
		}
	}
	for _, row := range manifest {
		if _, exists := report.Rows[row.ID]; !exists {
			return nil, report, fmt.Errorf("load: missing production site for ledger row %s", row.ID)
		}
	}
	return program, report, loadErr
}
