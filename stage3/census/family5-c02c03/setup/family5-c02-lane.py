from pathlib import Path
import time,json,subprocess,sys
p=Path('/tmp/family5-evidence/c02/whole-apply/status.json')
while not p.exists():time.sleep(5)
if json.loads(p.read_text())['exit_code']:sys.exit(1)
sys.exit(subprocess.call([sys.executable,'/tmp/family5-proof.py','/tmp/family5-evidence/c02/whole-lane','bash','stage3/lane/run.sh','/workspace/family5-c02-lane']))
