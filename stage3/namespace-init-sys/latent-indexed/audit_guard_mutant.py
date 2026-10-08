"""Drop an actual IR panic in a scratch-only overlay; guard evidence must fail."""
import json,subprocess,sys
from pathlib import Path
repo=Path(__file__).resolve().parents[3];baseline=Path(sys.argv[1]).resolve();scratch=Path(sys.argv[2]).resolve();scratch.mkdir(parents=True,exist_ok=True)
overlay=json.loads((baseline/'overlay.json').read_text());source=Path(overlay['Replace'][str(repo/'internal/lower/indexed_checks.go')]).read_text()
old='Panic: ir.StringConstant{Index: l.constant(message)}';assert source.count(old)==1
source=source.replace(old,'Panic: nil').replace('message := "indexed read is absent: " + l.program.Where(node)','_ = "indexed read is absent: " + l.program.Where(node)')
changed=scratch/'indexed.go';changed.write_text(source);overlay['Replace'][str(repo/'internal/lower/indexed_checks.go')]=str(changed);(scratch/'overlay.json').write_text(json.dumps(overlay))
with (scratch/'build.log').open('w') as log:subprocess.run(['go','build','-buildvcs=false','-overlay',str(scratch/'overlay.json'),'-o',str(scratch/'census'),'./stage3/census/latent/tool'],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
with (scratch/'audit.log').open('w') as log:result=subprocess.run(['python3',str(Path(__file__).with_name('audit.py')),str(scratch/'census')],stdout=log,stderr=subprocess.STDOUT)
assert result.returncode!=0 and 'assert len(checks)==1' in (scratch/'audit.log').read_text()
print('actual drop-Panic IR mutant caught by surviving-guard assertion')
