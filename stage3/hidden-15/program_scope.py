"""Filter census attempts to program.ts while retaining full project loading."""
import json
from pathlib import Path
import sys
scratch = Path(sys.argv[1]).resolve()
compiler = Path(sys.argv[2]).resolve()
source = scratch / "internal_lower_latent_units.go"
text = source.read_text()
needle = "for _, file := range program.Files() {"
assert text.count(needle) == 1
# Go string literals use the same escaping here as JSON string literals.
text = text.replace(needle, needle + "\n if program.FileName(file) != "
                    + json.dumps(str(compiler / "program.ts")) + " { continue }")
filtered = scratch / "program_latent_units.go"
filtered.write_text(text)
mapping = json.loads((scratch / "overlay.json").read_text())
for original, replacement in mapping["Replace"].items():
    if Path(replacement) == source:
        mapping["Replace"][original] = str(filtered)
        break
else:
    raise ValueError("census units overlay missing")
(scratch / "program-overlay.json").write_text(json.dumps(mapping, indent=2) + "\n")
