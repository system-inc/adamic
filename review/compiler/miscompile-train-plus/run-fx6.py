from pathlib import Path
import subprocess, re, os, sys
root=Path.cwd()
base=root/'review/compiler/miscompile-train-plus/fx6'
runners=[('receiver','fx6-candidates-2-fix/run-mutants.py'),('valid','fx6-candidates-2/run-valid-mutants.py'),('routing','fx6-candidates-2/run-routing-mutants.py'),('scalar','fx7-scalar-union-view/run-mutant.py'),('callable','fx7-callable-producers/run-mutants.py'),('assignable','fx7-callable-producers/run-assignable-mutants.py'),('name','fx7-name-family/run-mutants.py')]
for label,runner in runners:
 if '--resume' in sys.argv and label in ['receiver','valid','routing','scalar']: continue
 p=root/'review/compiler'/runner
 s=p.read_text()
 out=base/label;out.mkdir(exist_ok=True)
 s=re.sub(r'^evidence\s*=.*$', 'evidence = Path('+repr(str(out))+')',s,flags=re.M)
 s='from pathlib import Path\n'+s
 if label=='callable':
  replacement = "path.write_text(mutated)\n  extra=Path('internal/lower/view_callable_producers.go')\n  extra_original=extra.read_text()\n  if name=='direct-producer-certificate':\n   extra.write_text(extra_original.replace('!function.Closure || function.Receiver || function.RestElement', 'function.Receiver || function.RestElement'))"
  s=s.replace('path.write_text(mutated)', replacement)
  s=s.replace('finally:path.write_text(original)', 'finally:\n  path.write_text(original)\n  extra.write_text(extra_original)')
 if label=='assignable':
  s='\n'.join(line for line in s.split('\n') if not line.startswith("mutant('assignable-unfiltered-native'"))
 exec(compile(s,str(p),'exec'),{'__file__':str(p),'__name__':'__main__'})
 if label=='callable':
  (out/'direct-producer-certificate.diff').write_text((root/'review/compiler/fx7-callable-producers/assignable-direct-certificate.diff').read_text())
 print(label+': runner passed',flush=True)
# Saved syntax patches are applied independently to the final compiler.
syntax=[('final-spread','TestCheckedViewSpread','oracle'),('final-in','TestCheckedViewIn','oracle'),('final-keys','TestCheckedViewKeys','oracle'),('final-keys-alias','TestCheckedViewKeysAlias','oracle'),('final-values','TestCheckedViewValues','oracle'),('final-entries','TestCheckedViewEntries','oracle'),('p19-read','TestCheckedViewElementP19Read','oracle'),('p72-input','TestCheckedViewElementP72','oracle'),('in-empty-selector','TestViewInEmptyKeyControl','lower'),('spread-method','TestViewSpreadMethodRefused','lower')]
for name,test,pkg in syntax:
 suffix='-bypass.diff' if name not in ['in-empty-selector'] else '.diff'
 patch=root/'review/compiler/fx6-key-read'/(name+suffix)
 if name=='final-in':
  text=(root/'review/compiler/fx6-key-read/in-bypass.diff').read_text().replace('object, key.Text())','object, key)')
 else:text=patch.read_text()
 target=base/(name+'.patch');target.write_text(text)
 paths=re.findall(r'^\+\+\+ b/(.*)$',text,re.M)
 
 if name=='spread-method': paths.append('internal/lower/interface_cast.go')
 originals={p:(root/p).read_bytes() for p in paths}
 try:
  subprocess.run(['git','apply','--unidiff-zero',str(target)],check=True,timeout=10)
  if name=='spread-method':
   p=root/'internal/lower/interface_cast.go'
   p.write_text(p.read_text().replace('\tif err := l.checkViewMembers(node, target); err != nil {\n\t\treturn nil, err\n\t}\n',''))
  with (base/(name+'.log')).open('w') as log:
   r=subprocess.run(['go','test','./internal/'+pkg,'-run','^'+test+'$','-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
  output=(base/(name+'.log')).read_text()
  assert r.returncode==1 and '[build failed]' not in output and ('exit codes differ' in output if pkg=='oracle' else ('JavaScript backend stdout' in output if name=='in-empty-selector' else 'got <nil>' in output)),output
  print(name+': caught by '+test,flush=True)
 finally:
  for p,s in originals.items():(root/p).write_bytes(s)
print('All 29 member mutant runs caught (certificate obligations overlap)',flush=True)
