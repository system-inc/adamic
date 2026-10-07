#!/usr/bin/env python3
"""Compare shared parser backends with the independent Go parser on owned inputs."""
import json
from pathlib import Path
import subprocess
import tempfile

repository = Path(__file__).resolve().parents[5]
owned = Path(__file__).resolve().parent
for fixture in sorted(owned.glob('*.ts.txt')):
    observations = []
    commands = {
        'node': ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs',
                 'stage1/typescript/parser/main.ts', str(fixture), '--whole'],
        'javascript': ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs',
                       '/tmp/lint-wave1-09-parser.mjs', str(fixture), '--whole'],
        'native': ['/tmp/lint-wave1-09-parser', str(fixture), '--whole'],
    }
    for backend, command in commands.items():
        result = subprocess.run(command, cwd=repository, capture_output=True)
        (owned / (fixture.stem.removesuffix('.ts') + '-' + backend + '.log')).write_bytes(result.stdout + result.stderr)
        observations.append((result.returncode, result.stdout, result.stderr))
        print(fixture.name, backend, 'exit', result.returncode)
    assert all(item == observations[0] for item in observations)
    assert observations[0][0] == 0

with tempfile.TemporaryDirectory(prefix='wave109-go-parser-') as directory:
    scratch = Path(directory)
    virtual = repository / 'cohere/TypeScript/tsc/wave109_parser.go'
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {
        str(virtual): str(repository / 'stage1/typescript/parser/testdata/oracle.go'),
    }}))
    binary = scratch / 'parser'
    result = subprocess.run(['go', 'build', '-overlay=' + str(overlay), '-o', str(binary), str(virtual)],
                            cwd=repository / 'cohere/TypeScript/tsc', capture_output=True)
    assert result.returncode == 0, result.stderr
    for fixture in sorted(owned.glob('*.ts.txt')):
        result = subprocess.run([str(binary), str(fixture), '--whole'], capture_output=True)
        (owned / (fixture.stem.removesuffix('.ts') + '-go.log')).write_bytes(result.stdout + result.stderr)
        print(fixture.name, 'Go exit', result.returncode)
        assert result.returncode == 0
        expected = owned / (fixture.stem.removesuffix('.ts') + '-node.log')
        assert result.stdout + result.stderr == expected.read_bytes()
