// Probe records the real checker gate and only lowers accepted programs.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func main() {
	if len(os.Args) != 3 {
		panic("usage: probe entry output.json")
	}
	result := map[string]any{"entry": os.Args[1]}
	program, err := load.Load([]string{os.Args[1]})
	if err == nil {
		_, err = lower.Lower(context.Background(), program)
		result["outcome"] = "accepted"
	}
	if err != nil {
		var checked *load.CheckError
		var refused *lower.Refused
		var notYet *lower.NotYet
		switch {
		case errors.As(err, &checked):
			result["outcome"] = "Checker"
			result["diagnostics"] = checked.Diagnostics
		case errors.As(err, &refused):
			result["outcome"] = "Refused"
			result["where"] = refused.Where
			result["what"] = refused.What
		case errors.As(err, &notYet):
			result["outcome"] = "NotYet"
			result["where"] = notYet.Where
			result["what"] = notYet.What
		default:
			result["outcome"] = "error"
			result["what"] = err.Error()
		}
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[2], append(data, '\n'), 0644); err != nil {
		panic(err)
	}
}
