package main

import (
 "fmt"
 "os"
 "github.com/system-inc/adamic/internal/load"
)

func main() {
 path := "/tmp/u028-diagnostic-witness.a"
 if err := os.WriteFile(path, []byte("class Box { readonly value = 'initial'; } const box = new Box(); box.value = 'changed';"), 0644); err != nil { panic(err) }
 _, err := load.Load([]string{path})
 fmt.Println(err)
}
