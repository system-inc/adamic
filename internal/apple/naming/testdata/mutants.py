# Not parallel: mutates implementation files; run alone in this checkout.
from pathlib import Path
import subprocess
import json
root=Path('internal/apple/naming')
mutants=[
 ('omit-needless-words','swift.go','base, labels = omitNeedlessWords(base, labels, d)','// mutant: leave needless type words in place','TestDocumentationFixtures'),
 ('enum-one-character-too-many','swift.go','base = strings.TrimPrefix(d.Name, prefix)','if len(prefix) < len(d.Name) {prefix = d.Name[:len(prefix)+1]}\n\t\tbase = strings.TrimPrefix(d.Name, prefix)','TestDocumentationFixtures'),
 ('url-left-uppercase','words.go','w = strings.ToLower(w)','if w == "URL" {ws[i] = w; continue}\n\t\tw = strings.ToLower(w)','TestNamesAndWordBoundaries'),
 ('collision-accepted','naming.go','if other, exists := surface.Reverse[key]; exists {','if other, exists := surface.Reverse[key]; exists && false {','TestCollisionsAreRejected'),
 ('namespace-ns-retained','words.go','[]string{"NS", "UI", "CG", "CF", "CA", "AV"}','[]string{"UI", "CG", "CF", "CA", "AV"}','TestDocumentationFixtures'),
 ('rectangle-not-expanded','words.go','"rect": "Rectangle"','"rect": "Rect"','TestNamesAndWordBoundaries'),
 ('positional-moved-to-options','naming.go','layout.Positional = append(layout.Positional, arg)','layout.Options = append(layout.Options, arg)','TestArgumentLayout'),
 ('parameter-index-shifted','naming.go','Index: i, SwiftLabel: label','Index: i+1, SwiftLabel: label','TestArgumentLayout'),
 ('unspecified-nullability-dropped','naming.go','((pointer || t.Reference) && t.Nullability == Unspecified)','false','TestTypeMapping'),
 ('swift-name-ignored','swift.go','if explicitName(d) != "" {','if explicitName(d) != "" && false {','TestDocumentationFixtures'),
 ('refinement-marker-lost','swift.go','if d.RefinedForSwift {','if d.RefinedForSwift && false {','TestErrorsAndMetadata'),
 ('boolean-getter-ignored','swift.go','if omission(d.Result).boolean && d.Getter != "" {','if omission(d.Result).boolean && d.Getter != "" && false {','TestErrorsAndMetadata'),
 ('constructor-not-marked','naming.go','Constructor: strings.TrimPrefix(base, "__") == "init"','Constructor: false','TestArgumentLayout'),
 ('struct-pointer-accepted-as-object','naming.go','if pointer && !t.Object {','if pointer && !t.Object && false {','TestTypeMapping'),
 ('property-protection-dropped','swift.go','propertyMatch(strings.Join(ws[i:], ""), properties)','false','TestPortSafeguards'),
 ('back-map-not-recorded','naming.go','surface.Reverse[key] = original','// mutant: lose the reverse entry','TestFixtureBijection'),
]
results=[]
for name,file,old,new,test in mutants:
 p=root/file;original=p.read_bytes();s=original.decode()
 if s.count(old)!=1:raise RuntimeError((name,s.count(old),old))
 try:
  p.write_text(s.replace(old,new,1))
  logfile=Path('/tmp/apple-mutant-'+name+'.log')
  with logfile.open('w') as out:
   r=subprocess.run(['go','test','./internal/apple/naming','-count=1','-run','^'+test+'$'],stdout=out,stderr=subprocess.STDOUT)
  report=logfile.read_text()
  failed=[x.strip() for x in report.splitlines() if '--- FAIL:' in x]
  if r.returncode==0 or not failed:raise RuntimeError((name,'survived or did not fail a test',report))
  results.append(dict(mutant=name,test=test,exit=r.returncode,failures=failed[:3],log=str(logfile)))
  print(name, 'caught by',test,'exit',r.returncode)
 finally:
  p.write_bytes(original)
Path('/tmp/apple-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
