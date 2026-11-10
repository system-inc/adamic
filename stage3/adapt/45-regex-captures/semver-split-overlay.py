from pathlib import Path
import json
root=Path(__file__).resolve().parents[3]
source=root/'internal/load/load.go'
s=source.read_text()
a='\troots = append(roots, preludePath)'
b='''\townedRoots := append([]tspath.RootedFilePath{}, roots...)
\tproject, projectDiagnostics := tsoptions.GetParsedCommandLineOfConfigFile(currentDirectory.ResolveFile(os.Getenv("CENSUS_CONFIG")), nil, nil, fs, nil)
\tif project == nil || len(projectDiagnostics) != 0 || len(project.Errors) != 0 { panic("invalid census project config") }
\troots = project.ParsedConfig.FileNames'''
assert s.count(a)==1
s=s.replace(a,b)
a='tsoptions.NewParsedCommandLine(compilerOptions(), roots,'
assert s.count(a)==1
s=s.replace(a,'tsoptions.NewParsedCommandLine(project.CompilerOptions(), roots,')
a='options := compilerOptions()'
assert s.count(a)==1
s=s.replace(a,'options := project.CompilerOptions()')
a='roots[:len(paths)]'
assert s.count(a)==1
s=s.replace(a,'ownedRoots')
out=root/'stage3/adapt/45-regex-captures/semver-split-evidence/census-load.go.txt'
out.write_text(s)
overlay=Path(__import__('sys').argv[1]); overlay.parent.mkdir(parents=True,exist_ok=True)
overlay.write_text(json.dumps({'Replace':{str(source):str(out)}},indent=2)+'\n')
