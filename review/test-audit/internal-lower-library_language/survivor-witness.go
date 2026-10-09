package main
import (
 "context"
 "fmt"
 "os"
 "os/exec"
 "path/filepath"
 "time"
 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/internal/lower"
 "github.com/system-inc/adamic/internal/ir"
 "github.com/system-inc/adamic/internal/native"
)
func main() {
 directory, err := os.MkdirTemp("", "u034-witness-"); if err != nil { panic(err) }
 source := filepath.Join(directory,"main.a")
 if err := os.WriteFile(source, []byte("console.log(['10', '10', '10'].map(Number.parseInt).join(','));"),0600); err != nil { panic(err) }
 loaded,err:=load.Load([]string{source});if err!=nil{panic(err)}
 program,err:=lower.Lower(context.Background(),loaded);if err!=nil{panic(err)}
 for _,fn:=range program.Functions {if fn.Name=="library_map_parseInt" {for _,s:=range fn.Body {if ret,ok:=s.(ir.Return);ok {fmt.Printf("adapter: %#v\n",ret.Value)}}}}
 binary:=os.Args[1];start:=time.Now()
 err=native.Build(native.C(program),binary,native.Options{Sanitize:true})
 fmt.Printf("native_build_seconds: %.6f\nbuild_error: %v\n",time.Since(start).Seconds(),err)
 if err!=nil{return}
 output,err:=exec.Command(binary).CombinedOutput();fmt.Printf("native: %srun_error: %v\n",output,err)
}
