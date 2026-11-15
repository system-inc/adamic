from pathlib import Path
import json,subprocess
s=Path('/workspace/loose-slot-scratch');r=s/'head';d=r/'stage3/census/loose-slot-census'
o=s/'type-overlay.json';o.write_text(json.dumps({'Replace':{str(r/'internal/lower/loose_census_scratch.go'):str(d/'type_count.go.txt'),str(r/'stage3/census/tool/main.go'):str(d/'type_main.go.txt')}}))
subprocess.run(['go','build','-overlay',str(o),'-o',str(s/'type-count'),'./stage3/census/tool'],cwd=r,check=True)
