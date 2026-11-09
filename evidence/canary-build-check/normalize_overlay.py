import json, pathlib, sys
p=pathlib.Path(sys.argv[1])/'overlay.json'
data=json.loads(p.read_text())
for key,value in data['Replace'].items():
    path=pathlib.Path(value)
    if path.suffix=='.go':
        target=path.with_suffix('.go.txt')
        path.rename(target)
        data['Replace'][key]=str(target)
p.write_text(json.dumps(data,indent=2)+'\n')
