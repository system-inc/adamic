"""Extend the guarded latent overlay in scratch, never production lowering."""
import json,sys
from pathlib import Path
root=Path(sys.argv[1]); overlay=Path(sys.argv[2]); evidence=Path(__file__).parent
manifest=overlay/'overlay.json'
data=json.loads(manifest.read_text()); replacements=data['Replace']
def edit(rel,fn):
 p=Path(replacements[str(root/rel)]); p.write_text(fn(p.read_text()))
# The stock measurement discards top-level statement returns. Retain them only in
# the unit result used by this observer so Main is walked as the boundary does.
def hook(s):
 old='\t\t\t\tl := fresh()'
 assert s.count(old)==1
 s=s.replace(old,old+'\n defer func(){ l.boundaryObserve(unit.Where, record.File, len(latentFindings), false) }()')
 # Defer evaluates finding baseline at registration, not when called.
 old='_, err := l.statement(node)'
 assert s.count(old)==1
 s=s.replace(old,'body, err := l.statement(node)\n if err == nil { l.result.Main = append(l.result.Main, body...) }')
 return s
edit('internal/lower/latent_hook.go',hook)
def units(s):
 old='l := base.latentSnapshot()'
 assert s.count(old)==1
 s=s.replace(old,old+'\n boundaryStart := len(latentFindings)\n defer func(){ crash:=recover(); l.boundaryObserve(unit.Where, record.File, boundaryStart, true, crash!=nil); if crash!=nil {panic(crash)} }()')
 old='_, err := l.statement(node)'
 assert s.count(old)==1
 return s.replace(old,'body, err := l.statement(node)\n if err == nil { l.result.Main = append(l.result.Main, body...) }')
edit('internal/lower/latent_units.go',units)
# Correct first-error baseline: capture before lowering, evaluate count at defer.
edit('internal/lower/latent_hook.go',lambda s:s.replace('defer func(){ l.boundaryObserve(unit.Where, record.File, len(latentFindings), false) }()', 'boundaryStart := len(latentFindings)\n defer func(){ crash:=recover(); l.boundaryObserve(unit.Where, record.File, boundaryStart, false, crash!=nil); if crash!=nil {panic(crash)} }()'))
observer=overlay/'internal_lower_graph_boundary_census.go'
observer.write_text((evidence/'counting-hook.go.txt').read_text())
replacements[str(root/'internal/lower/graph_boundary_census.go')]=str(observer)
# Use the meter's tested entry reach: keep entry first, then resolved implementation dependencies.
import subprocess
manifest.write_text(json.dumps(data,indent=2)+'\n')
entry_overlay=overlay/'entry'
subprocess.run([sys.executable,str(root/'stage3/meter/entry_overlay.py'),str(overlay),str(entry_overlay)],check=True)
data=json.loads((entry_overlay/'overlay.json').read_text());replacements=data['Replace']
main_path=Path(replacements[str(root/'stage3/census/latent/tool/main.go')])
s=main_path.read_text()
s=s.replace('func main() {','func main() {\n if os.Getenv("BOUNDARY_SELFTEST") == "1" { lower.BoundaryCensusSelftest(); return }',1)
# The user requests a first-error latent pass even if the checker rejects the
# program. Keep diagnostics; remove only the meter driver's full-mode-only gate.
start=s.index('\tif len(diagnostics) != 0 && os.Getenv("LATENT_FULL") != "1" {')
end=s.index('\tif len(diagnostics) > 0 {',start)
s=s[:start]+s[end:]
s=s.replace('"sources": sources,','"sources": sources, "files": sources,')
main_path.write_text(s)
manifest.write_text(json.dumps(data,indent=2)+'\n')
