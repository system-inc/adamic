from pathlib import Path
import json,subprocess,tempfile
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[5];LINT=ROOT/'stage1/cohere/lint'
with tempfile.TemporaryDirectory() as tmp:
 replace={str(LINT/'selected_base_security_require_context_access_test.go'):str(HERE/'selected_test.go')}
 # wave2-03's testguard migration left an unused import in the shared Go harness.
 # Compile a temporary import-only correction; no shared or parser file is edited.
 shared=LINT/'shared_test.go';text=shared.read_text()
 if '\t"context"\n' in text and 'context.' not in text:
  side=Path(tmp)/'shared_test.go';side.write_text(text.replace('\t"context"\n','',1));replace[str(shared)]=str(side)
 overlay=Path(tmp)/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}))
 for pattern in ['^TestBaseSecurityRequireContextAccessSelected$','^TestMutants/context_access_alias_lost$']:
  subprocess.run(['go','test','-overlay='+str(overlay),'./stage1/cohere/lint','-run',pattern,'-count=1','-v','-timeout=20m'],cwd=ROOT,check=True)
