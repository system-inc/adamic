#!/usr/bin/env python3
"""Hold host-process's unchanged upstream fixtures to recorded Node observations."""
import argparse
import json
from pathlib import Path
import subprocess
import tempfile

repo = Path(__file__).resolve().parents[2]
bucket = repo / 'stage3/fixtures/host'
owned = {
    '14_getCurrentDirectory.a': ('callback = undefined!;', '// callback stays live'),
    '15_getExecutingFilePath.a': ('__filename.endsWith("sys.js")', '__filename.endsWith("sys.cjs")'),
    '16_getEnvironmentVariable.a': ('return process.env[name] || "";', 'return "";'),
    '17_write.a': ('process.stdout.write(s);', 'process.stdout.write(s + "?");'),
    '18_exit_0.a': ('nodeSystem.exit(0);', 'nodeSystem.exit(1);'),
    '19_exit_1.a': ('nodeSystem.exit(1);', 'nodeSystem.exit(2);'),
    '20_exit_2.a': ('nodeSystem.exit(2);', 'nodeSystem.exit(0);'),
    '23_newLine.a': ('newLine: _os.EOL', 'newLine: "\\r\\n"'),
}
parser = argparse.ArgumentParser()
parser.add_argument('--compiler', type=Path, required=True)
parser.add_argument('--logs', type=Path, required=True)
args = parser.parse_args()
args.compiler = args.compiler.resolve()
args.logs.mkdir(parents=True, exist_ok=True)
rows = {row['file']: row for row in json.loads((bucket / 'status.json').read_text())}


def run(command, label):
    result = subprocess.run(command, cwd=repo, capture_output=True, timeout=120)
    (args.logs / (label + '.stdout')).write_bytes(result.stdout)
    (args.logs / (label + '.stderr')).write_bytes(result.stderr)
    (args.logs / (label + '.exit')).write_text(str(result.returncode) + '\n')
    return {'stdout': result.stdout.decode(), 'stderr': result.stderr.decode(), 'exit': result.returncode}


def stage(observed):
    if observed['exit'] == 0:
        return 'Compiles'
    diagnostic = observed['stdout'] + observed['stderr']
    if ': error TS' in diagnostic:
        return 'Checker'
    if "stage 0 can't lower" in diagnostic:
        return 'NotYet'
    if 'Adamic 0.1 refuses' in diagnostic:
        return 'Refused'
    raise AssertionError('unexpected compiler/tool failure: ' + diagnostic)


def agree(observed, expected, label):
    assert observed == expected, label + ': exact stdout/stderr/exit disagreement'


summary = []
with tempfile.TemporaryDirectory(prefix='adamic-process-acceptance-') as scratch_name:
    scratch = Path(scratch_name)
    for name, (before, after) in owned.items():
        source = bucket / name
        expected = rows[name]['node']
        truth = run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(source)], name + '.node')
        agree(truth, expected, name + ' recorded Node')
        binary = scratch / name.replace('.a', '')
        built = run([str(args.compiler), 'build', str(source), '-o', str(binary), '--sanitize'], name + '.native-build')
        js = run([str(args.compiler), 'js', str(source)], name + '.js-build')
        native_stage, js_stage = stage(built), stage(js)
        if native_stage == 'Compiles':
            agree(run([str(binary)], name + '.native'), expected, name + ' native')
            native_stage = 'Agrees'
        if js_stage == 'Compiles':
            script = scratch / (name + '.mjs')
            script.write_text(js['stdout'])
            agree(run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(script)], name + '.javascript'), expected, name + ' JavaScript')
            js_stage = 'Agrees'
        original = source.read_text()
        assert before in original, (name, before)
        mutant = scratch / name
        mutant.write_text(original.replace(before, after, 1))
        bad = run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(mutant)], name + '.mutant-node')
        assert bad != expected, name + ': surviving semantic mutant'
        record = {'fixture': name, 'node': 'Agrees', 'native': native_stage, 'javascript': js_stage, 'node_mutant': 'Caught'}
        summary.append(record)
        print(json.dumps(record))
(args.logs / 'stages.json').write_text(json.dumps(summary, indent=2) + '\n')
raise SystemExit(0 if all(row['native'] == row['javascript'] == 'Agrees' for row in summary) else 1)
