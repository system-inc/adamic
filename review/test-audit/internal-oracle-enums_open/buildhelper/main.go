package main

import (
 "context"
 "encoding/json"
 "os"
 "time"
 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/internal/lower"
 "github.com/system-inc/adamic/internal/native"
)

func main() {
 program, err := load.Load([]string{"internal/oracle/testdata/enums_open_never.a"})
 if err != nil { panic(err) }
 ir, err := lower.Lower(context.Background(), program)
 if err != nil { panic(err) }
 code := native.C(ir)
 start := time.Now()
 if err = native.Build(code, os.Args[1], native.Options{Sanitize:true}); err != nil { panic(err) }
 json.NewEncoder(os.Stdout).Encode(map[string]any{"selector":os.Getenv("ADAMIC_MUTANT"), "native_build_seconds":time.Since(start).Seconds(), "fixture":"internal/oracle/testdata/enums_open_never.a"})
}
