package main
import("os"; "github.com/system-inc/adamic/internal/native")
func main(){ b,e:=os.ReadFile(os.Args[1]); if e!=nil{panic(e)}; if e=native.Build(string(b),os.Args[2],native.Options{Sanitize:true}); e!=nil{panic(e)} }
