#!/usr/bin/env python3
"""Produce the shared artifact and stock API parser before gate test units start."""
import argparse
import fcntl
import json
from pathlib import Path
import shutil
import tempfile
from artifact import prepare, payload_digest


def prepare_shared(cache):
    prepared = prepare(cache)
    artifact = Path(prepared['tree']).parent
    # The tree's locked stock parser, stage3/api's installed packages. Bundle it for per-unit caches.
    source = Path(__file__).resolve().parents[1] / 'api/node_modules/typescript'
    with (artifact / 'shards.lock').open('a') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        stock = artifact / 'api'
        if not stock.exists():
            temporary = Path(tempfile.mkdtemp(prefix='api.', dir=artifact)) / 'typescript'
            shutil.copytree(source, temporary)
            if json.loads((temporary / 'package.json').read_text())['version'] != '6.0.3':
                raise ValueError('shared API parser is not stock TypeScript 6.0.3')
            temporary.rename(stock)
        api_digest = payload_digest(stock)
        marker = artifact / 'shards-ready.json'
        ready = dict(artifact=prepared['key'], api_digest=api_digest)
        if marker.exists() and json.loads(marker.read_text()) != ready:
            raise ValueError('shared API parser bytes changed')
        marker.write_text(json.dumps(ready, indent=2) + '\n')
    return dict(prepared, artifact=str(artifact), api_digest=api_digest)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('cache', type=Path)
    args = parser.parse_args()
    print(json.dumps(prepare_shared(args.cache.resolve()), indent=2))
