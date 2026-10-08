#!/usr/bin/env python3
"""Mutate real enumeration, keep complete logs, and restore each implementation."""
from pathlib import Path
import os
import subprocess

repository = Path(__file__).resolve().parents[3]
logs = Path("/tmp/entries-provenance-mutants")
logs.mkdir(exist_ok=True)
mutants = [
 ("javascript-drop-readiness", "internal/javascript/entries_provenance.go", "adamicFieldReadiness.get(object)?.has(key) ? undefined : object[key]", "object[key]", 1, "^TestEntriesRuntimeReadiness$", "exit codes differ"),
 ("native-drop-readiness", "internal/native/runtime/library_object.c", "adamic_object_initialized(object)[cache.index] ? adamic_dynamic_property(&object->heap, name) : NULL", "adamic_dynamic_property(&object->heap, name)", 1, "^TestEntriesRuntimeReadiness$", "exit codes differ"),
 ("lowering-unproven-as-proven", "internal/lower/library_object.go", "call.Checked = !l.enumerationProven(written[0], element, 0)", "call.Checked = false", 1, "^TestEntriesProvenance$/entries_checked_misfit$", "exit codes differ"),
 ("native-drop-key", "internal/native/runtime/library_object.c", "for (size_t at = 0; at < object->shape->count; at++)", "for (size_t at = 1; at < object->shape->count; at++)", 2, "^TestEntriesProvenance$/entries_(proven|checked_fit)$", "stdout differs"),
 ("javascript-drop-key", "internal/javascript/entries_provenance.go", "Object.keys(object)", "Object.keys(object).slice(1)", 1, "^TestEntriesProvenance$/entries_checked_fit$", "stdout differs"),
]
environment = dict(os.environ, ADAMIC_GATE_UNCACHED="1")
for name, relative, before, after, count, selection, catcher in mutants:
 path = repository / relative
 original = path.read_text()
 if original.count(before) < count:
  raise SystemExit(f"{name}: mutation site missing")
 try:
  path.write_text(original.replace(before, after, count))
  log = logs / (name + ".log")
  with log.open("w") as output:
   result = subprocess.run(["go", "test", "./internal/oracle", "-run", selection, "-count=1", "-v"], cwd=repository, env=environment, stdout=output, stderr=subprocess.STDOUT)
  evidence = log.read_text()
  if result.returncode == 0 or catcher not in evidence or "clang:" in evidence or "build failed" in evidence:
   raise SystemExit(f"{name}: not caught by {catcher}; see {log}")
  print(f"{name}: caught by {catcher}; {log}")
 finally:
  path.write_text(original)
