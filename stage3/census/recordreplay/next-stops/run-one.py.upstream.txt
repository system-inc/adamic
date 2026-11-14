"""Schedule one unchanged ordinary census root; keep the scheduler below 30 minutes too."""
import json,sys
from pathlib import Path
from run import run,out,tree,scratch
roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79,len(roots)
(out/'roots.json').write_text(json.dumps(list(map(str,roots)),indent=2)+'\n')
index=int(sys.argv[1]);assert 1<=index<=79
root=roots[index-1];dest=scratch/f'next-stops-8e3a784a-root-{index:02}.jsonl'
run(f'root-{index:02}',[str(scratch/'next-stops-8e3a784a-census'),str(root),str(dest)])
data=dest.read_bytes();observations=[json.loads(line) for line in data.splitlines()]
assert len(observations)==2
assert {k:v for k,v in observations[0].items() if k!='seconds'}=={k:v for k,v in observations[1].items() if k!='seconds'}
(out/f'root-{index:02}.jsonl').write_bytes(data)
print(index,root.name,observations[0]['kind'],observations[0].get('where','-'),observations[0].get('reason','-'),flush=True)
