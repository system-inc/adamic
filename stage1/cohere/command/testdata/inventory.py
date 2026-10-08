"""Complete lexical inventories; candidates are not claims of semantic duplication."""
import collections, hashlib, json, pathlib, re, sys
original=pathlib.Path(__file__).resolve().parents[4]
root=pathlib.Path(sys.argv[1]) if len(sys.argv)>1 else original
entries=[sys.argv[2] if len(sys.argv)>2 else 'stage1/cohere/command/main.a']
seen=set();missing=[]
def graph(name):
    path=root/name
    if name in seen:return
    if not path.exists():missing.append(name);return
    seen.add(name)
    for spec in re.findall(r"\bfrom\s+['\"]([^'\"]+)['\"]",path.read_text()):
        if spec.startswith('.'):
            graph(str((path.parent/spec).resolve().relative_to(root)))
for entry in entries:graph(entry)
symbols=collections.defaultdict(list);state=[];hashes=collections.defaultdict(list)
for name in sorted(seen):
    data=(root/name).read_bytes();hashes[hashlib.sha256(data).hexdigest()].append(name)
    for line,text in enumerate(data.decode().splitlines(),1):
        m=re.match(r'(?:export )?(?:async )?(?:function|class|interface|type|const|let)\s+([A-Za-z_$][\w$]*)',text)
        if m:symbols[m[1]].append(f'{name}:{line}')
        if re.match(r'(?:export )?(?:const|let)\s+',text):state.append(dict(site=f'{name}:{line}',declaration=text))
go=[]
areas=['cohere/command/cohere','cohere/internal/format/native','cohere/internal/format/formatoptions']
for area in areas:
    for path in sorted((original/area).glob('*.go')):
        if path.name.endswith('_test.go'):continue
        for line,text in enumerate(path.read_text().splitlines(),1):
            if re.match(r'^var\b',text):go.append(dict(site=f'{path.relative_to(original)}:{line}',declaration=text))
report=dict(modules=sorted(seen),missing=missing,duplicate_names={k:v for k,v in sorted(symbols.items()) if len(v)>1},byte_identical_modules=[v for v in hashes.values() if len(v)>1],module_declarations=state,go_var_blocks=go)
path=pathlib.Path(sys.argv[3]) if len(sys.argv)>3 else pathlib.Path(__file__).with_name('inventory.json');path.write_text(json.dumps(report,indent=2)+'\n')
print('modules',len(seen),'duplicate names',len(report['duplicate_names']),'declaration sites',sum(map(len,report['duplicate_names'].values())),'module state declarations',len(state),'identical groups',len(report['byte_identical_modules']),'Go var blocks',len(go))
