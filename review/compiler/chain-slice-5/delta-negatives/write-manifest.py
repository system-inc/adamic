import json,pathlib,subprocess
root=pathlib.Path.cwd();sha=subprocess.run(['git','rev-parse','HEAD'],check=True,capture_output=True,text=True,timeout=10).stdout.strip();paths=json.loads((root/'review/compiler/chain-slice-5/admission-paths.json').read_text());blobs={}
for line in subprocess.run(['git','ls-tree','-r',sha],check=True,capture_output=True,text=True,timeout=20).stdout.splitlines():
 meta,path=line.split('\t',1);blobs[path]=meta.split()[2]
print(json.dumps({'sha':sha,'corpora':[{'name':'slice-5-census','programs':[{'path':p,'blob':blobs[p]} for p in paths]}]},indent=2))
