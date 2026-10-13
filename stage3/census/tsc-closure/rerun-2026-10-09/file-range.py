"""Scratch scheduling only: keep all roots loaded, emit a lexical file range."""
import argparse,json
from pathlib import Path
p=argparse.ArgumentParser(description=__doc__);p.add_argument('tools',type=Path);p.add_argument('output',type=Path);a=p.parse_args();a.output.mkdir()
source=a.tools/'overlay/internal_lower_latent_units.go'
text=source.read_text();needle='\tfor _, file := range program.Files() {\n\t\tif selection != nil'
assert text.count(needle)==1
text=text.replace('"fmt"','"fmt"\n\t"os"')
text=text.replace(needle,'\tfor _, file := range program.Files() {\n\t\tname := program.FileName(file)\n\t\tif start := os.Getenv("LATENT_FILE_START"); start != "" && name < start { continue }\n\t\tif end := os.Getenv("LATENT_FILE_END"); end != "" && name >= end { continue }\n\t\tif selection != nil')
replacement=a.output/'units.go';replacement.write_text(text)
overlay=json.loads((a.tools/'overlay/overlay.json').read_text())
keys=[k for k,v in overlay['Replace'].items() if Path(v)==source];assert len(keys)==1
overlay['Replace'][keys[0]]=str(replacement.resolve())
(a.output/'overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
print('File scheduling only; loader roots, registration, units and arithmetic unchanged')
