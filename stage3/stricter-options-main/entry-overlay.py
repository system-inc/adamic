from pathlib import Path
import subprocess,re,json
root=Path.cwd();out=Path('/tmp/stricter-options-entry-overlay');out.mkdir(exist_ok=True)
helper=subprocess.check_output(['git','show','88fe8de4:internal/load/project_references.go'],text=True)
imports=re.search(r'import \((.*?)\n\)',helper,re.S)[1];body=helper[helper.index('// projectSourceRoots'):]
p=root/'internal/load/load.go';s=p.read_text();pos=s.index('\n)',s.index('import ('))
for line in imports.splitlines():
 if line.strip() and line.strip() not in s[:pos]:s=s[:pos]+ '\n'+line +s[pos:];pos=s.index('\n)',s.index('import ('))
s+='\n'+body
s=s.replace('var projectConfig *tsoptions.ParsedCommandLine','var projectOwners map[string]bool\n\tvar projectConfig *tsoptions.ParsedCommandLine')
a=s.index('\t\tfor _, name := range projectConfig.FileNames()');b=s.index('\n\t}\n\tconfig :=',a)
s=s[:a]+'''\t\tvar sourceRoots []tspath.RootedFilePath
        var packages []string
        sourceRoots, projectOwners, packages, err = projectSourceRoots(fs, projectConfig)
        if err != nil { return nil, err }
        options = sourceProgramOptions(options, sourceRoots)
        options.Types = packages
        if len(packages) > 0 { options.SkipLibCheck = core.TSFalse }
        checkRoots = uniqueSourceRoots(fs, append(sourceRoots, preludePath))'''+s[b:]
s=s.replace('projectConfig.ProjectReferences(), tspath.RootedDirectoryPathFromAbsolute(filepath.Dir(project))','nil, tspath.RootedDirectoryPathFromAbsolute(filepath.Dir(project))')
s=s.replace('owner != project {','!projectOwners[owner] {')
q=out/'load.go';q.write_text(s);overlay={str(p):str(q)}
p=root/'internal/load/project_options.go';s=p.read_text();s=s.replace('roots := append([]tspath.RootedFilePath{}, config.FileNames()...)','roots, _, packages, err := projectSourceRoots(fs, config)\n\tif err != nil { return nil, err }')
s=s.replace('base := config.CompilerOptions().Clone()','base := sourceProgramOptions(config.CompilerOptions(), roots)\n\tbase.Types = packages\n\tif len(packages) > 0 { base.SkipLibCheck = core.TSFalse }')
s=s.replace('roots, config.ProjectReferences(), directory','roots, nil, directory')
q=out/'project_options.go';q.write_text(s);overlay[str(p)]=str(q)
(out/'overlay.json').write_text(json.dumps({'Replace':overlay}))
