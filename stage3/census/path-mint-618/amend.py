import json,hashlib,pathlib,difflib
root=pathlib.Path('/workspace/scratch/path-mint-618/work')
f=root/'stage3/adapt/99-correctness-timer-closures/owners.json'
original=f.read_text(); pins=json.loads(original)
pin=next(p for p in pins if p['owner']=='createDirectoryWatcherSupportingRecursive')
before=pin['after']; after=before
edits=[
('createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path','createGetCanonicalFileName(useCaseSensitiveFileNames)'),
('const dirPath = toCanonicalFilePath(dirName);','const dirPath = (toCanonicalFilePath(dirName) as Path);'),
('invokeCallbacks(link, toCanonicalFilePath(link), invokeMap,','invokeCallbacks(link, (toCanonicalFilePath(link) as Path), invokeMap,'),
('invokeCallbacks(link, toCanonicalFilePath(link), toPathInLink(fileName!));','invokeCallbacks(link, (toCanonicalFilePath(link) as Path), toPathInLink(fileName!));'),
('cache.get(toCanonicalFilePath(childWatcher.dirName))','cache.get((toCanonicalFilePath(childWatcher.dirName) as Path))')]
for a,b in edits:
 assert after.count(a)==1,a
 after=after.replace(a,b)
oldhash=pin['ownerAfterSha256']; newhash=hashlib.sha256(after.encode()).hexdigest()
# Preserve all unrelated JSON bytes, and update exactly the chosen after string.
old=json.dumps(before,ensure_ascii=False); new=json.dumps(after,ensure_ascii=False)
assert original.count(old)==1
assert original.count(oldhash)==1
f.write_text(original.replace(old,new).replace(oldhash,newhash))
e=pathlib.Path('/workspace/scratch/path-mint-618/evidence')
(e/'owner.patch').write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='main-owner',tofile='amended-owner')))
(e/'owner-hashes.json').write_text(json.dumps({'beforePin':pin['ownerBeforeSha256'],'mainAfterPin':oldhash,'amendedAfterPin':newhash},indent=2)+'\n')
print(newhash)
