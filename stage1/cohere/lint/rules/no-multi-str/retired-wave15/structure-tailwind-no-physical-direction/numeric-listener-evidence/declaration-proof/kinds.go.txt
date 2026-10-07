package main
import ("encoding/json"; "os"; "strings"; "github.com/microsoft/TypeScript/tsc/shim/ast")
func main() {
 values := map[string]int{}
 for kind := ast.Kind(0); kind < ast.KindCount; kind++ { values[strings.TrimPrefix(kind.String(), "Kind")] = int(kind) }
 if err := json.NewEncoder(os.Stdout).Encode(values); err != nil { panic(err) }
}
