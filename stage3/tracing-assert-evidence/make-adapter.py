from pathlib import Path
import re,json,hashlib
root=Path('/workspace/adamic/stage3/adapt/99-correctness-tracing-assert')
text=Path('/workspace/typescript-6.0.3/src/compiler/tracing.ts').read_bytes().decode()
before=re.search(r'export function stopTracing\(\): void \{.*?\r\n    \}',text,re.S).group().replace('legend[legend.length - 1].typesPath','legend[legend.length - 1]!.typesPath')
after=before.replace('Debug.assert(tracing,','Debug.assert(!!tracing,')
root.joinpath('pins.json').write_text(json.dumps(dict(before=before,after=after,beforeSha256=hashlib.sha256(before.encode()).hexdigest(),afterSha256=hashlib.sha256(after.encode()).hexdigest()),indent=2)+'\n')
