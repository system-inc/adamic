import json
from pathlib import Path
source=Path('/tmp/scout-entry-overlay/overlay.json'); overlay=json.loads(source.read_text())
key=next(k for k in overlay['Replace'] if k.endswith('/internal/lower/latent_units.go'))
text=Path(overlay['Replace'][key]).read_text()
needle='func latentFullSelected(ctx context.Context, program *load.Program, emit func(LatentFile), selection *latentReplaySelection) {\n\tfor _, file := range program.Files() {'
assert text.count(needle)==1
text=text.replace('import (\n','import (\n "os"\n',1)
text=text.replace(needle,needle+'\n if selected := os.Getenv("SCOUT_FILE"); selected != "" && program.FileName(file) != selected { continue }')
Path('/tmp/scout-file-units.go').write_text(text)
overlay['Replace'][key]='/tmp/scout-file-units.go'
Path('/tmp/scout-file-overlay.json').write_text(json.dumps(overlay))
rows=[json.loads(line) for line in Path('/tmp/scout-full.jsonl').read_text().splitlines()]
header=rows[0]; completed={r['file'] for r in rows[1:]}
assert len(completed)==len(rows)-1
remaining=sorted(set(header['sources'])-completed)
Path('/tmp/scout-remaining.json').write_text(json.dumps(remaining))
Path('/tmp/scout-census-parts').mkdir(exist_ok=True)
print(len(completed),'completed;',len(remaining),'remaining;',len(header['sources']),'resolved sources')
