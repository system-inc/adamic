package main
import("encoding/json";"fmt";"os";"path/filepath";"github.com/system-inc/adamic/internal/buildcache")
func main(){
 d,e:=os.MkdirTemp("/tmp","u003-keys-");if e!=nil{panic(e)};defer os.RemoveAll(d)
 key:=func(i buildcache.Inputs)string{k,e:=buildcache.Key(d,i);if e!=nil{panic(e)};return k}
 r:=map[string]any{}
 pair:=func(name,a,b string){r[name+"_before"]=a;r[name+"_after"]=b;r[name+"_equal"]=a==b}
 pair("flags",key(buildcache.Inputs{Name:"u003-witness",Flags:[]string{"-O1"}}),key(buildcache.Inputs{Name:"u003-witness",Flags:[]string{"-O2"}}))
 pair("toolchain",key(buildcache.Inputs{Name:"u003-witness",Toolchain:[]string{"compiler-A"}}),key(buildcache.Inputs{Name:"u003-witness",Toolchain:[]string{"compiler-B"}}))
 file:=filepath.Join(d,"source.txt");if e:=os.WriteFile(file,[]byte("alpha"),0644);e!=nil{panic(e)}
 a:=key(buildcache.Inputs{Name:"u003-witness",Files:[]string{"source.txt"}})
 if e:=os.WriteFile(file,[]byte("bravo"),0644);e!=nil{panic(e)}
 pair("files",a,key(buildcache.Inputs{Name:"u003-witness",Files:[]string{"source.txt"}}))
 b,e:=json.MarshalIndent(r,"","  ");if e!=nil{panic(e)};fmt.Println(string(b))
}
