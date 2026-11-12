import json,pathlib,sys
root=pathlib.Path('/workspace/ldpi'); repo=pathlib.Path('/workspace/latent-deferred-probe-isolation')
kind=sys.argv[1]; d=json.loads((root/f'{kind}-entry/overlay.json').read_text()); key=str(repo/'internal/lower/latent_hook.go')
s=pathlib.Path(d['Replace'][key]).read_text()
needle='for _, file := range program.Files() {'
assert s.count(needle)==1
s=s.replace(needle,needle+'\n if !latentFileSelected(program.FileName(file)) { continue }',1)
needle='witnesses := latentFormatInstantiations()'
assert s.count(needle)==1
s=s.replace(needle,needle+'''
 if path := os.Getenv("LDPI_INSTANTIATIONS"); path != "" {
  data, err := json.Marshal(map[string]any{"entered":latentInstantiated,"witnesses":witnesses})
  if err != nil { panic(err) }; if err := os.WriteFile(path,data,0600); err != nil { panic(err) }
 }
''',1)
out=root/f'{kind}-split';out.mkdir(exist_ok=True)
p=out/'hook.go.txt';p.write_text(s);d['Replace'][key]=str(p)
(out/'overlay.json').write_text(json.dumps(d,indent=2)+'\n')
