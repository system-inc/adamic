#!/usr/bin/env python3
"""Extend only post-refusal records; keep the original unit hook and real guard invocation."""
import json,sys,shutil
from pathlib import Path
root,overlay=map(lambda s:Path(s).resolve(),sys.argv[1:3]);evidence=Path(__file__).resolve().parent
manifest=overlay/'overlay.json';data=json.loads(manifest.read_text());p=Path(data['Replace'][str(root/'internal/lower/roads_guard_hook.go')]);s=p.read_text()
old='case "an arm","a source arm":r="union arms"';assert s.count(old)==1;s=s.replace(old,'case "an arm","a source arm","a split arm":r="union arms"')
old='steps:=[]map[string]string{}';assert s.count(old)==1;s=s.replace(old,old+';trace:=[]viewPairStep{}')
old='current:=source';assert s.count(old)==1;s=s.replace(old,'trace=pairFailure.through;'+old)
old='row["road"],row["roads"],row["steps"]=road,roads,steps;emit(row)';assert s.count(old)==1
s=s.replace(old,'row["road"],row["roads"],row["steps"]=road,roads,steps;row["diagnostics"]=func()(result map[string]any){defer func(){if p:=recover();p!=nil{result=map[string]any{"error":fmt.Sprint(p)}}}();return f.roadsGuardDiagnostics(source,target,trace)}();emit(row)')
p.write_text(s)
destination=overlay/'roads_guard_diagnostics.go';shutil.copyfile(evidence/'diagnostics.go.txt',destination);data['Replace'][str(root/'internal/lower/roads_guard_diagnostics.go')]=str(destination)
manifest.write_text(json.dumps(data,indent=2)+'\n');print('post-refusal diagnostic extension installed')
