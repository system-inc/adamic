"""Capture exact adapted runner programs without editing cmd/adamic-test262.

Usage: python3 docs/library-object/capture-sources.py TEST262 WORK
Run from the repository root after sourcing the toolchain environment.
"""
from pathlib import Path
import shutil
import subprocess
import sys

root = Path(__file__).resolve().parents[2]
directory = root / 'docs/library-object/.capture-runner'
work = Path(sys.argv[2]).resolve()
work.mkdir(parents=True, exist_ok=True)
assert not (work / 'captures.jsonl').exists(), 'use a fresh work directory'
directory.mkdir()
copied = []
try:
 for source in (root / 'cmd/adamic-test262').glob('*.go'):
  if not source.name.endswith('_test.go'):
   target = directory / source.name
   shutil.copy(source, target)
   copied.append(target)
 path = directory / 'run.go'
 source = path.read_text()
 before = '\t\treport.add(one)'
 after = '''		capture := struct { Result result `json:"result"`; Program string `json:"program"` }{one, classified.Program}
		encoded, encodeErr := json.Marshal(capture)
		if encodeErr != nil { return report, encodeErr }
		captureFile, openErr := os.OpenFile(filepath.Join(e.work, "captures.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if openErr != nil { return report, openErr }
		_, writeErr := captureFile.Write(append(encoded, '\\n'))
		captureFile.Close()
		if writeErr != nil { return report, writeErr }
		report.add(one)'''
 assert source.count(before) == 1
 path.write_text(source.replace(before, after, 1))
 with (work / 'report.json').open('w') as output, (work / 'runner.log').open('w') as log:
  result = subprocess.run(['go', 'run', './docs/library-object/.capture-runner', '-adapt', '-json', '-test262', sys.argv[1], '-work', str(work), 'built-ins/Object'], cwd=root, stdout=output, stderr=log)
 raise SystemExit(result.returncode)
finally:
 for path in copied:
  path.unlink()
 directory.rmdir()
