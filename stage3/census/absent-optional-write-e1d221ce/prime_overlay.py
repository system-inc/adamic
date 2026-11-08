"""Restore production optional-slot discovery in first-error measurement only."""
import json,sys
from pathlib import Path
scratch=Path(sys.argv[1])
metadata=json.loads((scratch/'overlay.json').read_text())
logical=[p for p in metadata['Replace'] if p.endswith('/internal/lower/latent_hook.go')]
assert len(logical)==1
p=Path(metadata['Replace'][logical[0]])
s=p.read_text()
anchor='\t\t\t\tl.noteAccessorNames(program.Files())\n'
assert s.count(anchor)==1
assert 'l.noteOmittedOptionals(program.Files())' not in s
s=s.replace(anchor,anchor+'\t\t\t\tl.noteOmittedOptionals(program.Files())\n')
p.write_text(s)
print('first-error production optional-slot prepass installed; no production source modified')
