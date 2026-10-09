from pathlib import Path
import json,difflib,subprocess,time,os,re
root=Path('/workspace/adamic'); ev=root/'review/test-audit/internal-regexp-parser'; ev.mkdir(exist_ok=True,parents=True)
p='internal/regexp/parser.go'; u='internal/unicodeproperties/canonicalize.go'
base={f:subprocess.check_output(['git','show','origin/main:'+f],cwd=root,text=True) for f in [p,u]}
menu=[
('M1',p,'seen[c] = true','seen[c] = false','change constant: stop tracking duplicate flags'),
('M2',p,'if f.Unicode && f.UnicodeSets {','if false {','drop u/v exclusion condition'),
('M3',p,'min.Cmp(max) > 0','min.Cmp(max) >= 0','off-by-one quantifier range comparison'),
('M4',p,"greedy := !p.take('?')","greedy := p.take('?')",'flip greediness'),
('M5',p,'if c.Value > right.Value {','if c.Value < right.Value {','flip class range comparison'),
('M6',p,'if !p.flags.Unicode && !p.flags.UnicodeSets && (kind == PositiveLookahead || kind == NegativeLookahead) {','if kind == PositiveLookahead || kind == NegativeLookahead {','drop Unicode guards for lookahead quantification'),
('M7',p,'\t\t\tn++','\t\t\tn += 2','change capture count increment constant'),
('M8',p,'value := p.peek() % 32','value := p.peek() % 31','change control escape modulus constant'),
('M9',p,'return classMayContainStrings(expression.Left) && classMayContainStrings(expression.Right)','return classMayContainStrings(expression.Left) || classMayContainStrings(expression.Right)','flip intersection condition'),
('M10',u,'return unicodeFold[i][0] >= cp','return unicodeFold[i][0] > cp','off-by-one Unicode search bound'),
('M11',u,'return legacyFold[index][1]','return legacyFold[index][0]','change legacy fold column constant'),
('M12',u,'codePoint > 0x10FFFF','codePoint > 0xFF','change Unicode range constant'),
]
metadata=[]
(ev/'diffs').mkdir(exist_ok=True)
for mid,f,old,new,kind in menu:
 assert base[f].count(old)==1 or mid=='M12',(mid,base[f].count(old))
 line=base[f][:base[f].index(old)].count('\n')+1
 mutated=base[f].replace(old,new,1)
 diff=''.join(difflib.unified_diff(base[f].splitlines(True),mutated.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
 (ev/'diffs'/f'{mid}.diff').write_text(diff)
 metadata.append(dict(id=mid,file=f,line=line,old=old,new=new,kind=kind))
(ev/'menu.json').write_text(json.dumps(metadata,indent=2)+'\n')
# Inventory from the clean coverage run, before any mutant execution.
lines=(ev/'reached-functions.txt').read_text().splitlines()
reached=[s for s in lines if len(s.split())>=3 and s.split()[-1]!='0.0%' and not s.startswith('total:')]
(ev/'code-and-oracles.md').write_text('Starting commit: 09fe4b54913753188a9357982bfd47cdf36ef97c\n\nCode under test: regexp Parse and reached parser helpers; unicodeproperties CanonicalizeUnicode and CanonicalizeLegacy. Oracle: handwritten parser admission and AST expectations; regexp generated tables for shared folding. All four are self oracles. Runtime canonicalization is the implementation under test in TestSharedCanonicalize; the regexp tables are never mutated. Package initializers also call set, which is preparation, not an entry.\n\nReached functions (clean coverage; callback closures execute within their enclosing functions):\n\n```\n'+'\n'.join(reached)+'\n```\n\nThe fixed 12-mutant menu is menu.json, recorded before any mutant outcome. No family wrappers in the four scoped rows. Distinct package checks are retained; TestMatcherOct6Mutants is a witness, so its production failures are excluded from uniqueness/subsumption. Its preconditions are recorded separately.\n')
# Build the selector scratch implementation. All standalone diffs above have no selector.
for f in base:
 scratch=base[f]
 scratch=scratch.replace('import (','import (\n\t"os"',1) if f==p else scratch.replace('import "sort"','import ("sort"; "os")')
 for mid,ff,old,new,kind in menu:
  if ff!=f:continue
  if old.startswith('if ') and old.endswith(' {'):
   origcond=old[3:-2]; mutcond=new[3:-2]
   replacement=f'if (os.Getenv("ADAMIC_MUTANT") == "{mid}" && ({mutcond})) || (os.Getenv("ADAMIC_MUTANT") != "{mid}" && ({origcond})) {{'
  elif old.startswith('return classMay') or old.startswith('return legacyFold'):
   replacement=f'if os.Getenv("ADAMIC_MUTANT") == "{mid}" {{ {new} }}\n\t\t'+old
  elif old.startswith('return unicodeFold'):
   replacement=f'return (os.Getenv("ADAMIC_MUTANT") == "{mid}" && unicodeFold[i][0] > cp) || (os.Getenv("ADAMIC_MUTANT") != "{mid}" && unicodeFold[i][0] >= cp)'
  elif mid=='M12':
   replacement='codePoint > func() rune { if os.Getenv("ADAMIC_MUTANT") == "M12" { return 0xFF }; return 0x10FFFF }()'
  elif mid=='M3':
   replacement='min.Cmp(max) > func() int { if os.Getenv("ADAMIC_MUTANT") == "M3" { return -1 }; return 0 }()'
  else:
   replacement=f'if os.Getenv("ADAMIC_MUTANT") == "{mid}" {{ {new} }} else {{ {old} }}'
   if ':=' in old:
    # Keep the existing variable declaration and override within its original scope.
    var=old.split(':=')[0].strip(); expr=new.split(':=')[1].strip()
    replacement=old+f'\n\tif os.Getenv("ADAMIC_MUTANT") == "{mid}" {{ {var} = {expr} }}'
    if mid=='M4':replacement="greedy := p.take('?')\n\tif os.Getenv(\"ADAMIC_MUTANT\") != \"M4\" { greedy = !greedy }"
  scratch=scratch.replace(old,replacement,1)
 entry='func Parse(pattern, flags string) (*Pattern, error) {' if f==p else 'func CanonicalizeUnicode(codePoint rune) rune {'
 probe='P1' if f==p else 'P2'; ret='return nil, nil' if f==p else 'return 0'
 scratch=scratch.replace(entry,entry+f'\n\tif os.Getenv("ADAMIC_MUTANT") == "{probe}" {{ {ret} }}',1)
 if f==u:scratch=scratch.replace('func CanonicalizeLegacy(codeUnit uint16) uint16 {','func CanonicalizeLegacy(codeUnit uint16) uint16 {\n\tif os.Getenv("ADAMIC_MUTANT") == "P3" { return 0 }',1)
 (root/f).write_text(scratch)
subprocess.run(['gofmt','-w',p,u],cwd=root,check=True)
for f in base:
 (ev/('scratch-'+Path(f).parent.name+'-'+Path(f).name+'.txt')).write_text((root/f).read_text())
