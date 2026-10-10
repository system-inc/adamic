"""Compare every oracle testdata program using the baseline and delivery CLIs."""
import concurrent.futures, json, subprocess, hashlib
from pathlib import Path
repo=Path(__file__).resolve().parents[3]
evidence=repo/"review/compiler/string-brand-proof"
files=sorted(p for p in (repo/"internal/oracle/testdata").rglob("*") if p.suffix in (".a",".ts") and not p.name.endswith(".d.ts"))

cachepath=evidence/"admission.json"
if not cachepath.exists(): cachepath=evidence/"admission-progress.json"
cachedrows=json.loads(cachepath.read_text()) if cachepath.exists() else []
if isinstance(cachedrows,dict): cachedrows=cachedrows["rows"]
baseline={r["path"]:r["main"] for r in cachedrows}
binary_hashes={phase:hashlib.sha256(Path(binary).read_bytes()).hexdigest() for phase,binary in [("main","/tmp/adamic-string-brand-main"),("after","/tmp/adamic-string-brand-after")]}

def measure(path):
 observations={}
 for name, binary in [("main","/tmp/adamic-string-brand-main"),("after","/tmp/adamic-string-brand-after")]:
  relative=str(path.relative_to(repo))
  if name=="main" and relative in baseline:
   observations[name]=baseline[relative]
  else:
   result=subprocess.run([binary,"c",str(path)],cwd=repo,capture_output=True,timeout=20)
   observations[name]={"exit":result.returncode,"stderr":result.stderr.decode(errors="replace")}
 return {"path":str(path.relative_to(repo)),**observations}
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 rows=[]
 for row in pool.map(measure,files):
  rows.append(row)
  if len(rows)%100==0:
   print("compared",len(rows),"of",len(files),flush=True)
   (evidence/"admission-progress.json").write_text(json.dumps(rows,indent=2)+"\n")
new=[r["path"] for r in rows if r["main"]["exit"]!=0 and r["after"]["exit"]==0]
lost=[r["path"] for r in rows if r["main"]["exit"]==0 and r["after"]["exit"]!=0]
reasons=["a value of type __String | undefined","a function returning __String | undefined"]
counts={phase:{reason:sum("can't lower "+reason+" yet" in r[phase]["stderr"] for r in rows) for reason in reasons} for phase in ["main","after"]}
summary={"binary_sha256":binary_hashes,"files":len(rows),"newly_admitted":new,"lost_admissions":lost,"first_error_counts":counts,"mode":"CLI first-error per file; not the historical full-mode 44-site census","rows":rows}
(evidence/"admission.json").write_text(json.dumps(summary,indent=2)+"\n")
print(json.dumps({k:v for k,v in summary.items() if k!="rows"},indent=2),flush=True)
