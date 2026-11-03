"""The script readers' counterpart of Require; recipes share configurations.json."""
import json
import os
from pathlib import Path

REGISTRY = json.loads(Path(__file__).with_name('configurations.json').read_text())


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
    return binary or None
