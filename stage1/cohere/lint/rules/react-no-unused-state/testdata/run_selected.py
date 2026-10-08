from pathlib import Path
import json,subprocess,tempfile
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[5];LINT=ROOT/'stage1/cohere/lint'
with tempfile.TemporaryDirectory() as tmp:
 overlay=Path(tmp)/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(LINT/'selected_react_no_unused_state_test.go'):str(HERE/'selected_test.go')}}))
 for pattern in ['^TestReactNoUnusedStateSelected$','^TestMutants/unused_state_read_inverted$']:
  subprocess.run(['go','test','-overlay='+str(overlay),'./stage1/cohere/lint','-run',pattern,'-count=1','-v','-timeout=20m'],cwd=ROOT,check=True)
