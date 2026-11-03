"""The script readers' counterpart of Require; recipes share configurations.json."""
import json
import os
import subprocess
import threading
from pathlib import Path

REGISTRY = json.loads(Path(__file__).with_name('configurations.json').read_text())

_PROBES = {}
_PROBE_LOCK = threading.Lock()


def probe(binary, expression, expected):
    key = (binary, expression, expected)
    with _PROBE_LOCK:
        if key not in _PROBES:
            try:
                result = subprocess.run([binary, '-p', expression], capture_output=True, timeout=10)
                actual = (result.stdout + result.stderr).decode(errors='replace').removesuffix('\n')
                error = None if result.returncode == 0 and actual == expected else f'configuration probe {binary!r} printed {actual!r}; expected {expected!r} (exit {result.returncode})'
            except (OSError, subprocess.TimeoutExpired) as exc:
                error = f'configuration probe {binary!r} printed no successful result; expected {expected!r}: {exc}'
            _PROBES[key] = error
        if _PROBES[key] is not None:
            raise ValueError(_PROBES[key])


def node_for(row, number):
    configuration = row.get('configuration')
    if configuration is None:
        return 'node'
    name = configuration.get('name')
    if name not in REGISTRY:
        raise ValueError(f'row {number}: unknown host configuration {name!r}')
    spec = REGISTRY[name]
    if any(configuration.get(key) != spec[key] for key in ('node', 'configure', 'ruling')):
        raise ValueError(f'row {number}: invalid metadata for host configuration {name}')
    binary = os.environ.get(spec['environment'])
    if not binary:
        print(f"SKIP: row {number} needs {name}; set {spec['environment']}", flush=True)
    if binary:
        probe(binary, spec['probe'], spec['expect'])
    return binary or None
