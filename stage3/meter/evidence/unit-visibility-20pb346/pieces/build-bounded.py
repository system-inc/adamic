"""Apply the same proof-only full-mode unit selector to both scratch overlays.

No repository compiler file or shipped census template is changed. Empty
LATENT_PROOF_UNITS preserves ordinary file selection; a nonempty value is exact
newline-separated Where|Kind membership. An inventory-only invocation loads the
program but attempts no units. The unmodified file-group binaries remain saved.
"""
from pathlib import Path
import json, shutil, subprocess
root=Path('/workspace/scratch/20pb346-2')
repo=Path('/workspace/adamic')
for side in ('before','after'):
    old=root/('pieces-'+side+'-overlay')
    new=root/('bounded-'+side+'-overlay')
    shutil.copytree(old,new,dirs_exist_ok=True)
    overlay=json.loads((new/'overlay.json').read_text())
    overlay['Replace']={k:v.replace(str(old),str(new),1) for k,v in overlay['Replace'].items()}
    (new/'overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
    path=new/'internal_lower_latent_units.go'
    text=path.read_text().replace('"context"','"context"\n"os"\n"strings"',1)
    needle='func latentFullSelected(ctx context.Context, program *load.Program, emit func(LatentFile), selection *latentReplaySelection) {'
    assert text.count(needle)==1
    text=text.replace(needle,needle+'''
    var proofUnits map[string]bool
    if requested := os.Getenv("LATENT_PROOF_UNITS"); requested != "" {
        proofUnits = map[string]bool{}
        for _, key := range strings.Split(requested,"\\n") { proofUnits[key] = true }
    }
''',1)
    needle='node := candidate.node'
    assert text.count(needle)==2
    prefix, suffix = text.rsplit(needle,1)
    text=prefix+needle+'''
    if proofUnits != nil && !proofUnits[program.Where(node)+"|"+node.Kind.String()] { continue }
'''+suffix
    text+='''
func LatentProofInventory(program *load.Program) map[string][]string {
    result := map[string][]string{}
    for _, file := range program.Files() {
        keys := []string{}
        for _, candidate := range latentCandidates(program,file) {
            keys = append(keys,program.Where(candidate.node)+"|"+candidate.node.Kind.String())
        }
        result[program.FileName(file)] = keys
    }
    return result
}
'''
    path.write_text(text)
    driver=new/'stage3_census_latent_tool_main.go'
    text=driver.read_text()
    needle='lower.Latent(context.Background(), program,'
    assert text.count(needle)==1
    text=text.replace(needle,'''if os.Getenv("LATENT_PROOF_INVENTORY") == "1" {
        if err := encoder.Encode(lower.LatentProofInventory(program)); err != nil { panic(err) }
        return
    }
    '''+needle,1)
    driver.write_text(text)
    subprocess.run(['gofmt','-w',str(path),str(driver)],check=True)
    subprocess.run(['go','build','-buildvcs=false','-overlay='+str(new/'overlay.json'),'-o',str(root/('bounded-'+side)),'./stage3/census/latent/tool'],cwd=repo,check=True)
