"""The required nothing-else-new gate; deliberately red for the recorded stop."""
import json,pathlib,sys
root=pathlib.Path('/workspace/scratch/sys-timer-stack')
result=json.loads((root/'comparison.json').read_text())['head2']
print(json.dumps({'pass':result['pass'],'newSites':len(result['new']),'unexpected':result['unexpected'],'stoppingReason':result['stoppingReason']},indent=2))
sys.exit(0 if result['pass'] else 1)
