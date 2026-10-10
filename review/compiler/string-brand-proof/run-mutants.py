"""Run source mutants through Go overlays, without changing production files."""
from pathlib import Path
import json, os, subprocess

repo = Path(__file__).resolve().parents[3]
evidence = repo / "review/compiler/string-brand-proof"
mutants = [
 ("real-field-phantom", "internal/lower/view_unions_primitive_brands.go",
  "return flags&checker.TypeFlagsVoid != 0 || optional && flags&checker.TypeFlagsUndefined != 0", "return true",
  "^TestStringBrandRealRefused$", "want named refusal for actual"),
 ("drop-undefined", "internal/lower/expression.go",
  "if stringShape, missing := l.stringBrandRepresentation(proven); stringShape && missing {\n\t\treturn true",
  "if stringShape, missing := l.stringBrandRepresentation(proven); stringShape && missing {\n\t\treturn false",
  "^TestStringBrandOptional$", "exit codes differ; Node"),
 ("read-brand", "internal/lower/string_brand.go", "observed := readName == field.Name", "observed := false && readName == field.Name",
  "^TestStringBrandReadRefused$", "want named refusal for __brand"),
 ("literal-refinement", "internal/lower/string_brand.go", "return branded && l.checker.IsTypeAssignableTo(fromType, toType)", "return branded && l.checker.IsTypeAssignableTo(fromType, fromType) && toType != nil", "^TestStringBrandCastKeepsLiterals$", "unproven string literal refinement admitted"),
 ("unchecked-missing-cast", "internal/lower/cast.go", 'return ir.Defined{Value: value, Message: "cast failed: undefined is not a " + l.checker.TypeToString(target)}, nil', 'return value, nil', "^TestStringBrandCastChecksMissing$", "missing checked undefined removal"),
]
results=[]
for name, source, old, new, test, catcher in mutants:
 text = (repo/source).read_text()
 assert text.count(old) == 1, (name,text.count(old))
 path = evidence / (name + ".go.txt")
 path.write_text(text.replace(old,new))
 overlay = evidence / (name + "-overlay.json")
 overlay.write_text(json.dumps({"Replace": {str(repo/source): str(path)}}))
 command = ["go","test","-overlay="+str(overlay),("./internal/lower" if name in ("literal-refinement", "unchecked-missing-cast") else "./internal/oracle"),"-run",test,"-v","-count=1","-timeout","90s"]
 with (evidence/(name+".log")).open("w") as log:
  result = subprocess.run(command,cwd=repo,env={**os.environ,"ADAMIC_GATE_UNCACHED":"1"},stdout=log,stderr=subprocess.STDOUT,timeout=300)
 output=(evidence/(name+".log")).read_text()
 caught=result.returncode != 0 and catcher in output and "build failed" not in output
 results.append({"name":name,"exit":result.returncode,"caught":caught,"command":command})
 print(name, "caught" if caught else "NOT CAUGHT",result.returncode,flush=True)
 if not caught: raise SystemExit(1)
(evidence/"mutants.json").write_text(json.dumps(results,indent=2)+"\n")
