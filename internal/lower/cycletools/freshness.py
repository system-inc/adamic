"""Build two decision probes with temporary, non-production freshness tracing.

Usage: python3 internal/lower/cycletools/freshness.py BEFORE_REVISION SCRATCH
The probes run the same current lowering with old/final freshness implementations.
The overlay prints the completed ProveWrites result without altering any facts.
"""
import json
import pathlib
import subprocess
import sys

repository = pathlib.Path(__file__).resolve().parents[3]
revision, directory = sys.argv[1:]
scratch = pathlib.Path(directory).resolve()
scratch.mkdir(parents=True, exist_ok=True)
source = repository / 'internal/fresh/fresh.go'
old = subprocess.check_output(['git', 'show', revision + ':internal/fresh/fresh.go'],
                              cwd=repository, text=True)
for name, text in [('old', old), ('new', source.read_text())]:
    assert text.count('\treturn writes\n') == 1
    text = text.replace('import (\n', 'import (\n\t"encoding/json"\n\t"os"\n', 1)
    text = text.replace('\treturn writes\n', '''\tif os.Getenv("ADAMIC_TRACE_FRESH") == "1" {
        encoded, err := json.Marshal(writes)
        if err != nil { panic(err) }
        fmt.Fprintf(os.Stderr, "fresh decisions: %s\\n", encoded)
    }
    return writes
''', 1)
    instrumented = scratch / (name + '-fresh.go')
    instrumented.write_text(text)
    subprocess.run(['gofmt', '-w', str(instrumented)], check=True)
    overlay = scratch / (name + '-overlay.json')
    overlay.write_text(json.dumps({'Replace': {str(source): str(instrumented)}}) + '\n')
    with (scratch / (name + '-build.log')).open('w') as log:
        subprocess.run(['go', 'build', '-buildvcs=false', '-overlay', str(overlay),
                        '-o', str(scratch / name), './internal/lower/cycletools'],
                       cwd=repository, stdout=log, stderr=log, check=True)
