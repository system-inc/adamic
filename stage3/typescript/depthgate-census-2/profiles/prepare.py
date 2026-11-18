from pathlib import Path
import json,hashlib,difflib,subprocess
W=Path('/workspace');root=W/'depthgate2-profiles';root.mkdir(exist_ok=True)
base=(W/'depthgate2-tools-base/entry-overlay/entry_main.go').read_text();head=(W/'depthgate2-tools-head/entry-overlay/entry_main.go').read_text()
assert base==head
hook='''func main() {
	if profilePath := os.Getenv("ADAMIC_CPU_PROFILE"); profilePath != "" {
		profile, err := os.Create(profilePath)
		if err != nil { panic(err) }
		if err := pprof.StartCPUProfile(profile); err != nil { panic(err) }
		defer func() {
			pprof.StopCPUProfile()
			if err := profile.Close(); err != nil { panic(err) }
		}()
	}
'''
main=base.replace('"path/filepath"','"path/filepath"\n\t"runtime/pprof"').replace('func main() {\n',hook)
for label in ['base','head']:
 out=root/label;out.mkdir(exist_ok=True)
 p=out/'entry_main.go';p.write_text(main)
 subprocess.run(['/workspace/adamic-tools/go/bin/gofmt','-w',str(p)],check=True)
 overlay=json.loads((W/f'depthgate2-tools-{label}/entry-overlay/overlay.json').read_text())
 overlay['Replace'][str(W/f'depthgate2-{label}/stage3/census/latent/tool/main.go')]=str(p)
 (out/'overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
 (out/'profile-hook.diff').write_text(''.join(difflib.unified_diff(base.splitlines(True),p.read_text().splitlines(True),fromfile='canonical-entry-main',tofile='scratch-profile-entry-main')))
assert (root/'base/entry_main.go').read_bytes()==(root/'head/entry_main.go').read_bytes()
print('Identical scratch-only profiling hooks prepared for both tools.')
