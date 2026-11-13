import json
from pathlib import Path
repo=Path('/workspace/adamic');out=Path('/workspace/scratch/995spev')
path=repo/'stage3/census/latent/overlaytext/module_format_test.go'
s=path.read_text()
old='moduleFormatGap(string(p), string(c))'
new='moduleFormatGap(string(p), strings.Replace(string(c), "moduleFormat: moduleformat.Resolve(ModuleFormatCommonJSBundle, ModuleFormatEcmaScript), ", "", 1))'
assert s.count(old)==1
mutant=out/'format_guard_mutant_test.go';mutant.write_text(s.replace(old,new))
(out/'guard-mutant-overlay.json').write_text(json.dumps({'Replace':{str(path):str(mutant)}},indent=2)+'\n')
