import json, os, pathlib, subprocess, time
out = pathlib.Path(__file__).resolve().parent
root = out.parents[2]
groups = json.loads((out / "focused-selection-after-grain.json").read_text())
results=[]
for package,names in groups.items():
 for i in range(0,len(names),4):
  group=names[i:i+4]
  command=["go","test",package,"-run","^("+"|".join(group)+")$","-count=1","-json","-timeout","90s"]
  log=out/"logs"/(package.rsplit("/",1)[-1]+"-%03d.jsonl"%i)
  start=time.monotonic()
  with log.open("w") as stream:
   result=subprocess.run(command,cwd=root,stdout=stream,stderr=subprocess.STDOUT,timeout=180)
  rows=[]
  for line in log.read_text().splitlines():
   try: event=json.loads(line)
   except ValueError: continue
   if event.get("Test") and event.get("Action") in ["pass","fail","skip"]: rows.append(event)
  record=dict(package=package,tests=group,command=command,exit=result.returncode,seconds=time.monotonic()-start,leaves=rows)
  results.append(record)
  (out/"focused-after-grain-results.json").write_text(json.dumps(results,indent=2)+"\n")
  print(package,i,result.returncode,round(record["seconds"],3),flush=True)
  if result.returncode: raise SystemExit(1)
  if any(r.get("Elapsed",0)>=60 for r in rows): raise SystemExit("red leaf exceeded 60 seconds; do not rerun")
