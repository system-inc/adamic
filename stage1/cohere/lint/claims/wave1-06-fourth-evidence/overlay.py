"""Compose the existing .a proposal and owned extra tests without shared edits."""
from pathlib import Path
import json
import subprocess
root = Path(__file__).resolve().parents[5]
evidence = Path(__file__).resolve().parent
result = subprocess.run(['python3', str(root / 'stage1/cohere/lint/claims/wave1-06-next-evidence/propose-a-support.py')], check=True, capture_output=True, text=True)
path = Path(result.stdout.strip())
data = json.loads(path.read_text())
data['Replace'][str(root / 'stage1/cohere/lint/wave06_candidate_test.go')] = str(evidence / 'candidate_test.go.txt')
path.write_text(json.dumps(data))
print(path)
