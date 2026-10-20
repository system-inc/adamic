#!/usr/bin/env python3
"""Run assertion-targeted report mutants without touching meter sources."""
import os
from pathlib import Path
import shutil
import subprocess

run = Path(__file__).resolve().parents[1]
source = Path('/workspace/adamic/stage3/meter')
base = run / 'validation/report-mutants'
base.mkdir(exist_ok=True)
mutants = [
    ('own-file-is-whole-program', "'checker_own_file': Path(name).suffix in ('.ts', '.a') and kind != 'error' and name not in located,", "'checker_own_file': checked,", 'report_test.ReportTests.test_planted_diagnostic_changes_only_its_own_file'),
    ('drop-refused', "reasons = Counter((site[0], site[2]) for site in sites if site[0] in ('NotYet', 'Refused'))", "reasons = Counter((site[0], site[2]) for site in sites if site[0] == 'NotYet')", 'report_test.LatentReportTests.test_table_groups_same_reason_across_kinds'),
]
for name, before, after, test in mutants:
    folder = base / name
    folder.mkdir(exist_ok=True)
    for filename in ['report.py', 'report_test.py', 'owners.json']:
        shutil.copy2(source / filename, folder / filename)
    path = folder / 'report.py'
    text = path.read_text()
    assert text.count(before) == 1, name
    path.write_text(text.replace(before, after))
    with (folder / 'witness.log').open('w') as log:
        result = subprocess.run(['python3', '-m', 'unittest', '-v', test], cwd=folder, stdout=log, stderr=log, timeout=120)
    observed = (folder / 'witness.log').read_text()
    assert result.returncode == 1 and 'AssertionError' in observed and 'FAILED (failures=1)' in observed, (name, observed)
    print(f'{name}: caught by {test}, exit {result.returncode}')
