package main
import (
 "context"
 "fmt"
 "os"
 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/internal/lower"
)
func main() {
 const source = `let calls = 0; const source: { readonly value: string | undefined } = { get value(): string | undefined { calls++; return calls === 1 ? 'first' : undefined; } }; if (source.value !== undefined) console.log(source.value);`
 path := "/tmp/u028-narrowed-witness.a"
 if err := os.WriteFile(path, []byte(source), 0644); err != nil { panic(err) }
 checked, err := load.Load([]string{path}); if err != nil { panic(err) }
 _, err = lower.Lower(context.Background(), checked)
 fmt.Println(err)
}
