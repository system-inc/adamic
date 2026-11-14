from pathlib import Path
import sys,json,hashlib
repo=Path(sys.argv[1]);tree=Path(sys.argv[2]);p=repo/'stage3/adapt/99-correctness-timer-closures/owners.json';pins=json.loads(p.read_text())
pin=next(x for x in pins if x['owner']=='createDynamicPriorityPollingWatchFile');old=pin['after'];new=old
replacements=[('interface PollingIntervalQueue extends Array<WatchedFileWithUnchangedPolls> {','interface PollingIntervalQueue {\n        files: (WatchedFileWithUnchangedPolls | undefined)[];'),('''        const queue = [] as WatchedFileWithUnchangedPolls[] as PollingIntervalQueue;
        queue.pollingInterval = pollingInterval;
        queue.pollIndex = 0;
        queue.pollScheduled = false;
        return queue;''','''        return { files: [], pollingInterval, pollIndex: 0, pollScheduled: false };'''),('pollQueue(queue, queue.pollingInterval','pollQueue(queue.files, queue.pollingInterval'),('if (queue.length) {','if (queue.files.length) {'),('pollingIntervalQueue(pollingInterval).push(file);','pollingIntervalQueue(pollingInterval).files.push(file);')]
for before,after in replacements:assert new.count(before)==1,before;new=new.replace(before,after)
assert old.count('pollScheduled: unknown;')==new.count('pollScheduled: unknown;')==1
source=tree/pin['file'];text=source.read_text();assert text.count(old)==1
pin['after']=new;pin['ownerAfterSha256']=hashlib.sha256(new.encode()).hexdigest();p.write_text(json.dumps(pins,indent=2)+'\n');source.write_text(text.replace(old,new))
print('one owner, five exact edits; pollScheduled remains unknown')
