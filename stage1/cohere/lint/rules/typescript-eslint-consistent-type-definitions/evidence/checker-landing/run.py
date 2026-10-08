import os,sys,time,json,subprocess,pathlib,threading
out=pathlib.Path(sys.argv[2]);out.mkdir(exist_ok=True)
env=os.environ.copy();env.update(GOMAXPROCS="4",GOFLAGS="-buildvcs=false",GOPROXY="https://proxy.golang.org|direct",ADAMIC_LINT_BENCH="1",ADAMIC_TYPESCRIPT_SOURCE="/workspace/wave-07-typescript",ADAMIC_LINT_PROFILE_DIR=str(out/"profiles"),ADAMIC_LINT_PROFILE_SNAPSHOTS=str(out/"profiles"))
(out/"profiles").mkdir(exist_ok=False)
args=["go","test","-json","-count=1","-timeout=60m","./stage1/cohere/lint"]
start=time.monotonic();samples=[]
with (out/"events.jsonl").open("w") as log, (out/"stderr.log").open("w") as err:
 p=subprocess.Popen(args,cwd=sys.argv[1],env=env,stdout=log,stderr=err)
 while p.poll() is None:
  samples.append({"seconds":time.monotonic()-start,"load":os.getloadavg()});time.sleep(5)
 code=p.returncode
results={"command":args,"exit":code,"wall_seconds":time.monotonic()-start,"nproc":len(os.sched_getaffinity(0)),"load":samples,"inputs":{k:v for k,v in env.items() if k.startswith("ADAMIC_")}}
for scope in ["top","all"]:
 counts={"pass":0,"fail":0,"skip":0};skips=[]
 for line in (out/"events.jsonl").read_text().splitlines():
  e=json.loads(line)
  if e.get("Test") and (scope=="all" or "/" not in e["Test"]) and e.get("Action") in counts:
   counts[e["Action"]]+=1
   if e["Action"]=="skip":skips.append(e["Test"])
 results[scope]={"counts":counts,"skips":skips}
(out/"receipt.json").write_text(json.dumps(results,indent=2)+"\n")
print(json.dumps({k:v for k,v in results.items() if k not in ["load","inputs"]}));sys.exit(code)
