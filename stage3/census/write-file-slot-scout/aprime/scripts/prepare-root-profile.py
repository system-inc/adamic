from pathlib import Path
import json,hashlib
repo=Path('/workspace/scratch/write-file-slot-scout');r=Path('/workspace/scratch/write-file-slot-aprime');original=(repo/'stage3/census/tool/main.go').read_text()
s=original.replace('\tsort.Strings(entries)\n','\tsort.Strings(entries)\n\tif name := os.Getenv("SCOUT_ROOT_MANIFEST"); name != "" {\n\t data, err := os.ReadFile(name); if err != nil { panic(err) }; if err := json.Unmarshal(data, &entries); err != nil { panic(err) }\n\t}\n')
s=s.replace('\tif nonSourceOnly {\n\t\treturn\n\t}', '\tif nonSourceOnly || os.Getenv("SCOUT_ROOT_MANIFEST") != "" {\n\t\treturn\n\t}')
assert s!=original
# The entire production inspect implementation is untouched.
a=original[original.index('func inspect('):original.index('\nfunc main()')];b=s[s.index('func inspect('):s.index('\nfunc main()')];assert a==b
(r/'root-driver.go.txt').write_text(s);(r/'root-overlay.json').write_text(json.dumps({'Replace':{str(repo/'stage3/census/tool/main.go'):str(r/'root-driver.go.txt')}},indent=2)+'\n');(r/'root-driver-receipt.json').write_text(json.dumps({'productionInspectByteIdentical':True,'inspectSha256':hashlib.sha256(a.encode()).hexdigest(),'change':'measurement driver only: select disjoint original root entries from a manifest; omit duplicate whole-program inspection; ordinary Load and Lower unchanged','workers':4},indent=2)+'\n')
