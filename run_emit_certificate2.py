import json,pathlib,subprocess,difflib
out=pathlib.Path(__file__).parent;a=pathlib.Path('/workspace/scratch/tracing-control2');b=pathlib.Path('/workspace/scratch/tracing-after2');names=sorted(str(p.relative_to(a/'built/local')) for p in (a/'built/local').rglob('*.js'));assert len(names)==10;assert names==sorted(str(p.relative_to(b/'built/local')) for p in (b/'built/local').rglob('*.js'))
rows=[];changed=[];target=out/'emit2-certificates';target.mkdir()
for i,name in enumerate(names):
 output=target/(str(i)+'.json');args=['node','--max-old-space-size=3072',str(out/'certify_emit.cjs'),str(a),str(b),name,str(output)]
 with (target/(str(i)+'.log')).open('w') as log:child=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT)
 rows.append(dict(file=name,argv=args,exit=child.returncode));(out/'emit2-certificate-commands.json').write_text(json.dumps(rows,indent=2)+'\n');print(name,'certificate child exit='+str(child.returncode),flush=True)
 if child.returncode:raise SystemExit(child.returncode)
 if (a/'built/local'/name).read_bytes()!=(b/'built/local'/name).read_bytes():changed.append(name)
with (out/'emit2.diff').open('w') as diff:
 for name in changed:diff.writelines(difflib.unified_diff((a/'built/local'/name).read_text().splitlines(True),(b/'built/local'/name).read_text().splitlines(True),fromfile='control/'+name,tofile='after-final/'+name))
certs=[json.loads((target/(str(i)+'.json')).read_text()) for i in range(len(names))]
(out/'emit2-inventory.json').write_text(json.dumps(dict(files=names,changed=changed,note='Exact catch/export reference census is recorded separately with shorthand value-symbol resolution',certificates=certs),indent=2)+'\n')
