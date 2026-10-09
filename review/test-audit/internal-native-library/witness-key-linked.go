package main
import("fmt";_ "unsafe";_ "github.com/system-inc/adamic/internal/native")
type runtimeFile struct{name string;contents []byte}
//go:linkname productionKey github.com/system-inc/adamic/internal/native.runtimeKey
func productionKey([]runtimeFile,[]string,string,string)string
func main(){a:=productionKey(nil,[]string{"a","bc"},"clang","version");b:=productionKey(nil,[]string{"ab","c"},"clang","version");fmt.Printf("equal=%t a=%s b=%s\n",a==b,a,b)}
