#!/usr/bin/env python3
"""Extend an existing entry overlay; production files stay unchanged. Fails closed on drift."""
import json, pathlib, sys
source, output = map(pathlib.Path, sys.argv[1:3])
output.mkdir(parents=True, exist_ok=True)
m = json.loads((source/'overlay.json').read_text())
r = m['Replace']
key = next(k for k in r if k.endswith('/internal/lower/latent_hook.go'))
s = pathlib.Path(r[key]).read_text()
a = 'for _, node := range file.Statements.Nodes {\n\t\t\tlatentFindingOwner'
assert s.count(a)==1
s=s.replace(a,'for _, node := range file.Statements.Nodes {\n if !lwpSelected() { continue }\n\t\t\tlatentFindingOwner')
a='fresh().latentScan(file)'
assert s.count(a)==1
s=s.replace(a,'if os.Getenv("LWP_SKIP_SCAN") != "1" { fresh().latentScan(file) }')
if '--stock' not in sys.argv:
 a='latentRecord(l.lowerFunction(index, node, -1))'
 assert s.count(a)==1
 s=s.replace(a,'if err := l.lowerFunction(index, node, -1); err != nil { latentRecord(err); return }')
 a='_, err := l.statement(node)\n\t\t\t\t\tlatentRecord(err)'
 assert s.count(a)==1
 s=s.replace(a,'body, err := l.statement(node)\n if err != nil { latentRecord(err); return }; l.result.Main = append(l.result.Main, body...)')
 a='latentRecord(l.unlowerable)'
 assert s.count(a)==1
 s=s.replace(a,'if l.unlowerable != nil { latentRecord(l.unlowerable); return }; latentRecord(l.lwpTail(node, modules))')
p=output/'hook.go';p.write_text(s);r[key]=str(p.resolve())
key=key.replace('latent_hook.go','latent_whole_passes.go')
p=output/'passes.go';p.write_text(pathlib.Path(__file__).with_name('passes.go.txt').read_text());r[key]=str(p.resolve())
(output/'overlay.json').write_text(json.dumps(m,indent=2)+'\n')
