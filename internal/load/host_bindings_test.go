package load

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// The host-binding table is an axiom under Language's audit (#00qn5az): docs/axioms/host-bindings.md's
// Rows section and host_bindings.json say the same rows, cell for cell, both ways, and every row names
// the ruling or task that made it. Mutant (b), internal/oracle/testdata/host_bindings/mutants.md: a row's
// ruling dropped from the JSON goes red here.
func TestHostBindingsAudit(test *testing.T) {
	test.Parallel()
	rows, err := HostBindings()
	if err != nil {
		test.Fatal(err)
	}
	audited, err := hostBindingsAudited("../../docs/axioms/host-bindings.md")
	if err != nil {
		test.Fatal(err)
	}
	table := map[string]bool{}
	for _, row := range rows {
		table[row.Name] = true
		cells, found := audited[row.Name]
		if !found {
			test.Errorf("%s: a row in host_bindings.json the audit doesn't list", row.Name)
			continue
		}
		want := hostBindingCells(row)
		for index, heading := range hostBindingHeadings {
			if cells[index] != want[index] {
				test.Errorf("%s: the audit's %s is %q, the table's %q", row.Name, heading, cells[index], want[index])
			}
		}
	}
	for name, cells := range audited {
		if !table[name] {
			test.Errorf("%s: a row the audit lists that host_bindings.json doesn't hold", name)
		}
		if strings.TrimSpace(cells[len(cells)-1]) == "" {
			test.Errorf("%s: the audit's row names no ruling", name)
		}
	}
}

// hostBindingHeadings are the audit table's columns, in order.
var hostBindingHeadings = []string{
	"Name",
	"Declaration",
	"Node",
	"native",
	"javascript",
	"wasm32",
	"Model",
	"Divergence",
	"Ruling",
}

// hostBindingCells is a row as the audit writes it, unescaped.
func hostBindingCells(row HostBinding) []string {
	column := func(target HostBindingColumn) string {
		if target.Answer == HostBindingPending {
			return fmt.Sprintf("Pending (%s)", target.Owner)
		}
		return string(target.Answer)
	}
	model := "none"
	switch {
	case row.Model.Path != "":
		model = "`" + row.Model.Path + "`"
	case row.Model.Task != "":
		model = row.Model.Task + ", not built"
	}
	divergence := "none"
	if row.Divergence != "" {
		divergence = row.Divergence
	}
	return []string{
		row.Name,
		fmt.Sprintf(
			"%s, `declare %s %s: %s`",
			row.Declaration.File,
			row.Declaration.Keyword,
			row.Name,
			row.Declaration.Type,
		),
		fmt.Sprintf("%s, %s module, `%s`", row.Node.Version, row.Node.Module, row.Node.TypeOf),
		column(row.Native), column(row.JavaScript), column(row.WASM32),
		model, divergence, row.Ruling,
	}
}

// hostBindingsAudited reads the audit's Rows table: each row's cells by its name, with `\|` unescaped.
func hostBindingsAudited(file string) (map[string][]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	rows := map[string][]string{}
	section, found, heading := false, false, false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "## ") {
			section = strings.TrimSpace(line) == "## Rows"
			found = found || section
			continue
		}
		if !section || !strings.HasPrefix(line, "|") {
			continue
		}
		var cells []string
		for _, cell := range strings.Split(strings.ReplaceAll(strings.Trim(line, "|"), `\|`, "\x00"), "|") {
			cells = append(cells, strings.ReplaceAll(strings.TrimSpace(cell), "\x00", "|"))
		}
		if !heading {
			if strings.Join(cells, ",") != strings.Join(hostBindingHeadings, ",") {
				return nil, fmt.Errorf(
					"%s: the Rows table's headings are %q, want %q",
					file,
					cells,
					hostBindingHeadings,
				)
			}
			heading = true
			continue
		}
		if strings.Trim(strings.Join(cells, ""), "- ") == "" {
			continue
		}
		if len(cells) != len(hostBindingHeadings) {
			return nil, fmt.Errorf(
				"%s: a row of %d cells, want %d: %s",
				file,
				len(cells),
				len(hostBindingHeadings),
				line,
			)
		}
		if _, twice := rows[cells[0]]; twice {
			return nil, fmt.Errorf("%s: %s is listed twice", file, cells[0])
		}
		rows[cells[0]] = cells
	}
	if !found || !heading {
		return nil, fmt.Errorf("%s has no Rows table", file)
	}
	return rows, nil
}

// A row the table can't trust is no row: the parser refuses it, naming why, and every reader of the
// table then fails closed (ambientRefusal reports the error at the first ambient value declaration).
func TestHostBindingsRefuseAnUntrustedRow(test *testing.T) {
	test.Parallel()
	const row = `{"name": "onProfilerEvent", "declaration": {"file": "performance.ts", "keyword": "let", "type": "number"}, "node": {"version": "v24.21.0", "module": "ES", "typeof": "undefined"}, "native": {"answer": "Absent"}, "javascript": {"answer": "Absent"}, "wasm32": {"answer": "Absent"}, "model": {"path": "", "task": ""}, "divergence": "", "ruling": "#00qn5az"}`
	if _, err := parseHostBindings([]byte("[" + row + "]")); err != nil {
		test.Fatalf("the control row: %v", err)
	}
	for _, probe := range []struct{ name, from, to, want string }{
		{"no ruling", `"ruling": "#00qn5az"`, `"ruling": " "`, "no ruling or task names what made the row"},
		{
			"a pending column with no owner",
			`"native": {"answer": "Absent"}`,
			`"native": {"answer": "Pending"}`,
			"the native column is pending with no owner to clear it",
		},
		{
			"an owner on a cleared column",
			`"wasm32": {"answer": "Absent"}`,
			`"wasm32": {"answer": "Absent", "owner": "Platforms"}`,
			"the wasm32 column names an owner but isn't pending",
		},
		{
			"an answer the table doesn't know",
			`"javascript": {"answer": "Absent"}`,
			`"javascript": {"answer": "Maybe"}`,
			"the javascript column answers neither Provided, Absent nor Pending",
		},
		{
			"a Provided column with no model",
			`"javascript": {"answer": "Absent"}`,
			`"javascript": {"answer": "Provided"}`,
			"a model is named exactly when a column is Provided",
		},
		{
			"a model with no Provided column",
			`"path": ""`,
			`"path": "node:process"`,
			"a model is named exactly when a column is Provided",
		},
		{
			"a CommonJS answer",
			`"module": "ES"`,
			`"module": "CommonJS"`,
			"Node's answer needs its version, module mode ES and a typeof answer",
		},
		{"a var", `"keyword": "let"`, `"keyword": "var"`, "the declaration's keyword is neither const nor let"},
		{
			"a path",
			`"file": "performance.ts"`,
			`"file": "compiler/performance.ts"`,
			"the declaration's file isn't the basename of a .ts file",
		},
		{"an unknown field", `"divergence": ""`, `"divergence": "", "trusted": true`, `unknown field "trusted"`},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			changed := strings.Replace(row, probe.from, probe.to, 1)
			if changed == row {
				test.Fatalf("%q isn't in the control row", probe.from)
			}
			if _, err := parseHostBindings(
				[]byte("[" + changed + "]"),
			); err == nil || !strings.Contains(err.Error(), probe.want) {
				test.Errorf("got %v, want an error containing %q", err, probe.want)
			}
		})
	}
	if _, err := parseHostBindings(
		[]byte("[" + row + ", " + row + "]"),
	); err == nil || !strings.Contains(err.Error(), "a second row for the name") {
		test.Errorf("two rows for one name: got %v", err)
	}
}
