"""Exercise review completeness, source identity, and one-witness-per-event checks."""
import json
import subprocess
import sys
import tempfile
from pathlib import Path

root = Path(__file__).resolve().parents[2]
script = root / 'notes/discriminant-writes/count-typescript.mjs'
review = root / 'notes/discriminant-writes/construction-review.json'
source_root, package_root, independent_path = sys.argv[1:4]
work = Path(tempfile.mkdtemp(prefix='adamic-census-mutants-'))

for name, expected_error in (
    ('missing_review', 'construction review missing'),
    ('stale_source_hash', 'review source hash changed'),
    ('duplicate_supplied_event', 'duplicate supplied event'),
):
    reviews = json.loads(review.read_text())
    independent = json.loads(Path(independent_path).read_text())
    if name == 'missing_review':
        reviews.pop(0)
    elif name == 'stale_source_hash':
        reviews[0]['source_sha256'] = '0' * 64
    else:
        independent['writes'].append(independent['writes'][0])
    changed_review = work / (name + '-review.json')
    changed_reference = work / (name + '-reference.json')
    changed_review.write_text(json.dumps(reviews))
    changed_reference.write_text(json.dumps(independent))
    log = Path(tempfile.gettempdir()) / ('discriminant-census-mutant-' + name + '.log')
    with log.open('w') as output:
        result = subprocess.run(
            ['node', str(script), source_root, package_root, str(changed_reference), str(changed_review)],
            cwd=root, stdout=output, stderr=subprocess.STDOUT,
        )
    assert result.returncode != 0, 'survived ' + name
    assert expected_error in log.read_text(), 'wrong killer ' + name
    print(name, result.returncode, expected_error, str(log), flush=True)
