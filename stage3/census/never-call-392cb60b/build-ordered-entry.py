import pathlib,subprocess,os,sys,json,hashlib,shutil
out=pathlib.Path(__file__).resolve().parent
for side in ['parent','head']:
 repo=pathlib.Path('/workspace/never-call-report' if side=='parent' else '/workspace/never-call-head');tool=pathlib.Path('/workspace/never-call-tools')/side
 hook=tool/'overlay/internal_load_latent_hook.go';source=(repo/'stage3/census/latent/load.go.txt').read_text()
 old='''func (p *Program) LatentReplayEntryReach() {
	p.files = nil
	for _, source := range p.compiler.GetSourceFiles() {
		if !source.IsDeclarationFile && !IsLibrary(source) {
			p.files = append(p.files, source)
		}
	}
}'''
 new='''func (p *Program) LatentReplayEntryReach() {
    // Production source-module identities start at Files()[0]. Keep the original
    // entry first while expanding only the measurement ledger to its closure.
    roots := append([]*ast.SourceFile{}, p.LatentRoots()...)
    p.files = roots
    seen := map[*ast.SourceFile]bool{}
    for _, source := range roots { seen[source] = true }
    for _, source := range p.compiler.GetSourceFiles() {
        if !source.IsDeclarationFile && !IsLibrary(source) && !seen[source] {
            p.files = append(p.files, source)
            seen[source] = true
        }
    }
}'''
 assert source.count(old)==1;hook.write_text(source.replace(old,new));shutil.copyfile(hook,out/(side+'-ordered-entry-load.go.txt'))
 main=tool/'overlay/stage3_census_latent_tool_main.go';source=(repo/'stage3/census/latent/main.go.txt').read_text();needle='\tdiagnostics := program.LatentDiagnostics()';assert source.count(needle)==1
 source=source.replace(needle,'''    if len(program.LatentRoots()) != 1 { panic("entry census requires one original root") }
    program.LatentReplayEntryReach()
    if program.Files()[0] != program.LatentRoots()[0] { panic("entry root lost first position") }
    fmt.Fprintln(os.Stderr, "entry census ledger files", len(program.Files()), "first root", program.FileName(program.Files()[0]))
'''+needle).replace('"status": "measurement",','"status": "measurement", "root_files": paths, "entry_reach": true, "entry_first": true,')
 main.write_text(source);shutil.copyfile(main,out/(side+'-ordered-entry-driver.go.txt'))
 cmd=['go','build','-buildvcs=false','-overlay='+str(tool/'overlay/overlay.json'),'-o',str(tool/'census-entry-ordered'),'./stage3/census/latent/tool']
 with (out/(side+'-ordered-entry-build.log')).open('wb') as f:r=subprocess.run(cmd,cwd=repo,env={**os.environ,'GOWORK':str(tool/'go.work')},stdout=f,stderr=subprocess.STDOUT)
 print(side,'ordered entry build child exit code:',r.returncode,flush=True)
 if r.returncode:sys.exit(r.returncode)
 (out/(side+'-ordered-entry-build.json')).write_text(json.dumps(dict(command=cmd,child_exit_code=r.returncode,sha256=hashlib.sha256((tool/'census-entry-ordered').read_bytes()).hexdigest()),indent=2)+'\n')
