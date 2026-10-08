"""Expose main's castProof only to the measurement binary, never to an output path."""
import json
from pathlib import Path
import sys

root=Path(__file__).resolve().parents[2]
out=Path(sys.argv[1]).resolve();out.mkdir(exist_ok=True,parents=True)
loader=(root/'internal/load/load.go').read_text()
needle='return nil, &CheckError{Diagnostics: diagnostics}'
assert loader.count(needle)==1
loader=loader.replace(needle,'_ = diagnostics // measurement only: preserve the populated checker')
loader=loader.replace('func Load(paths []string)', 'func Step09Load(paths []string)')
loader+='\nfunc Load(paths []string) (*Program,error) {return nil, fmt.Errorf("step09 measurement only; no output path")}\n'
source=out/'load.go';source.write_text(loader)
replace={str(root/'internal/load/load.go'):str(source),str(root/'internal/lower/step09_cast_probe.go'):str(root/'stage3/step09-ledger/cast_probe.go.txt'),str(root/'stage3/step09-ledger/probe/main.go'):str(root/'stage3/step09-ledger/probe/main.go.txt')}
(out/'overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
