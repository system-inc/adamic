import pathlib,json,hashlib,difflib
s=pathlib.Path('/workspace/scratch/path-mint-618');a=s/'base';b=s/'lane/adapted-tree';e=s/'evidence'
def manifest(root):return {str(f.relative_to(root)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((root/'src').rglob('*')) if f.is_file()}
x=manifest(a);y=manifest(b);changed=[f for f in sorted(x.keys()&y.keys()) if x[f]!=y[f]]
r={'baseFiles':len(x),'afterFiles':len(y),'new':sorted(y.keys()-x.keys()),'gone':sorted(x.keys()-y.keys()),'changed':changed};assert not r['new'] and not r['gone'] and changed==['src/compiler/sys.ts'],r
pins=json.loads((s/'base-owners.json').read_text());p=next(p for p in pins if p['owner']=='createDirectoryWatcherSupportingRecursive');q=next(p for p in json.loads((s/'work/stage3/adapt/99-correctness-timer-closures/owners.json').read_text()) if p['owner']=='createDirectoryWatcherSupportingRecursive')
old=(a/'src/compiler/sys.ts').read_text();new=(b/'src/compiler/sys.ts').read_text();assert old.count(p['after'])==new.count(q['after'])==1
assert old.replace(p['after'],q['after'])==new
raw=(a/'src/compiler/sys.ts').read_bytes(); afterraw=(b/'src/compiler/sys.ts').read_bytes();ending='\r\n' if b'\r\n' in raw else '\n'
oldowner=p['after'].replace('\n',ending).encode();newowner=q['after'].replace('\n',ending).encode();assert raw.count(oldowner)==1 and raw.replace(oldowner,newowner)==afterraw
r['onlyFiveEditsIncludingLineEndings']=True
(e/'source-comparison.json').write_text(json.dumps(r,indent=2)+'\n');(e/'source.patch').write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='main/src/compiler/sys.ts',tofile='amendment/src/compiler/sys.ts')));print(r)
