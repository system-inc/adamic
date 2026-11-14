import hashlib,json
from pathlib import Path
root=Path('/workspace/scratch/changed-files-undefined-1')
p=root/'stage3/adapt/99-correctness-timer-closures/owners.json'
raw=p.read_text(); pins=json.loads(raw); pin=pins[0]
assert pin['owner']=='createDynamicPriorityPollingWatchFile'
old='const changedFilesInLastPoll: WatchedFileWithUnchangedPolls[] = [];'
new='const changedFilesInLastPoll: (WatchedFileWithUnchangedPolls | undefined)[] = [];'
assert pin['after'].count(old)==1
assert hashlib.sha256(pin['after'].encode()).hexdigest()==pin['ownerAfterSha256']
after=pin['after'].replace(old,new)
# Preserve all other JSON bytes, including the upstream before pin.
old_json=json.dumps(pin['after']); new_json=json.dumps(after)
assert raw.count(old_json)==1
raw=raw.replace(old_json,new_json)
raw=raw.replace('"ownerAfterSha256": "'+pin['ownerAfterSha256']+'"','"ownerAfterSha256": "'+hashlib.sha256(after.encode()).hexdigest()+'"',1)
p.write_text(raw)
updated=json.loads(raw)
assert updated[1:]==pins[1:]
assert updated[0]['before']==pin['before']
print(hashlib.sha256(after.encode()).hexdigest())
