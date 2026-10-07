#!/usr/bin/env python3
"""Parking oracle: current compiler, frozen dependencies, actual stage1 corpus.
Shared repository files are neither changed nor generated. The dependency
snapshot is the lint tree at 661b9e04, before this worker's owned ports.
"""
import io, os, shutil, subprocess, tarfile, tempfile
from pathlib import Path
owned=Path(__file__).resolve().parent
root=owned.parents[4]
scratch=Path(tempfile.mkdtemp(prefix="wave14-parking-"))
evidence=owned/"evidence"
evidence.mkdir(exist_ok=True)
print("isolated checkout="+str(scratch),flush=True)
archive=subprocess.check_output(["git","archive","HEAD"],cwd=root)
with tarfile.open(fileobj=io.BytesIO(archive)) as tree:
 tree.extractall(scratch,filter="data")
with tarfile.open(owned/"parking-foundation.tar.gz") as tree:
 tree.extractall(scratch,filter="data")
# Include uncommitted owned verification changes without replacing shared files.
for directory in (root/"stage1/cohere/lint/rules").iterdir():
 if directory.is_dir():shutil.copytree(directory,scratch/"stage1/cohere/lint/rules"/directory.name,dirs_exist_ok=True)
cohere=scratch/"cohere"
if cohere.exists():cohere.rmdir()
cohere.symlink_to(root/"cohere",target_is_directory=True)
# Canonical shim paths reuse setup caches and preserve Go internal boundaries.
module=scratch/"go.mod"
module.write_text(module.read_text().replace("=> ./cohere/", "=> "+str((root/"cohere").resolve())+"/"))
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE=os.environ.get("ADAMIC_TYPESCRIPT_SOURCE","/tmp/lint-wave1-14-typescript-pinned"),ADAMIC_STAGE1_CORPUS=str(root/"stage1"))
rules=scratch/"stage1/cohere/lint/rules"
def check(relative,args=()):
 label=relative.replace("/","-")
 with (evidence/("parking-"+label+".log")).open("w") as log:
  result=subprocess.run(["python3",rules/relative,*map(str,args)],cwd=scratch,env=env,stdout=log,stderr=subprocess.STDOUT)
 script_directory=Path(relative).parent
 names={"validate.py":["validation.log"],"validate_complete.py":["complete.log","upstream-complete.log"],"validate_backlog.py":["backlog.log","upstream-backlog.log"],"validate_edges.py":["edges.log"],"validate_boundaries.py":["boundaries.log"]}[Path(relative).name]
 for name in names:
  file=rules/script_directory/"evidence"/name
  shutil.copyfile(file,evidence/("parking-"+script_directory.name+"-"+file.name))
 print(relative+" exit="+str(result.returncode),flush=True)
 if result.returncode:raise SystemExit(result.returncode)
selection=os.environ.get("ADAMIC_PARKING_START","")
started=not selection
for relative in ["nexus-import-require-module-alias/validate.py","typescript-eslint-no-duplicate-enum-values/validate.py","typescript-eslint-no-misused-new/validate.py","typescript-eslint-prefer-as-const/validate_complete.py","typescript-eslint-prefer-as-const/validate_backlog.py","no-lonely-if/validate_complete.py"]:
 if relative==selection:started=True
 if started:check(relative)
last=(rules/"no-lonely-if/evidence/complete.log").read_text().splitlines()[0].split("=",1)[1]
check("no-loss-of-precision/validate_edges.py",[last])
check("typescript-eslint-prefer-as-const/validate_boundaries.py")
print("PASS parking comparison on Node, emitted JavaScript and sanitized native with current compiler",flush=True)
