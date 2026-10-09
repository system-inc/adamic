package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/optionschema"
	"github.com/system-inc/cohere/policy"
	"io"
	"os"
)

type Case struct {
	Kind, Rule, Id, Input, Schema, Expected string
	Values                                  map[string]string
	Choices                                 []policy.MessageOption
}

func render(c Case) (text string) {
	defer func() {
		if p := recover(); p != nil {
			text = "panic: " + fmt.Sprint(p)
		}
	}()
	return policy.Messages.Render(policy.MessageHandle{Rule: c.Rule, Id: c.Id}, c.Values, c.Choices...)
}
func main() {
	if os.Args[1] == "descriptors" {
		os.Stdout.Write(descriptors())
		return
	}
	if os.Args[1] == "catalog" {
		os.Stdout.Write(policy.AdamicCatalog())
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var corpus struct {
		Definitions map[string]string
		Cases       []Case
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		panic(err)
	}
	for _, c := range corpus.Cases {
		if c.Schema != "" {
			c.Schema = corpus.Definitions[c.Schema]
		}
		var result string
		switch c.Kind {
		case "json":
			var value any
			decoder := json.NewDecoder(bytes.NewReader([]byte(c.Input)))
			decoder.UseNumber()
			err := decoder.Decode(&value)
			var extra any
			if err == nil && decoder.Decode(&extra) == io.EOF {
				result = "valid"
			} else {
				result = "invalid"
			}
		case "schema":
			schema, err := optionschema.Compile([]byte(c.Schema))
			if err != nil {
				panic(err)
			}
			var elements []json.RawMessage
			if json.Unmarshal([]byte(c.Input), &elements) != nil || string(c.Input) == "null" {
				result = "invalid"
			} else if schema.Validate(elements) == nil {
				result = "valid"
			} else {
				result = "invalid"
			}
		case "decode":
			result = decode(c)
		case "message":
			result = render(c)
		default:
			panic(c.Kind)
		}
		if c.Expected != "" && result != c.Expected {
			panic("Go disagrees with ESLint sample for " + c.Rule)
		}
		fmt.Println(result)
	}
}
