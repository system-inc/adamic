#!/usr/bin/env python3
"""Scratch overlay: retain checker errors; disable production output entry points."""
import json,sys
from pathlib import Path
repo=Path(sys.argv[1]).resolve();out=Path(sys.argv[2]).resolve();out.mkdir(parents=True,exist_ok=True)
unit=Path(__file__).resolve().parent;replace={}
def overlay(name,text):
 f=out/name.replace('/','_');f.write_text(text);replace[str(repo/name)]=str(f)
s=(repo/'internal/load/load.go').read_text();assert s.count('return nil, &CheckError{Diagnostics: diagnostics}')==1
s=s.replace('type Program struct {','type Program struct {\n hatchDiagnostics []string\n hatchAdamic bool')
s=s.replace('return nil, &CheckError{Diagnostics: diagnostics}','loaded.hatchDiagnostics = diagnostics')
s=s.replace('func Load(paths []string)','func HatchLoad(paths []string)').replace('func LoadOverlay(paths []string, overlay map[string]string)','func HatchLoadOverlay(paths []string, overlay map[string]string)')
s=s.replace('return p.fs.displayName(sourceFile.FileName())','name := p.fs.displayName(sourceFile.FileName()); if p.hatchAdamic { return strings.TrimSuffix(name,".ts")+".a" }; return name')
s+='\nfunc Load(paths []string) (*Program,error) { return nil,fmt.Errorf("predicates: measurement only; no output path") }\nfunc LoadOverlay(paths []string,overlay map[string]string) (*Program,error) {return Load(paths)}\nfunc (p *Program) HatchDiagnostics() []string {return p.hatchDiagnostics}\nfunc (p *Program) HatchAdamic(value bool) {p.hatchAdamic=value}\n'
overlay('internal/load/load.go',s)
s=(repo/'internal/lower/lower.go').read_text();start=s.index('\tfiles := program.Files()');end=s.index('\n}\n\ntype lowering',start)
s=s[:start]+'\treturn nil,fmt.Errorf("predicates: measurement only; no IR output")'+s[end:]
s=s.replace('\n\t"path/filepath"','');overlay('internal/lower/lower.go',s)
overlay('internal/lower/hatch_predicate_hook.go',(unit/'measure-hook.go.txt').read_text())
(out/'overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
print(out/'overlay.json')
