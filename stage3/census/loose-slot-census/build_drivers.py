import pathlib,json,subprocess
s=pathlib.Path('/workspace/loose-slot-scratch')
for which in ['head','main']:
 r=s/which
 text=(r/'stage3/census/tool/main.go').read_text(); text=text[:text.index('func main() {')]+'''func main() {
 output,err:=os.Create(os.Args[2]); if err!=nil {panic(err)}; defer output.Close()
 if err:=json.NewEncoder(output).Encode(inspect([]string{os.Args[1]}));err!=nil{panic(err)}
}
'''
 # Imports used only by original directory walker.
 for name in ['"path/filepath"','"sort"','"strings"','"fmt"']: text=text.replace(name,'')
 p=s/(which+'-driver.go');p.write_text(text)
 o=s/(which+'-driver.json');o.write_text(json.dumps({'Replace':{str(r/'stage3/census/tool/main.go'):str(p)}}))
 subprocess.run(['go','build','-overlay',str(o),'-o',str(s/(which+'-census')),'./stage3/census/tool'],cwd=r,check=True)
