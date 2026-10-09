import pathlib,json
s=pathlib.Path('review/test-audit/stage1-cohere-markdownblocks-structure_layout_shards/run_mutations.py').read_text()
exec(s[:s.index('# Mutants and probes')])
exec(s[s.index('def body('):s.index('for id,name,anchor,value in probes:')])
meta=json.loads((OUT/'matrix-meta.json').read_text())
exec(s[s.index('# Construction mutations'):])
