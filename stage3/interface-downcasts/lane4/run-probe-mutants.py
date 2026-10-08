#!/usr/bin/env python3
"""Prove shared readiness and owner evidence used by the lane 4 slot probe."""
import json, pathlib, subprocess
repo=pathlib.Path(__file__).resolve().parents[3]
lane=pathlib.Path(__file__).resolve().parent
source=repo/"internal/native/runtime/object.c"
original=source.read_bytes()
mutants={
 "skip-readiness":("if (slot == NULL || !adamic_object_initialized(object)[cache->index]) {", "if (slot == NULL) {"),
 "ignore-owner":("unsigned char storage = adamic_object_field_types(owner)[cache->index];", "unsigned char storage = adamic_object_field_types(object)[cache->index];"),
 "collapse-null":("else if (storage == 12) { value.kind = adamic_view_union_null; }", "else if (storage == 12) { value.kind = adamic_view_union_undefined; }"),
 "ignore-reference-storage":("if (value.kind != adamic_view_union_undefined && value.kind != expected) { value.kind = adamic_view_union_unknown; }", "(void)expected; /* mutant: trust reference kind without storage evidence */"),
 "trust-unknown":("adamic_view_union_value value = {adamic_view_union_unknown, *slot};", "if (storage == 0) { storage = 1; }\n    adamic_view_union_value value = {adamic_view_union_unknown, *slot};")
}
results=[]
try:
 for name,(before,after) in mutants.items():
  text=original.decode();assert text.count(before)==1
  source.write_text(text.replace(before,after))
  log=lane/"logs"/("probe-mutant-"+name+".log")
  with log.open("w") as output:
   result=subprocess.run(["go","test","./internal/oracle","-run","^TestCheckedViewPrimitiveProbe$","-count=1","-v","-timeout","10m"],cwd=repo,stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and "--- FAIL: TestCheckedViewPrimitiveProbe" in observed
  assert "clang failed" not in observed and "[build failed]" not in observed
  assert "sanitized=false" in observed,observed
  results.append({"mutation":name,"caught":True,"log":str(log.relative_to(repo))})
  source.write_bytes(original)
finally:source.write_bytes(original)
(lane/"probe-mutants.json").write_text(json.dumps(results,indent=2)+"\n")
