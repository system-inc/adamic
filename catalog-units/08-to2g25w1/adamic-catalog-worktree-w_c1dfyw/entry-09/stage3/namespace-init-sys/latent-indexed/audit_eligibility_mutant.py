"""Scheduled production contracts must not become latent checker-body skips."""
import json,subprocess,sys
from pathlib import Path
repo=Path(__file__).resolve().parents[3];baseline=Path(sys.argv[1]).resolve();scratch=Path(sys.argv[2]).resolve();scratch.mkdir(parents=True,exist_ok=True)
overlay=json.loads((baseline/'overlay.json').read_text());key=str(repo/'internal/load/latent_hook.go');source=Path(overlay['Replace'][key]).read_text()
old='row.Site.Message == message {\n\t\t\treturn true';assert source.count(old)==1
changed=scratch/'load.go';changed.write_text(source.replace(old,'row.Site.Message == message {\n\t\t\treturn false'))
overlay['Replace'][key]=str(changed);(scratch/'overlay.json').write_text(json.dumps(overlay))
with (scratch/'build.log').open('w') as log:subprocess.run(['go','build','-buildvcs=false','-overlay',str(scratch/'overlay.json'),'-o',str(scratch/'census'),'./stage3/census/latent/tool'],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
with (scratch/'audit.log').open('w') as log:result=subprocess.run(['python3',str(Path(__file__).with_name('audit.py')),str(scratch/'census')],stdout=log,stderr=subprocess.STDOUT)
assert result.returncode!=0 and "assert not header['diagnostic_sites']" in (scratch/'audit.log').read_text()
print('scheduled-as-checker-error mutant caught by production eligibility assertion')
