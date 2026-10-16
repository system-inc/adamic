#!/usr/bin/env python3
"""Outlining regressions, each isolated with a Go source overlay."""
import json, os, subprocess, tempfile
from pathlib import Path
root = Path(__file__).resolve().parents[3]
mutants = [
 ('split-argument-buffer','outline.go','token.text == "[" && index > 0','token.text == "unused_array_token" && index > 0','TestInitializerAggregateBuffers'),
 ('missing-auxiliary-type','outline.go','"adamic_view_union_value", ','','TestInitializerAuxiliaryValueTypes'),
 ('missing-value-type','outline.go','"adamic_typed_array", ','','TestInitializerEveryValueType'),
 ('split-control-flow','outline.go','case "if", "for", "while", "do", "goto", "return":','case "unused_control_token":','TestInitializerFlatBoundaries'),
 ('unbounded-chunks','outline.go','const initializerChunkStatements = 128','const initializerChunkStatements = 1000000','TestInitializerUnitOwnership'),
 ('inline-helpers','outline.go','const initializerAttribute = " __attribute__((noinline))"','const initializerAttribute = ""','TestInitializerStorageSplit'),
 ('reverse-modules','outline.go','for partIndex, indexes := range modules {','for reverse := range modules {\n partIndex := len(modules)-1-reverse\n indexes := modules[partIndex]','TestInitializerStorageSplit'),
 ('unowned-global','units_stable.go','item.owner = helper.initializationOwner','item.owner = "helpers_00.c"','TestInitializerUnitOwnership'),
 ('early-root-release','outline.go','definitions.WriteString("}\\n\\n")','for _, parameter := range parameters { if types[parameter] == "adamic_array *" { fmt.Fprintf(&definitions, "adamic_release(*%s);\\n", parameter) } }; definitions.WriteString("}\\n\\n")','TestInitializerCrossModuleRoot'),
]
# Global readiness writes are the placement proof; erase only that proof.
selected = os.environ.get('ADAMIC_INITIALIZER_MUTANTS', '').split(',')
with tempfile.TemporaryDirectory(prefix='initializer-mutants-') as d:
 for name, file, before, after, test in mutants:
  if selected != [''] and name not in selected: continue
  path=root/'internal/native'/file
  source=path.read_text()
  if before not in source: raise RuntimeError('anchor '+name)
  mutation=Path(d)/(name+'.go.txt'); mutation.write_text(source.replace(before,after,1))
  overlay=Path(d)/'overlay.json'; overlay.write_text(json.dumps({'Replace':{str(path):str(mutation)}}))
  log=Path('/tmp/initializer-mutant-'+name+'.log')
  with log.open('w') as out:
   result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/native','-run','^'+test+'$','-count=1','-timeout=55s','-v'],cwd=root,stdout=out,stderr=subprocess.STDOUT,timeout=58)
  text=log.read_text()
  if result.returncode==0 or '--- FAIL: '+test not in text or '[build failed]' in text: raise RuntimeError(name+' not caught: '+text)
  print(name+' caught by '+test,flush=True)
