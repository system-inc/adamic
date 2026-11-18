"""Resolve class-method targets on a disposable snapshot before the graph query."""
import json
import sys
from pathlib import Path

original, destination = (Path(arg).resolve() for arg in sys.argv[1:])
destination.mkdir(parents=True,exist_ok=True)
overlay=json.loads(original.read_text())
for source, replacement in overlay['Replace'].items():
    if source.endswith('/record_probe.go'):
        old='probe := l.latentSnapshot()'
        text=Path(replacement).read_text()
        assert text.count(old)==1
        helper=destination/'record_probe.go'
        helper.write_text(text.replace(old,old+'\n        probe.finishClassCalls()\n        row["prepared"] = "finishClassCalls"'))
        overlay['Replace'][source]=str(helper)
        break
else:
    raise ValueError('missing record probe file')
(destination/'overlay.json').write_text(json.dumps(overlay))
