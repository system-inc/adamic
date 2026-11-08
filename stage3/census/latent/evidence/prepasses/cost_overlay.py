"""Make a cost-only sibling of the changed, guarded entry overlay."""
import json
from pathlib import Path
import sys
source, output = map(Path, sys.argv[1:])
output.mkdir(parents=True, exist_ok=True)
metadata = json.loads((source / 'overlay.json').read_text())
for target, path in list(metadata['Replace'].items()):
    if target.endswith('/internal/lower/latent_hook.go'):
        text = Path(path).read_text() + Path(__file__).with_name('cost-driver.go.txt').read_text()
    elif target.endswith('/stage3/census/latent/tool/main.go'):
        text = Path(path).read_text()
        needle = 'lower.Latent(context.Background(), program, func(record lower.LatentFile) {'
        assert text.count(needle) == 1
        text = text.replace(needle, 'lower.LatentPrepassCostProbe(context.Background(), program)\n return\n '+needle)
    else:
        continue
    destination = output / Path(path).name
    destination.write_text(text)
    metadata['Replace'][target] = str(destination.resolve())
(output / 'overlay.json').write_text(json.dumps(metadata, indent=2)+'\n')
