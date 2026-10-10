import json,subprocess,time,pathlib
p=pathlib.Path("review/compiler/checked-writes-follow/merge-main")
results=[]
for item in json.loads((p/"mutants.json").read_text()):
 start=time.monotonic()
 with (p/(item["name"]+".log")).open("w") as log:
  result=subprocess.run(["timeout","89","go","test","-overlay",item["overlay"],item["package"],"-run",item["selector"],"-count=1","-v","-timeout=85s"],stdout=log,stderr=subprocess.STDOUT)
 body=(p/(item["name"]+".log")).read_text()
 caught=result.returncode==1 and "--- FAIL:" in body and "error: " not in body and "build failed" not in body
 results.append(dict(item,exit=result.returncode,caught=caught,seconds=round(time.monotonic()-start,3)))
 (p/"mutant-results.json").write_text(json.dumps(results,indent=2))
 print(item["name"],"caught" if caught else "NOT CAUGHT",flush=True)
