import json,os,subprocess,time
from pathlib import Path
p=Path("review/test-audit/stage1-cohere-estree-acceptance_grain")
groups=json.loads((p/"groups.json").read_text())
for i,g in enumerate(groups):
 for run in range(3):
  cmd=["timeout","100","go","test","-json","-count=1","-timeout","90s","./stage1/cohere/estree/","-run","^("+"|".join(g)+")$"]
  with (p/f"timing-{i}-{run}.log").open("w") as out:
   start=time.monotonic(); done=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
  with (p/"commands.jsonl").open("a") as out:out.write(json.dumps({"kind":"timing","group":i,"run":run,"command":cmd,"returncode":done.returncode,"wall":time.monotonic()-start})+"\n")
  if done.returncode:break
