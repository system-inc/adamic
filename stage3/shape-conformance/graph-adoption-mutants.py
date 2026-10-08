"""Run valid C mutants; each must fail the adoption semantic regression."""
import pathlib, subprocess
root = pathlib.Path.cwd()
source_path = root/'internal/native/runtime/graph_regions.c'
source = source_path.read_text()
mutants = {
 'stale-inline-map': ('if (inline_map) { ((adamic_map *)heap)->entries = ((adamic_map *)heap)->small; }', '(void)inline_map;'),
 'overwrite-external-map': ('if (inline_map) { ((adamic_map *)heap)->entries = ((adamic_map *)heap)->small; }', 'if (heap->kind == adamic_kind_map) { ((adamic_map *)heap)->entries = ((adamic_map *)heap)->small; } (void)inline_map;'),
 'truncate-object-metadata': ('if (heap->kind == adamic_kind_object) { bytes = adamic_object_size(((adamic_object *)heap)->shape->count); }', '/* mutant: trust the stale caller size */'),
}
try:
 for name, (before, after) in mutants.items():
  assert source.count(before) == 1, name
  source_path.write_text(source.replace(before,after))
  logpath = root/'stage3/shape-conformance/logs'/('graph-mutant-'+name+'.log')
  with logpath.open('w') as log:
   status = subprocess.run(['go','test','./internal/native','-run','^TestShapeGraphMapAdoption$','-count=1','-v'],stdout=log,stderr=subprocess.STDOUT).returncode
  text = logpath.read_text()
  assert status != 0 and 'sanitize=false:' in text and '[build failed]' not in text and 'compile generated C' not in text, (name,text)
  print(name+': adoption assertion caught valid C mutant',flush=True)
finally:
 source_path.write_text(source)
