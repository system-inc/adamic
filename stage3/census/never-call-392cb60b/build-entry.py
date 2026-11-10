import pathlib,os,subprocess,sys,shutil,hashlib,json
out=pathlib.Path('/workspace/never-call-report/stage3/census/never-call-392cb60b')
for side in ['parent','head']:
 repo=pathlib.Path('/workspace/never-call-report' if side=='parent' else '/workspace/never-call-head');tool=pathlib.Path('/workspace/never-call-tools')/side
 main=tool/'overlay/stage3_census_latent_tool_main.go'
 source=main.read_text();needle='\tdiagnostics := program.LatentDiagnostics()'
 assert source.count(needle)==1
 source=source.replace(needle,'''\tif os.Getenv("LATENT_ENTRY_REACH") == "1" {
        if len(program.LatentRoots()) != 1 { panic("entry census requires one original root") }
        program.LatentReplayEntryReach()
        fmt.Fprintln(os.Stderr, "entry census ledger files", len(program.Files()))
    }
'''+needle).replace('"status": "measurement",','"status": "measurement", "root_files": paths, "entry_reach": os.Getenv("LATENT_ENTRY_REACH") == "1",')
 main.write_text(source)
 shutil.copyfile(main,out/(side+'-entry-driver.go.txt'))
 cmd=['go','build','-buildvcs=false','-overlay='+str(tool/'overlay/overlay.json'),'-o',str(tool/'census-entry'),'./stage3/census/latent/tool']
 with (out/(side+'-entry-build.log')).open('wb') as f:r=subprocess.run(cmd,cwd=repo,env={**os.environ,'GOWORK':str(tool/'go.work')},stdout=f,stderr=subprocess.STDOUT)
 print(side,'entry build child exit code:',r.returncode,flush=True)
 if r.returncode:sys.exit(r.returncode)
 (out/(side+'-entry-build.json')).write_text(json.dumps(dict(command=cmd,child_exit_code=r.returncode,sha256=hashlib.sha256((tool/'census-entry').read_bytes()).hexdigest()),indent=2)+'\n')
