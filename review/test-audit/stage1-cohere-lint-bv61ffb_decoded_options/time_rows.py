import json,os,subprocess,time,signal
from pathlib import Path
p=Path("review/test-audit/stage1-cohere-lint-bv61ffb_decoded_options")
groups=json.loads((p/"groups.json").read_text())
for i,g in enumerate(groups):
 pattern="^TestCompilerAndStage1Agree_[0-9]+$" if i==8 else "^("+"|".join(g)+")$"
 for run in range(3):
  cmd=["timeout","120","go","test","-json","-count=1","-timeout","90s","./stage1/cohere/lint/","-run",pattern]
  with (p/f"timing-{i}-{run}.log").open("w") as out:
   start=time.monotonic();process=subprocess.Popen(cmd,stdout=out,stderr=subprocess.STDOUT,start_new_session=True);rc=process.wait()
   try:os.killpg(process.pid,signal.SIGKILL)
   except ProcessLookupError:pass
  with (p/"commands.jsonl").open("a") as out:out.write(json.dumps({"kind":"timing","group":i,"run":run,"command":cmd,"returncode":rc,"wall":time.monotonic()-start})+"\n")
  text=(p/f"timing-{i}-{run}.log").read_text()
  normalfails=[]
  for line in text.splitlines():
   try:r=json.loads(line)
   except:continue
   if r.get("Action")=="fail" and r.get("Test"):normalfails.append(r["Test"])
  if normalfails:
   (p/"RED-BASELINE.json").write_text(json.dumps({"group":i,"run":run,"fails":normalfails},indent=2));raise SystemExit("red baseline; stop")
