import pathlib,subprocess,json,time,concurrent.futures
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/internal-lower-interface_cast/round2';(P/'coverage').mkdir(exist_ok=True)
pairs=[('TestDefaultTaggedInterfaceNeedsNoFlag','TestDefaultTaggedInterfaceAdmission'),('TestIteratorGapsAreExplicit','TestIteratorViewsCannotEraseReceivers')];res=[]
def run(name):
 cmd=['timeout','120','go','test','-count=1','-timeout','90s','-coverpkg=github.com/system-inc/adamic/internal/lower','-coverprofile='+str(P/'coverage'/(name+'.cover')),'./internal/lower/','-run','^'+name+'$'];t=time.monotonic()
 with (P/'coverage'/(name+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=R,stdout=f,stderr=subprocess.STDOUT)
 return dict(test=name,command=cmd,exit=r.returncode,seconds=time.monotonic()-t)
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for r in pool.map(run,[n for pair in pairs for n in pair]):res.append(r);(P/'coverage-times.json').write_text(json.dumps(res,indent=2));print(r,flush=True)
def blocks(name):
 return {l.split()[0] for l in (P/'coverage'/(name+'.cover')).read_text().splitlines()[1:] if int(l.split()[2])}
def lines(blocks):
 result=set()
 for b in blocks:
  f,pos=b.rsplit(':',1);a,z=pos.split(',');result.update(f+':'+str(n) for n in range(int(a.split('.')[0]),int(z.split('.')[0])+1))
 return result
out=[]
for a,b in pairs:
 ca,cb=blocks(a),blocks(b);out.append(dict(test=a,subsumer=b,row_only_blocks=sorted(ca-cb),row_only_lines=sorted(lines(ca)-lines(cb))))
(P/'coverage-differences.json').write_text(json.dumps(out,indent=2));print([(r['test'],len(r['row_only_lines'])) for r in out],flush=True)
