from pathlib import Path
import shutil,json
root=Path(__file__).resolve().parent
# Supply the scratch parent as the first argument when running from this evidence directory.
import sys
if len(sys.argv)>1:root=Path(sys.argv[1]).resolve()
base=root/'baseline';candidate=root/'candidate'
candidate.mkdir();shutil.copytree(base/'src',candidate/'src')
changes=[]
for rel in ['src/compiler/sys.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts','src/harness/watchUtils.ts']:
 p=candidate/rel;b=p.read_bytes();a=b.replace(b'recursive: boolean',b'recursive: boolean | undefined');p.write_bytes(a)
 changes.append({'file':rel,'count':b.count(b'recursive: boolean')})
assert [r['count'] for r in changes]==[9,1,2]
(root/'experiment.json').write_text(json.dumps({'change':'widen the result alias and every reached recursive parameter contract, type-only','files':changes},indent=2)+'\n')
