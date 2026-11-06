"""Piece provenance and inventory checks, shared by runner and union audit."""
from collections import Counter
import gzip
import hashlib
import json
from pathlib import Path

OUT = Path(__file__).resolve().parent
E = OUT / 'evidence/pieces'
S = OUT / 'scratch'
TREE = S / 'adapted'

def digest(path): return hashlib.sha256(Path(path).read_bytes()).hexdigest()
def dump(path, value):
    path = Path(path); pending = path.with_name(path.name+'.pending')
    pending.write_text(json.dumps(value, indent=2)+'\n'); pending.replace(path)
def load(path):
    opener = gzip.open if str(path).endswith('.gz') else open
    with opener(path, 'rt') as f: return [json.loads(s) for s in f if s.strip()]
def norm(value, tree=TREE):
    if isinstance(value, str): return value.replace(str(tree)+'/', '')
    if isinstance(value, list): return [norm(v, tree) for v in value]
    if isinstance(value, dict): return {k:norm(v, tree) for k,v in value.items()}
    return value

def validate_piece(piece, rows, timers, header, tree=TREE):
    rows, timers = norm(rows, tree), norm(timers, tree)
    assert rows[0] == header and rows[0]['latent_mode'] == 'first-error', 'piece header changed'
    assert Counter(r['file'] for r in rows[1:]) == Counter(f['file'] for f in piece['files']), 'piece file inventory'
    assigned = set(piece['expected_units'])
    actual = [u['where'] for r in rows[1:] for u in r['units']]
    assert Counter(actual) == Counter({u:1 for u in assigned}), 'piece unit inventory'
    assert Counter(t['where'] for t in timers) == Counter({u:1 for u in piece['expected_timed_units']}), 'piece timer inventory'
    for row in rows[1:]:
        for f in row['findings']:
            if f['phase'] == 'lowering':
                assert f['unit'] in assigned, 'finding moved outside assigned unit'
            else:
                assert f['unit'].rsplit(':',2)[0] == row['file'], 'scan owner moved outside selected file'
    return rows, timers

def verify_partition(pieces, wanted):
    counts = Counter(u for p in pieces for u in p['expected_units'])
    assert counts == Counter({u:1 for u in wanted}), 'pieces omit or duplicate units'
