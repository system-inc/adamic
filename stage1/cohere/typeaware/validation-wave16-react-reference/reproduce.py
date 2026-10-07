import argparse
import gzip
import hashlib
import json
import subprocess
from collections import Counter
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('oracle', type=Path)
parser.add_argument('--rounds', type=int, default=64)
args = parser.parse_args()
base = Path(__file__).resolve().parent
summary = {'rounds': args.rounds, 'cases': {}, 'sources': {}}
for case in ['two-creator', 'one-creator']:
    variants = Counter()
    creators = Counter()
    for round_ in range(args.rounds):
        result = subprocess.run([str(args.oracle.resolve()), str(base/'tsconfig.json'), str(base/(case+'.manifest'))], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        (base/f'{case}-{round_:02d}.stdout.gz').write_bytes(gzip.compress(result.stdout, mtime=0))
        (base/f'{case}-{round_:02d}.stderr').write_bytes(result.stderr)
        if result.returncode:
            raise SystemExit(f'{case} round {round_}: exit {result.returncode}')
        if not result.stdout.endswith(b'findings 1\n'):
            raise SystemExit(f'{case} round {round_}: missing positive finding')
        digest = hashlib.sha256(result.stdout).hexdigest()
        variants[digest] += 1
        for name in ['createA', 'createB']:
            if f'at `{name}`'.encode() in result.stdout:
                creators[name] += 1
    summary['cases'][case] = {'distinct_stdout':len(variants), 'hash_frequencies':dict(variants), 'creator_frequencies':dict(creators)}
for name in ['two-creator.a','one-creator.a','tsconfig.json']:
    summary['sources'][name] = hashlib.sha256((base/name).read_bytes()).hexdigest()
summary['sources']['oracle_wave16_react.go'] = hashlib.sha256((base.parent/'testdata/oracle_wave16_react.go').read_bytes()).hexdigest()
(base/'results.json').write_text(json.dumps(summary, indent=2, sort_keys=True)+'\n')
print(json.dumps(summary['cases'], sort_keys=True))
