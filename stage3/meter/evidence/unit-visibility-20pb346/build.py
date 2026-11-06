from pathlib import Path
import json, subprocess, sys
side=sys.argv[1]
root=Path('/workspace/scratch/20pb346-2')
repo=Path('/workspace/adamic')
overlay=root/(side+'-overlay')
entry=root/(side+'-entry-overlay')
if side=='after':
    subprocess.run(['python3',str(repo/'stage3/census/latent/make_overlay.py'),str(repo),str(overlay)],check=True)
    subprocess.run(['python3',str(repo/'stage3/meter/entry_overlay.py'),str(overlay),str(entry)],check=True)
# Only the scratch entry driver is instrumented. Stop profiling at 120 s while
# allowing the measured census to continue. The compiler driver is unchanged.
path=entry/'entry_main.go'
s=path.read_text().replace('"context"','"context"\n"runtime/pprof"\n"sync"\n"time"',1)
needle='\tlower.Latent(context.Background(), program,'
assert s.count(needle)==1
s=s.replace(needle,'''    if path := os.Getenv("LATENT_CPU_PROFILE"); path != "" {
        profile, err := os.Create(path); if err != nil { panic(err) }
        if err := pprof.StartCPUProfile(profile); err != nil { panic(err) }
        var once sync.Once
        stop := func() { once.Do(func() { pprof.StopCPUProfile(); if err:=profile.Close(); err!=nil {panic(err)} }) }
        timer := time.AfterFunc(120*time.Second,stop)
        defer func() { timer.Stop(); stop() }()
    }
'''+needle)
path.write_text(s)
for folder in (overlay,entry):
    subprocess.run(['gofmt','-w',*[str(p) for p in folder.glob('*.go')]],check=True)
for mode,folder in [('entry',entry),('compiler',overlay)]:
    subprocess.run(['go','build','-buildvcs=false','-overlay='+str(folder/'overlay.json'),'-o',str(root/(side+'-'+mode)), './stage3/census/latent/tool'],check=True,cwd=repo)
