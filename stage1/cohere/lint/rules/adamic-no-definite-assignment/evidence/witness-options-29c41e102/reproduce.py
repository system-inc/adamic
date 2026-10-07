"""Run the retained raw source through the shared parser without treating refusal as parity."""
from pathlib import Path
import json,subprocess,tempfile
here=Path(__file__).resolve().parent
repo=next(p for p in here.parents if (p/'go.mod').exists())
subprocess.run(['go','run','./cmd/lint-registry'],cwd=repo,check=True)
with tempfile.TemporaryDirectory(prefix='wave11-decorated-class-') as directory:
 scratch=Path(directory);source=scratch/'SecurityRequireContextAccess.ts'
 source.write_bytes((here/'decorated-class.ts.txt').read_bytes())
 options=json.dumps(json.loads((here/'decorated-class.options.json').read_text()),separators=(',',':'))
 manifest=scratch/'manifest.txt'
 manifest.write_text(str(source)+'\tbase/security-require-context-access\t\t\tfalse\t'+options+'\n')
 result=subprocess.run(['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(repo/'stage1/cohere/lint/main.ts'),'--manifest',str(manifest)],cwd=repo)
 raise SystemExit(result.returncode)
