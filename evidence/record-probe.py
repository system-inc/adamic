"""Add a reporting-only observer to an already generated head latent overlay."""
import json
import sys
from pathlib import Path

source, original, destination = map(lambda name: Path(name).resolve(), sys.argv[1:])
destination.mkdir(parents=True, exist_ok=True)
overlay = json.loads(original.read_text())
replacements = overlay['Replace']

def replace(relative, old, new):
    name = str(source / relative)
    content = Path(replacements.get(name, name)).read_text()
    assert content.count(old) == 1, (relative, old)
    output = destination / relative.replace('/', '_')
    output.write_text(content.replace(old, new))
    replacements[name] = str(output)

replace('internal/lower/records_named.go',
        'l.result.ViewOrigins = append(l.result.ViewOrigins, array)',
        '''l.result.ViewOrigins = append(l.result.ViewOrigins, array)
        if sites, ok := censusRecordReads[l]; ok {
            censusRecordReads[l] = append(sites, l.program.Where(node))
        }''')
replace('internal/lower/records_named.go',
        'if !l.isArrayType(declared) {\n\t\treturn nil\n\t}\n\treturn declared',
        'if !l.isArrayType(declared) {\n\t\treturn nil\n\t}\n censusRecordCandidate(l.program.Where(n))\n\treturn declared')
for file, line in [
    ('latent_hook.go', 'l := fresh()'),
    ('latent_units.go', 'l := base.latentSnapshot()'),
]:
    replace(f'internal/lower/{file}', line, line + '\n defer l.censusRecordProbe()()')
helper = destination / 'record_probe.go'
helper.write_text(r'''package lower

import (
    "encoding/json"
    "os"
)

var censusRecordCandidates = map[string]bool{}

func censusRecordCandidate(site string) {
    if !censusRecordCandidates[site] {
        censusRecordCandidates[site] = true
        json.NewEncoder(os.Stderr).Encode(map[string]string{"slot_candidate": site})
    }
}

var censusRecordReads = map[*lowering][]string{}

func (l *lowering) censusRecordProbe() func() {
    censusRecordReads[l] = nil
    unit := latentFindingOwner
    return func() {
        sites := censusRecordReads[l]
        delete(censusRecordReads, l)
        if len(sites) == 0 {
            return
        }
        row := map[string]any{"unit": unit, "slot_reads": sites}
        defer func() {
            if failure := recover(); failure != nil {
                row["probe_panic"] = failure
            }
            json.NewEncoder(os.Stderr).Encode(row)
        }()
        // Number allocation sites on a disposable IR copy. The fixture's production
        // query uses the same numbering and graph, without changing the latent run.
        probe := l.latentSnapshot()
        assignAllocationSites(probe.result)
        graph := newAllocationFlowGraph(probe.result)
        unknown, why := graph.recordValuesUnknown()
        row["unknown"] = unknown
        row["write_or_reason"] = why
        row["record_writes"] = len(graph.recordWrites)
        row["view_origins"] = len(probe.result.ViewOrigins)
        reached, allUnknown := graph.viewedFrom(probe.result.ViewOrigins, true)
        row["viewed_sites"] = len(reached)
        row["viewed_unknown"] = allUnknown
    }
}
''')
replacements[str(source / 'internal/lower/record_probe.go')] = str(helper)
(destination / 'overlay.json').write_text(json.dumps(overlay))
