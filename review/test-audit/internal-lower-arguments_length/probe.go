package main

import (
 "context"
 "encoding/json"
 "fmt"
 "os"

 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/internal/lower"
)

func main() {
 checked, err := load.Load([]string{os.Args[1]})
 if err != nil { fmt.Println("Load:", err); os.Exit(1) }
 program, err := lower.Lower(context.Background(), checked)
 if err != nil { fmt.Println("Lower:", err); return }
 if err := json.NewEncoder(os.Stdout).Encode(program); err != nil { panic(err) }
}
