package main
import("fmt";"os/exec";"time";"github.com/system-inc/adamic/internal/testguard")
func main(){ c:=exec.Command("node","-e","const until=process.hrtime.bigint()+1500000000n;while(process.hrtime.bigint()<until){}"); fmt.Printf("guard error=%v\n",testguard.Run(c,3*time.Second,5*time.Second)) }
