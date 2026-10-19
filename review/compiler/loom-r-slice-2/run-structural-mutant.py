import json, os, subprocess, time
from pathlib import Path
root=Path.cwd();out=Path(__file__).parent
os.sched_setaffinity(0, sorted(os.sched_getaffinity(0))[:4])
folder=out/"mutants"/"static-inline-definition";folder.mkdir(parents=True,exist_ok=True)
path="internal/native/units.go";source=Path(path).read_text();before='if d.function && d.tokens[0].text == "static" && d.tokens[1].text == "inline" {'
assert source.count(before)==1
mutant=(folder/"source.go.txt").resolve();mutant.write_text(source.replace(before,"if false {"))
overlay=(folder/"overlay.json").resolve();overlay.write_text(json.dumps({"Replace":{str(root/path):str(mutant)}}))
command=["timeout","180","go","test","./internal/native","-overlay="+str(overlay),"-run","^TestSplitSelfCompareAgreesWithNode$","-count=1","-v","-timeout=90s"]
start=time.monotonic()
with (folder/"test.log").open("w") as log: ran=subprocess.run(command,env=dict(os.environ,GOMAXPROCS="4"),stdout=log,stderr=subprocess.STDOUT,timeout=190)
text=(folder/"test.log").read_text();marker="inline function 'adamic_unit_adamic_same_number' is not defined"
row=dict(mutant="static-inline-definition",caught=ran.returncode==1 and marker in text and "test timed out" not in text,exit=ran.returncode,seconds=round(time.monotonic()-start,3),command=command,markers=[marker],structural_only=True)
with (out/"mutants.jsonl").open("a") as f:f.write(json.dumps(row)+"\n")
print(json.dumps(row),flush=True)
assert row["caught"]
