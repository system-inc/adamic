import json,os,subprocess,time,difflib
from pathlib import Path
p=Path("review/test-audit/stage1-cohere-estree-acceptance_grain")
base=subprocess.check_output(["git","rev-parse","HEAD"]).decode().strip()
port="stage1/cohere/estree/"
experiments=[
 ("M1",port+"protocol.ts","unit >= 32 && unit <= 126 && unit !== 92","unit >= 33 && unit <= 126 && unit !== 92",["TestAcceptanceGrammar","TestAcceptanceDiagnostics"],"production"),
 ("M2",port+"protocol.ts","depth = 0","depth = 1",["TestAcceptanceGrammar","TestAcceptanceDiagnostics"],"production"),
 ("M3",port+"pipeline.ts","ESTree parser: ${syntax}","ESTree validator: ${syntax}",["TestAcceptanceGrammar","TestAcceptanceDiagnostics"],"production"),
 ("M4",port+"pipeline.ts",'return panic(`ESTree parser: ${syntax}`);',"return panic('ESTree parser: rejected');",["TestAcceptanceGrammar","TestAcceptanceDiagnostics"],"production"),
 ("P1",port+"pipeline.ts","export function answer(path: string, text: string): string {","export function answer(path: string, text: string): string {\n    return '';",["TestAcceptanceGrammar","TestAcceptanceDiagnostics"],"probe"),
 ("W1",port+"estree_test.go","if bytes.Equal(want, got) {","if len(want) >= 0 {",["TestAcceptanceMutants_000","TestAcceptanceMutants_001","TestAcceptanceMutantsUnion","TestDeepMutants_000","TestDeepMutants_001","TestDeepMutants_002"],"witness"),
 ("W2",port+"acceptance_grain_test.go",'if firstDifference(want, got) == "" {','if len(want) < 0 {',["TestAcceptanceMutantsPlanted_000","TestAcceptanceMutantsPlanted_001"],"witness"),
 ("W3",port+"stalls_test.go","func refusedBeforeDeadline(t *testing.T, argv []string, diagnostic string) {","func refusedBeforeDeadline(t *testing.T, argv []string, diagnostic string) {\n    if diagnostic != " + chr(34) + chr(34) + " { return }",["TestAcceptanceDiagnosticControl"],"witness"),
 ("S1",port+"acceptance_grain_test.go",'return os.WriteFile(filepath.Join(directory, "oracle"), data, 0755)','_ = data; return nil',["TestProduct_AcceptanceOracle"],"setup"),
 ("S2",port+"acceptance_grain_test.go","func acceptanceMutantProducts(t *testing.T, item acceptanceMutant, compile bool) (string, string) {",'func acceptanceMutantProducts(t *testing.T, item acceptanceMutant, compile bool) (string, string) {\n    if item.name != "" { return "/tmp/u084-missing-main.ts", "/tmp/u084-missing-port" }',["TestProduct_AcceptanceCatchLowered","TestProduct_AcceptanceCatchNative","TestProduct_AcceptanceClassLowered","TestProduct_AcceptanceClassNative"],"setup"),
 ("S3",port+"deep_mutants_product_shards_test.go",'return os.WriteFile(filepath.Join(dir, "want"), output, 0644)','_ = output; return nil',["TestDeepMutants_Setup"],"setup"),
 ("S4",port+"acceptance_grain_test.go",'filepath.Join(directory, "oracle"), data, 0755','filepath.Join(directory, "missing/oracle"), data, 0755',["TestProduct_AcceptanceOracle"],"setup"),
 ("S5",port+"acceptance_grain_test.go",'filepath.Join(dir, "port.c"), []byte(native.C(lowered))','filepath.Join(dir, "missing/port.c"), []byte(native.C(lowered))',["TestProduct_AcceptanceCatchLowered","TestProduct_AcceptanceCatchNative","TestProduct_AcceptanceClassLowered","TestProduct_AcceptanceClassNative"],"setup"),
 ("S6",port+"deep_mutants_product_shards_test.go",'filepath.Join(dir, "want"), output, 0644','filepath.Join(dir, "missing/want"), output, 0644',["TestDeepMutants_Setup"],"setup"),
]
(p/"diffs").mkdir(exist_ok=True)
for id,file,old,new,rows,kind in experiments:
 original=subprocess.check_output(["git","show",base+":"+file]).decode()
 assert original.count(old)==1,(id,original.count(old))
 modified=original.replace(old,new,1)
 diff="".join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile="a/"+file,tofile="b/"+file))
 (p/"diffs"/(id+".diff")).write_text(diff)
 line=original[:original.index(old)].count("\n")+1
 with (p/"experiment-menu.jsonl").open("a") as out:out.write(json.dumps({"id":id,"kind":kind,"file":file,"line":line,"from":old,"to":new,"rows":rows})+"\n")
 Path(file).write_text(modified)
 try:
  env=os.environ.copy();env["ADAMIC_BUILD_CACHE_DIR"]="/tmp/u084/cache/"+id
  if file.endswith(".go"):
   # Validate every standalone Go harness experiment with go vet.
   with (p/(id+"-vet.log")).open("w") as out:
    vet=subprocess.run(["go","vet","./stage1/cohere/estree/"],env=env,stdout=out,stderr=subprocess.STDOUT)
  cmd=["timeout","100","go","test","-json","-count=1","-timeout","90s","./stage1/cohere/estree/","-run","^("+"|".join(rows)+")$"]
  with (p/(id+".log")).open("w") as out:
   start=time.monotonic();done=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  with (p/"commands.jsonl").open("a") as out:out.write(json.dumps({"kind":kind,"id":id,"command":cmd,"cache":env["ADAMIC_BUILD_CACHE_DIR"],"returncode":done.returncode,"wall":time.monotonic()-start})+"\n")
 finally:Path(file).write_text(original)
 if id == "M4":
  # Reapply only for a direct observable output witness, then restore.
  Path(file).write_text(modified)
  try:
   with (p/"M4-witness-before.stderr").open("w") as out:
    Path(file).write_text(original)
    before=subprocess.run(["node","--disable-warning=ExperimentalWarning","oracle/node.mjs",port+"main.ts","/tmp/u084/diagnostic.ts"],stdout=subprocess.DEVNULL,stderr=out).returncode
   Path(file).write_text(modified)
   with (p/"M4-witness-after.stderr").open("w") as out:
    after=subprocess.run(["node","--disable-warning=ExperimentalWarning","oracle/node.mjs",port+"main.ts","/tmp/u084/diagnostic.ts"],stdout=subprocess.DEVNULL,stderr=out).returncode
  finally:Path(file).write_text(original)
  (p/"M4-witness.json").write_text(json.dumps({"command":["node","--disable-warning=ExperimentalWarning","oracle/node.mjs",port+"main.ts","/tmp/u084/diagnostic.ts"],"input":"++await 42;","before_exit":before,"after_exit":after},indent=2)+"\n")
 if kind == "setup":
  root=Path(env["ADAMIC_BUILD_CACHE_DIR"])
  products=[{"directory":str(d),"files":[str(f.relative_to(d)) for f in d.rglob("*") if f.is_file()]} for d in root.iterdir() if d.is_dir()] if root.exists() else []
  (p/(id+"-artifacts.json")).write_text(json.dumps({"products":products,"missing_main_exists":Path("/tmp/u084-missing-main.ts").exists(),"missing_port_exists":Path("/tmp/u084-missing-port").exists()},indent=2)+"\n")
