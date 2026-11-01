import json,pathlib,collections
b=pathlib.Path('/workspace/scratch/axolotl')
result={}
for mode in ['main','es2020','es2020-capped-node']:
 result[mode]={}
 for label in ['tsc','empty','node']:
  rows=[json.loads(s) for s in (b/mode/('probe-'+label+'.jsonl')).read_text().splitlines()]
  head=rows[0]
  libs=sorted(r['file'] for r in rows if 'file' in r and r['file'].startswith('bundled:///libs/'))
  result[mode][label]={'target':head['target'],'lib':head['lib'],'diagnostics':head['diagnostics'] or [],'loaded_libs':libs,'set_declarations':[r for r in rows if 'set_declaration' in r],'set_members':[r for r in rows if 'set_member' in r]}
  assert not any('lib.es2025.collection.d.ts' in s for s in libs)
  assert sum('union<U>' in r['text'] for r in rows if 'set_member' in r)==1
base=result['main']['tsc']['diagnostics']
assert result['es2020']['tsc']['diagnostics']==base
assert all(not result[m][p]['diagnostics'] for m in ['main','es2020'] for p in ['empty','node'])
assert not result['es2020-capped-node']['empty']['diagnostics']
(b/'probe-summary.json').write_text(json.dumps(result,indent=2)+'\n')
lines=['| Probe diagnostics | Main | ES2020 | ES2020, capped Node references |','| --- | ---: | ---: | ---: |']
for label in ['empty','node','tsc']:
 lines.append('| '+label+' | '+' | '.join(str(len(result[m][label]['diagnostics'])) for m in result)+' |')
for mode in ['es2020','es2020-capped-node']:
 before=set(result['main']['tsc']['loaded_libs']);after=set(result[mode]['tsc']['loaded_libs'])
 lines+=['','Loaded lib changes, main to '+mode+':','', 'Removed:']
 lines+=['- `'+x+'`' for x in sorted(before-after)] or ['- None.']
 lines+=['','Added:']
 lines+=['- `'+x+'`' for x in sorted(after-before)] or ['- None.']
(b/'probe-tables.md').write_text('\n'.join(lines)+'\n')
print('\n'.join(lines[:5]))
