from pathlib import Path
import subprocess
out=Path('review/compiler/miscompile-train-plus/count-attribution');out.mkdir(exist_ok=True)
paths=['internal/oracle/testdata/review/agree/fxspptb_oct9_native_p04_undefined_slot_write.a']+['internal/lower/testdata/scalar_union_views/p54_'+n+'.a' for n in ['n','b','s']]+['stage3/interface-downcasts/v2/snapshot-maybe-'+n+'.a' for n in ['boolean','number']]+['stage3/interface-downcasts/v2/undefined-read-write.a']
for p in paths:
 name=Path(p).stem
 with (out/(name+'-train.c.txt')).open('w') as f:
  subprocess.run(['/tmp/train-plus-baseline-adamic','c',p],stdout=f,stderr=subprocess.STDOUT,timeout=30,check=True)
 with (out/(name+'-train-build.log')).open('w') as f:
  subprocess.run(['/tmp/train-plus-baseline-adamic','build',p,'-o','/tmp/train-plus-count-'+name,'--count'],stdout=f,stderr=subprocess.STDOUT,timeout=90,check=True)
 with (out/(name+'-train-count.log')).open('w') as f:
  subprocess.run(['/tmp/train-plus-count-'+name],stdout=f,stderr=subprocess.STDOUT,timeout=30,check=True)
 print(name+': measured train',flush=True)
