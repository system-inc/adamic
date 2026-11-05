"""Read live timer progress without modifying either census or its output."""
import datetime
import json
from pathlib import Path
import time

root = Path(__file__).resolve().parent / 'evidence'
print(datetime.datetime.now(datetime.timezone.utc).isoformat())
for label in ('base', 'head'):
    timer = root / f'{label}-times.jsonl'
    if timer.exists():
        lines = timer.read_text().splitlines()
        rows = [json.loads(line) for line in lines]
        where = rows[-1]['where'].split('/adapted/')[-1] if rows else 'no completed attempt'
        print(label, len(rows), 'attempts;', where, '; timer age', round(time.time() - timer.stat().st_mtime, 1), 's')
    census = root / f'{label}-latent.jsonl'
    if census.exists():
        print(label, 'findings stream rows:', sum(1 for _ in census.open()))
    exit_file = root / f'{label}-exit.txt'
    if exit_file.exists():
        print(label, 'exit:', exit_file.read_text().strip())
