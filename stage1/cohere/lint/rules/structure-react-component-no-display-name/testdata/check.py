"""Select owned parity or the owned sanitized native mutant without shared edits."""
import json
import subprocess
import sys
import tempfile
from pathlib import Path
ROOT = Path(__file__).resolve().parents[6]
SOURCE = ROOT / "stage1/cohere/lint/lint_test.go"
with tempfile.TemporaryDirectory(prefix="display-name-check-") as scratch:
    scratch = Path(scratch)
    text = SOURCE.read_text()
    if len(sys.argv) > 1 and sys.argv[1] == "mutant":
        old = 'const nativeCanaryRule = "react/jsx-no-comment-textnodes"'
        assert text.count(old) == 1
        text = text.replace(old, 'const nativeCanaryRule = "structure/react-component-no-display-name"')
        selected = "^TestMutants/display-name-wrapper-exemption-lost$"
    else:
        text += "\n" + Path(__file__).with_name("selected_test.go.txt").read_text()
        selected = "^TestStructureReactComponentNoDisplayNamePort$"
    patched = scratch / "lint_test.go"
    patched.write_text(text)
    overlay = scratch / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(SOURCE):str(patched)}}))
    subprocess.run(["go","test","-overlay="+str(overlay),"./stage1/cohere/lint","-run",selected,"-count=1","-v","-timeout=60m"], cwd=ROOT, check=True)
