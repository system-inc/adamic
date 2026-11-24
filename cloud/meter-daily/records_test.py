#!/usr/bin/env python3
"""No meter record newer than the last full one holds a raw log or a file over 1 MB (#jnx6n47, Kirk's ruling of Nov 24).

The 43 records up to 20261122T131040Z.tNkys1 keep their logs and their full report.json: rewriting public history isn't
worth it, by Kirk's word. Every record after them is slim.py's shape, its raw files left with the job.

Mutant: a 2 MB file planted under a new stamp in stage3/meter/runs turns test_newer_records_hold_no_log_and_no_bulk
red, as test_a_planted_file_goes_red shows on a scratch copy of the shape.
"""
import os
from pathlib import Path
import tempfile
import unittest

RUNS = Path(__file__).resolve().parents[2] / 'stage3/meter/runs'
# The newest record committed whole; only records after it are held to the slim shape.
LAST_FULL = '20261122T131040Z.tNkys1'
# slim.py's limit on any file a record commits.
LIMIT = 1_048_576


def newer(stamp):
    """A record after LAST_FULL: a later second, or the same second under another suffix."""
    second = stamp.split('.', 1)[0]
    return second > LAST_FULL.split('.', 1)[0] or (second == LAST_FULL.split('.', 1)[0] and stamp != LAST_FULL)


def offenses(runs):
    """Each file under a newer record that a record may not hold, by its path under runs."""
    found = []
    for record in sorted(path for path in runs.iterdir() if path.is_dir() and newer(path.name)):
        for directory, directories, files in os.walk(record):
            directories.sort()
            for name in sorted(files):
                path = Path(directory) / name
                relative = path.relative_to(runs)
                if name.endswith(('.log', '.log.gz')):
                    found.append(f'{relative} is a log')
                elif not path.is_symlink() and path.stat().st_size > LIMIT:
                    found.append(f'{relative} is {path.stat().st_size:,} bytes, past {LIMIT:,}')
    return found


class Records(unittest.TestCase):
    def test_newer_records_hold_no_log_and_no_bulk(self):
        self.assertTrue((RUNS / LAST_FULL).is_dir(), f'{RUNS / LAST_FULL} is gone, so nothing marks where slim records start')
        self.assertEqual(offenses(RUNS), [])

    def test_a_planted_file_goes_red(self):
        with tempfile.TemporaryDirectory() as scratch:
            runs = Path(scratch)
            for stamp in (LAST_FULL, '20261121T163134Z.JQGsxl'):
                (runs / stamp / 'main').mkdir(parents=True)
                (runs / stamp / 'main/census.log.gz').write_bytes(b'kept as it was')
                (runs / stamp / 'report.json').write_bytes(b' ' * (2 * LIMIT))
            self.assertEqual(offenses(runs), [], 'a record up to the last full one is held to the slim shape')
            for stamp in ('20261125T131040Z.Abc123', '20261122T131040Z.zzzzzz'):
                (runs / stamp / 'main').mkdir(parents=True)
                (runs / stamp / 'report.md').write_text('Whole program: main: 2/79\n')
                (runs / stamp / 'main/report.json').write_bytes(b' ' * (2 * LIMIT))
                (runs / stamp / 'main/census.log.gz').write_bytes(b'raw')
                (runs / stamp / 'job-meter.log').write_text('raw\n')
            self.assertEqual(offenses(runs), [
                '20261122T131040Z.zzzzzz/job-meter.log is a log',
                '20261122T131040Z.zzzzzz/main/census.log.gz is a log',
                '20261122T131040Z.zzzzzz/main/report.json is 2,097,152 bytes, past 1,048,576',
                '20261125T131040Z.Abc123/job-meter.log is a log',
                '20261125T131040Z.Abc123/main/census.log.gz is a log',
                '20261125T131040Z.Abc123/main/report.json is 2,097,152 bytes, past 1,048,576',
            ])


if __name__ == '__main__':
    unittest.main()
