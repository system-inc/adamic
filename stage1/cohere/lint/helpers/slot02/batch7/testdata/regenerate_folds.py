import pathlib, json, subprocess, tempfile
here=pathlib.Path(__file__).resolve().parent
with tempfile.TemporaryDirectory() as scratch:
 out=pathlib.Path(scratch)/'folds.json'
 subprocess.run(['go','run',str(here/'folds.go'),str(out)],check=True)
 data=json.loads(out.read_text())
 target=here.parent/'match_ignoring_case.a'
 source=target.read_text().split('// BEGIN GENERATED SIMPLE FOLD TABLE')[0]
 source+='// BEGIN GENERATED SIMPLE FOLD TABLE\n// Go Unicode '+data['Version']+'. Regenerate with testdata/regenerate_folds.py.\nconst foldPairs: readonly number[] = [\n'
 p=data['Pairs']
 for i in range(0,len(p),2):source+=f'    {p[i]}, {p[i+1]},\n'
 target.write_text(source+'];\n')
 (here/'folds.json').write_text(json.dumps(data,separators=(',',':'))+'\n')
 print('Unicode',data['Version'],'noncanonical scalar mappings',len(p)//2)
