from pathlib import Path
import json
out=Path(__file__).resolve().parent;build=Path('/workspace/cbp-tool');build.mkdir(exist_ok=True)
m=json.loads(Path('/workspace/orm-build-main/scan-overlay/overlay.json').read_text())
keys=[k for k,p in m['Replace'].items() if 'func latentRun(' in Path(p).read_text()];assert len(keys)==1
s=Path(m['Replace'][keys[0]]).read_text();start=s.index('func latentRun(');a=s[:start];b=s[start:]
x='\tfor _, file := range program.Files() {';assert b.count(x)==1
b=b.replace(x,x+'\n        if !latentFileSelected(program.FileName(file)) { continue }',1)
x='\t\tfor _, node := range file.Statements.Nodes {';assert b.count(x)==1
b=b.replace(x,x+'''\n            if node.Name()==nil { continue }
            switch node.Name().Text() { case "createBuilderProgram", "createSemanticDiagnosticsBuilderProgram", "createEmitAndSemanticDiagnosticsBuilderProgram": default: continue }
''',1)
s=a+b;p=build/'selected.go';p.write_text(s);(out/'selected-lower.go.txt').write_text(s);m['Replace'][keys[0]]=str(p)
(build/'overlay.json').write_text(json.dumps(m,indent=2)+'\n');(out/'selected-overlay.json').write_text(json.dumps(m,indent=2)+'\n');print('selector retains full registration and normal per-unit first-error lowering')
