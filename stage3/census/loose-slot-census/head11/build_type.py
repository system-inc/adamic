from pathlib import Path
import subprocess,json
s=Path('/workspace/loose-slot-scratch/eleven');r=s/'head';out=Path(__file__).resolve().parent
p=s/'type-overlay.json';p.write_text(json.dumps({'Replace':{str(r/'internal/lower/loose_census_scratch.go'):str(out/'type_count.go.txt'),str(r/'stage3/census/tool/main.go'):str(out/'type_main.go.txt')}}))
subprocess.run(['go','build','-overlay',str(p),'-o',str(s/'type-count'),'./stage3/census/tool'],cwd=r,check=True)
