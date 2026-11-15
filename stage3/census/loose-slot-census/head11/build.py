from pathlib import Path
import concurrent.futures,json,subprocess
s=Path('/workspace/loose-slot-scratch/eleven');out=Path(__file__).resolve().parent

def build(which):
 r=s/which;t=(r/'stage3/census/tool/main.go').read_text();t=t[:t.index('func main() {')]+'''func main(){output,err:=os.Create(os.Args[2]);if err!=nil{panic(err)};defer output.Close();if err:=json.NewEncoder(output).Encode(inspect([]string{os.Args[1]}));err!=nil{panic(err)}}\n'''
 for name in ['"path/filepath"','"sort"','"strings"','"fmt"']:t=t.replace(name,'')
 p=s/(which+'-driver.go');p.write_text(t);overlay=s/(which+'-driver.json');overlay.write_text(json.dumps({'Replace':{str(r/'stage3/census/tool/main.go'):str(p)}}))
 subprocess.run(['go','build','-overlay',str(overlay),'-o',str(s/(which+'-census')),'./stage3/census/tool'],cwd=r,check=True)
 print(which,'ordinary binary ready',flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=2)as pool:list(pool.map(build,['head','base']))
