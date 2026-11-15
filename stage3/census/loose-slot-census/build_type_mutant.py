"""Disable the head's actual reach judgment; retain the active collector."""
from pathlib import Path
import json,subprocess
s=Path('/workspace/loose-slot-scratch');r=s/'head';d=r/'stage3/census/loose-slot-census'
source=(r/'internal/lower/loose_slots.go').read_text()
signature='func (f *cycleFinder) looseReaches(value *checker.Type, target cycleNode) bool {'
start=source.index(signature);end=source.index('\n// looseScalar',start)
mutant=s/'mutant-loose-slots.go';mutant.write_text(source[:start]+signature+'\n return false\n}\n'+source[end:])
overlay=json.loads((s/'type-overlay.json').read_text());overlay['Replace'][str(r/'internal/lower/loose_slots.go')]=str(mutant)
p=s/'type-mutant-overlay.json';p.write_text(json.dumps(overlay,indent=2)+'\n')
subprocess.run(['go','build','-overlay',str(p),'-o',str(s/'type-count-mutant'),'./stage3/census/tool'],cwd=r,check=True)
