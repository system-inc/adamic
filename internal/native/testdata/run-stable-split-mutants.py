#!/usr/bin/env python3
"""Run isolated splitter mutants through Go overlays; production files stay intact."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
path = root / "internal/native/units_stable.go"
source = path.read_text()
mutants = [
 ("positional-units", 'compilationUnit{unitName, header.String()', 'compilationUnit{fmt.Sprintf("position_%04d.c", len(units)), header.String()', "TestStableUnitOrderDoesNotNameUnits"),
 ("broad-header", 'common += "#endif\\n"', 'for _, item := range definitions { common += item.declaration }; common += "#endif\\n"', "TestStableUnitInsertion"),
 ("descriptor-migration", 'strings.HasPrefix(d.name, "adamic_string_")', 'false', "TestSharedIdentitiesDoNotMigrate"),
 ("constant-migration", 'strings.Contains(" "+declaration, " const ")', 'false', "TestConstantDescriptorDoesNotMigrate"),
 ("adapter-walk", 'pending = append(pending, target.references...)', 'pending = nil', "TestDeclarationAdapterFollowsCallee"),
 ("abi-guard", 'name + "\\x00" + declaration', 'name + "\\x00"', "TestUnitDeclarationDisagreement"),
 ("abi-conflict", 'if old.declaration != declaration {', 'if false {', "TestSplitClosurePrototypesUseSharedHeader"),
 ("async-ordinal", 'prefix + e.functionName(index)', 'fmt.Sprintf("%s%d", prefix, index)', "TestCanonicalAsyncPrivateIdentity"),
 ("async-literal", 'out.WriteString(source[previous:])\n\tsource = out.String()', 'out.WriteString(source[previous:])\n\tsource = out.String()\n for old, replacement := range replacements { source = strings.ReplaceAll(source, old, replacement) }', "TestCanonicalAsyncPrivateIdentity"),
]
selected = os.environ.get("ADAMIC_SPLIT_MUTANTS", "").split(",")
with tempfile.TemporaryDirectory(prefix="stable-split-mutants-") as temporary:
 for name, before, after, test in mutants:
  if selected != [""] and name not in selected:
   continue
  if before not in source:
   raise RuntimeError("missing mutation anchor: " + name)
  directory = Path(temporary)
  mutation = directory / (name + ".go.txt")
  mutation.write_text(source.replace(before, after, 1))
  overlay = directory / "overlay.json"
  overlay.write_text(json.dumps({"Replace": {str(path): str(mutation)}}))
  log = Path("/tmp/stable-split-mutant-" + name + ".log")
  with log.open("w") as output:
   result = subprocess.run(["go", "test", "-overlay=" + str(overlay), "./internal/native", "-run", "^" + test + "$", "-count=1", "-timeout=55s", "-v"], cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=58)
  text = log.read_text()
  if result.returncode == 0 or "--- FAIL: " + test not in text or "[build failed]" in text:
   raise RuntimeError(name + " was not caught by " + test + ": " + text)
  print(name + " caught by " + test, flush=True)
