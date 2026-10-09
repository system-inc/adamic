from pathlib import Path
import json,os,subprocess,tempfile
root=Path('/workspace/area-stack-review')
mutants=[
('lowering-optional-flag','internal/lower/object.go','Element: element, Optional: optional}), nil','Element: element, Optional: false}), nil','exit codes differ'),
('native-eager-key','internal/native/emit_slots.go','\t\te.line("if (%s != NULL) {", array)\n\t\te.out.WriteString(text)','\t\te.out.WriteString(text)\n\t\te.line("if (%s != NULL) {", array)','stdout differs'),
('native-repeat-receiver','internal/native/emit_slots.go','\t\te.out.WriteString(text)','''		repeated, _, repeatedOwned := e.aside(expression.Array)
		e.out.WriteString(repeated)
		e.indent++
		for i := len(repeatedOwned)-1; i >= 0; i-- { e.line("adamic_release(%s);", repeatedOwned[i]) }
		e.indent--
		e.out.WriteString(text)''','stdout differs'),
]
for name,relative,before,after,catcher in mutants:
 source=root/relative;clean=source.read_text();assert clean.count(before)==1,(name,clean.count(before))
 with tempfile.TemporaryDirectory(prefix='index-mutant-') as d:
  p=Path(d);changed=p/source.name;changed.write_text(clean.replace(before,after));overlay=p/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(changed)}}));log=Path('/tmp/area-stack-group3-'+name+'.log')
  with log.open('w') as f:r=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/element_access_optional','-count=1','-timeout','10m'],cwd=root,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=f,stderr=subprocess.STDOUT)
  output=log.read_text();assert r.returncode==1 and catcher in output and '[build failed]' not in output and 'clang failed' not in output,(name,output)
  print(name+': caught by '+catcher,flush=True)
