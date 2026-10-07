// Supplies raw dependencies to the unchanged .a source running on Node.
package main

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/system-inc/adamic/bridge/tsgo/checker"
)

type request struct {
	Command  string
	Config   string
	Files    []string
	File     string
	Start    uint64
	End      uint64
	Kind     string
	Question string
}

func main() {
	input := bufio.NewScanner(os.Stdin)
	input.Buffer(make([]byte, 4096), 16*1024*1024)
	output := json.NewEncoder(os.Stdout)
	var program *checker.Program
	for input.Scan() {
		var call request
		if err := json.Unmarshal(input.Bytes(), &call); err != nil {
			panic(err)
		}
		value := ""
		var err error
		switch call.Command {
		case "open":
			program, err = checker.Open(call.Config, call.Files)
		case "inspect":
			value, err = program.Inspect(call.File, call.Start, call.End, call.Kind, call.Question)
		case "release":
			program = nil
		default:
			panic("unknown raw dependency request")
		}
		message := ""
		if err != nil {
			message = err.Error()
		}
		if err := output.Encode(struct {
			Value string
			Error string
		}{value, message}); err != nil {
			panic(err)
		}
		if call.Command == "release" {
			return
		}
	}
	if err := input.Err(); err != nil {
		panic(err)
	}
}
