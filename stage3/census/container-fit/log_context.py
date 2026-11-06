"""Mark scan boundaries so setup/scan walks cannot inherit the previous unit."""
import hashlib,json,pathlib,subprocess
O=pathlib.Path(__file__).resolve().parent;D=pathlib.Path('/workspace/container-fit-build-v2/log');p=O/'tooling/first-error-piece.go.txt'
s=p.read_text();old='if found && unit != "" && (selected.AllUnits || anyMembersCensusUnits[file][unit]) {';new='if found && (unit == "" || selected.AllUnits || anyMembersCensusUnits[file][unit]) {'
assert s.count(old)==1;s=s.replace(old,new)
old='f.WriteString("unit "+unit+"\\n"); f.Close()';new='if unit == "" { f.WriteString("scan "+file+"\\n") } else { f.WriteString("unit "+unit+"\\n") }; f.Close()'
assert s.count(old)==1;s=s.replace(old,new);p.write_text(s)
helper=D/'first-error-piece.go.txt';helper.write_text(s)
with (O/'evidence/log-context-build.log').open('w') as f:
 subprocess.run(['gofmt','-w',str(helper)],check=True,stdout=f,stderr=subprocess.STDOUT)
 subprocess.run(['go','build','-buildvcs=false','-overlay='+str(D/'overlay.json'),'-o',str(D/'latent'),'./stage3/census/latent/tool'],cwd='/workspace/container-fit-log',check=True,stdout=f,stderr=subprocess.STDOUT)
meta=O/'evidence/latent-build/log.json';v=json.loads(meta.read_text());v['previous_selector_sha256']=v['selector_sha256'];v['selector_sha256']=hashlib.sha256((D/'latent').read_bytes()).hexdigest();v['helper_sha256']=hashlib.sha256(helper.read_bytes()).hexdigest();v['log_context']='scan markers reset unit attribution; no lowering change';meta.write_text(json.dumps(v,indent=2)+'\n')
p=O/'latent_report.py';s=p.read_text().replace("if line.startswith('unit '):owner=line[5:];continue", "if line.startswith('scan '):owner=None;continue\n   if line.startswith('unit '):owner=line[5:];continue");compile(s,str(p),'exec');p.write_text(s)
