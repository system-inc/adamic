#!/usr/bin/env python3
"""Exercise the two stock stores with pre-store checks on the pinned CLI inputs.
This is a Node control, not a native build of stock tsc. Scratch files only.
"""
import argparse, hashlib, json, os, pathlib, shutil, subprocess
parser=argparse.ArgumentParser()
parser.add_argument("--typescript",required=True)
parser.add_argument("--out",required=True)
a=parser.parse_args()
repo=pathlib.Path(__file__).resolve().parents[2]
pkg=pathlib.Path(a.typescript).resolve()
out=pathlib.Path(a.out).resolve(); out.mkdir(parents=True,exist_ok=True)
assert json.loads((pkg/"package.json").read_text())["version"]=="6.0.3"
lib=out/"typescript"; shutil.copytree(pkg,lib,dirs_exist_ok=True)
evidence=out/"witnesses"; evidence.mkdir(exist_ok=True)
pin="3255eb1e"
for name in ["probe.cjs","flags.cjs","parent.cjs"]:
 data=subprocess.check_output(["git","show",pin+":stage3/evidence/two-writes/"+name],cwd=repo)
 (evidence/name).write_bytes(data)
preload=evidence/"checked.cjs"
preload.write_text(r'''const probe = require('./probe.cjs');
const domains = new WeakMap();
const register = probe.flagsDomain;
probe.flagsDomain = (node, values) => { domains.set(node, values); register(node, values); };
globalThis.__twoWritesChecked = (site, target, value, original) => {
 const allowed = domains.get(original);
 const valid = site === "parent" ? value !== undefined : allowed ? allowed.includes(value) : Number.isInteger(value) && value >= 0 && value <= 0x7fffffff;
 if (!valid) { const expression = site === "flags" ? "(node as Mutable<T>).flags" : "(visited as Mutable<T>).parent"; const expected = site === "flags" ? "NodeFlags.Synthesized" : "Node"; process.stderr.write("adamic: panic: write failed: " + expression + " expects " + expected + ", got " + String(value) + "\n"); process.exit(70); }
};
''')
changes=[]
for name in ["typescript.js","_tsc.js"]:
 path=lib/"lib"/name; source=path.read_text(); before=hashlib.sha256(path.read_bytes()).hexdigest()
 for site,statement,arguments in [("flags","node.flags = newFlags;","node, newFlags, node"),("parent","visited.parent = void 0;","visited, void 0, node")]:
  assert source.count(statement)==1,(name,statement)
  source=source.replace(statement, 'globalThis.__twoWritesChecked("'+site+'", '+arguments+'); '+statement+' globalThis.__twoWrites.write("'+site+'", '+arguments+');')
 path.write_text(source); changes.append({"file":name,"original_sha256":before,"checked_sha256":hashlib.sha256(path.read_bytes()).hexdigest()})
env=dict(os.environ,NODE_OPTIONS="--require="+str(preload))
runs=[]
for name,cmd,exit_code in [("flags",["node",str(evidence/"flags.cjs"),str(lib/"lib/typescript.js")],70),("parent",["node",str(evidence/"parent.cjs"),str(lib/"lib/typescript.js")],70),("acceptance",["bash",str(repo/"stage3/drivers/tsc/run.sh"),"node",str(lib/"lib/tsc.js")],0)]:
 counts=out/(name+"-counts"); counts.mkdir(exist_ok=True)
 runenv=dict(env,TWO_WRITES_RESULTS=str(counts),TSC_RESULTS=str(out/"acceptance-projects"))
 with (out/(name+".log")).open("w") as log: result=subprocess.run(cmd,cwd=repo,env=runenv,stdout=log,stderr=subprocess.STDOUT)
 assert result.returncode==exit_code,(name,result.returncode)
 totals={site:{key:0 for key in ["assignments","violations"]} for site in ["flags","parent"]}
 for path in counts.glob("*.json"):
  for row in json.loads(path.read_text())["rows"]:
   for key in ["assignments","violations"]: totals[row["site"]][key]+=row[key]
 if name=="acceptance": assert totals=={"flags":{"assignments":61,"violations":0},"parent":{"assignments":0,"violations":0}},totals
 else: assert "write failed:" in (out/(name+".log")).read_text()
 runs.append({"name":name,"exit":result.returncode,"totals":totals,"command":cmd})
report={"measurement":"Node stock TypeScript 6.0.3 with explicit pre-store contracts; no native stock compiler execution", "witness_pin":pin,"changes":changes,"runs":runs}
(out/"report.json").write_text(json.dumps(report,indent=2)+"\n")
print(json.dumps(report,indent=2))
