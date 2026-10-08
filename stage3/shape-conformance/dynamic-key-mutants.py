"""Every mutant builds valid Go and must fail its semantic allocation query."""
import json
import pathlib
import subprocess
root = pathlib.Path.cwd()
scratch = pathlib.Path('/tmp/shape-dynamic-mutants'); scratch.mkdir(exist_ok=True)
source = (root/'internal/lower/shape_flow.go').read_text()
mutants = {
 'drop-cached-sources': ('return cached.sources, cached.reasons, true', 'return cached.sources[:0], cached.reasons, true'),
 'drop-production-array-edge': ('if _, constant := value.Index.(ir.NumberConstant); !constant {', 'if _, constant := value.Index.(ir.NumberConstant); !constant && false {'),
 'drop-opaque-slot-store': ('sources = append(sources, index.opaqueSources[name]...)', 'sources = append(sources, index.opaqueSources[name][:0]...)'),
 'ignore-projection-depth': ('if depth >= 32 {', 'if depth >= 32 && false {'),
 'drop-key-arm': ('pending = append(pending, value.WhenTrue, value.WhenNot)', 'pending = append(pending, value.WhenTrue)'),
 'ignore-open-key': ('return result, unknown || len(result) == 0', 'return result, len(result) == 0 && unknown'),
 'ignore-absent-slot': ('if !found {\n\t\t\t\treasons = append(reasons, "dynamic projected slot', 'if !found && false {\n\t\t\t\treasons = append(reasons, "dynamic projected slot'),
 'ignore-indexed-store': ('if len(index.dynamicStores) != 0 {', 'if len(index.dynamicStores) != 0 && false {'),
 'drop-slot-store': ('sources = append(sources, index.stores[site][name]...)', 'sources = append(sources, index.stores[site][name][:0]...)'),
 'ignore-array-mutation': ('if index.arrayMutation {', 'if index.arrayMutation && false {'),
 'ignore-record-spread': ('if record.Spread != nil {', 'if record.Spread != nil && false {'),
 'drop-open-slots': ('if open {\n\t\t\tif record,', 'if open && false {\n\t\t\tif record,'),
}
for name,(before,after) in mutants.items():
 assert before in source, name
 changed = scratch/(name+'.go'); changed.write_text(source.replace(before,after))
 overlay = scratch/(name+'.json'); overlay.write_text(json.dumps({'Replace':{str(root/'internal/lower/shape_flow.go'):str(changed)}}))
 logpath = root/'stage3/shape-conformance/logs'/('dynamic-mutant-'+name+'.log')
 with logpath.open('w') as log:
  status = subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lower','-run','^TestShapeDynamicKeys$','-count=1','-v'],stdout=log,stderr=subprocess.STDOUT).returncode
 text = logpath.read_text()
 assert status != 0 and 'lost keys, slots or Unknown' in text and '[build failed]' not in text, (name,text)
 print(name+': semantic allocation assertion caught valid Go mutant',flush=True)
