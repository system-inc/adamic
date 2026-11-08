"""Archive generated overlay text with .go.txt evidence filenames."""
import hashlib,json
from pathlib import Path
out=Path(__file__).resolve().parent
for side in ['parent','head']:
 for mode,source in [('base',Path('/tmp/branded-'+side+'-overlay/overlay.json')),('entry',Path('/tmp/branded-'+side+'-entry/overlay.json')),('focus',Path('/tmp/branded-'+side+'-focus-overlay.json'))]:
  dest=out/'overlays'/side/mode;dest.mkdir(parents=True,exist_ok=True)
  rows=[]
  for i,(logical,actual) in enumerate(sorted(json.loads(source.read_text())['Replace'].items())):
   name=str(i).zfill(2)+'-'+Path(actual).name
   if name.endswith('.go'):name+='.txt'
   data=Path(actual).read_bytes();(dest/name).write_bytes(data)
   rows.append(dict(logical=logical,evidence=str((dest/name).relative_to(out)),sha256=hashlib.sha256(data).hexdigest()))
  (dest/'manifest.json').write_text(json.dumps(rows,indent=2)+'\n')
