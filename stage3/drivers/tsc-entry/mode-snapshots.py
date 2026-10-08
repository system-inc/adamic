#!/usr/bin/env python3
"""Probe stricter-mode admission in each saved site's actual owning TS project."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('checkout', type=Path)
parser.add_argument('tree', type=Path)
parser.add_argument('output', type=Path)
args = parser.parse_args()
here = Path(__file__).resolve().parent
baseline = here / 'evidence/main-efe9f404'
records = json.loads((baseline / 'stops.json').read_text())
checkout = args.checkout.resolve()
tree = args.tree.resolve()
out = args.output.resolve()
out.mkdir()
# This tool must be compiled against the scratch mode, not publication main.
probe_source = '''package main
import (
 "encoding/json"
 "errors"
 "os"
 "github.com/system-inc/adamic/internal/load"
)
func main() {
 program, err := load.Load(os.Args[1:])
 report := map[string]any{"loaded": program != nil, "error": "", "ordinary_diagnostics": []string{}, "dispositions": []load.OptionDisposition{}}
 if err != nil { report["error"] = err.Error() }
 if program != nil { report["dispositions"] = program.OptionDispositions() }
 var rejected *load.CheckError
 if errors.As(err, &rejected) {
  report["ordinary_diagnostics"] = rejected.OrdinaryDiagnostics
  report["dispositions"] = rejected.OptionDispositions
 }
 if err := json.NewEncoder(os.Stdout).Encode(report); err != nil { panic(err) }
}
'''
probe_directory = checkout / 'stage3/tsc-entry-mode-probe'
probe_directory.mkdir()
(probe_directory / 'main.go').write_text(probe_source)
with (out / 'probe-build.log').open('wb') as log:
    subprocess.run(['go', 'build', '-o', str(out / 'probe'), './stage3/tsc-entry-mode-probe'], cwd=checkout, stdout=log, stderr=subprocess.STDOUT, check=True)
rows = []
with tempfile.TemporaryDirectory(prefix='tsc-entry-mode-snapshot-') as temporary:
    scratch = Path(temporary)
    shutil.copytree(tree / 'src', scratch / 'src')
    (scratch / 'node_modules').symlink_to(tree / 'node_modules', target_is_directory=True)
    for record in records:
        ordinal = record['ordinal']
        file = scratch / record['file']
        with (out / f'{ordinal:02d}-load.json').open('wb') as stdout, (out / f'{ordinal:02d}-load.stderr').open('wb') as stderr:
            subprocess.run([str(out / 'probe'), str(file)], stdout=stdout, stderr=stderr, check=True)
        report = json.loads((out / f'{ordinal:02d}-load.json').read_text())
        sites = [row for row in (report['dispositions'] or []) if row['site']['file'] == str(file) and
                 row['site']['line'] == record['line'] and row['site']['column'] == record['column'] and
                 row['site']['code'] == int(record['message'].split(':', 1)[0].removeprefix('error TS'))]
        ordinary = [row for row in (report['ordinary_diagnostics'] or []) if row.startswith(f"{file}:{record['line']}:{record['column']}: ")]
        if sites:
            status = 'scheduled' if all(row['state'] == 'scheduled-check' for row in sites) else 'remaining'
        elif ordinary:
            status = 'ordinary-error'
        elif report['loaded'] or report['ordinary_diagnostics']:
            status = 'not-diagnosed'
        else:
            status = 'blocked'
        row = {**record, 'mode_status': status, 'dispositions': sites, 'ordinary_diagnostics': ordinary,
               'project_loaded': report['loaded'], 'load_error': report['error']}
        rows.append(row)
        print(json.dumps({'ordinal': ordinal, 'status': status, 'project_loaded': report['loaded']}), flush=True)
        replacement = baseline / f'{ordinal:02d}-replacement.json'
        if replacement.exists():
            data = json.loads(replacement.read_text())
            raw = file.read_bytes().decode('utf8').encode('utf-16-le')
            start, end = data['start'] * 2, data['end'] * 2
            if raw[start:end].decode('utf-16-le') != data['original']:
                raise RuntimeError('saved replacement does not match')
            file.write_bytes((raw[:start] + data['replacement'].encode('utf-16-le') + raw[end:]).decode('utf-16-le').encode('utf8'))
(out / 'admission.json').write_text(json.dumps(rows, indent=2) + '\n')
