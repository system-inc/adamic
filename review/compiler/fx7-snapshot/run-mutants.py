from pathlib import Path
import subprocess
p=Path("internal/native/runtime/object.c"); q=Path("internal/native/view_unions_read.go")
r=Path("review/compiler/fx7-snapshot")
a=(r/"fixed-object.c.txt").read_text(); b=(r/"fixed-reader.go.txt").read_text()
olda=subprocess.check_output(["git","show","cf735d9fba9e38de6368575e5630e44375a86eaf:internal/native/runtime/object.c"],text=True)
oldb=subprocess.check_output(["git","show","cf735d9fba9e38de6368575e5630e44375a86eaf:internal/native/view_unions_read.go"],text=True)
cases=[("runtime-alone",a,oldb,0),("compiler-alone",olda,b,0),("runtime-mutant",a.replace("adamic_maybe_boolean boolean =", "value.payload = *slot;\n        adamic_maybe_boolean boolean ="),oldb,1),("compiler-mutant",olda,oldb,1)]
try:
 for name,aa,bb,want in cases:
  p.write_text(aa);q.write_text(bb)
  with (r/(name+".log")).open("w") as log:
   result=subprocess.run(["timeout","90","go","test","./internal/oracle","-run","^TestViewSnapshotMaybeBoolean$","-v","-count=1","-timeout","60s"],stdout=log,stderr=subprocess.STDOUT)
  print(name,"exit",result.returncode,"expected",want,flush=True)
  if result.returncode!=want: raise SystemExit("unexpected mutant result")
finally:
 p.write_text(a);q.write_text(b)
