// Built only in the rule package overlay for the descriptor oracle.
package rule

import (
	"bytes"
	"encoding/json"
	"reflect"
)

var AdamicRegistrations []Registrant

// AdamicUnmarshalOptions keeps the oracle's dynamically selected target while using
// cohere's strict key checking, which UnmarshalOptions now reaches through a type parameter.
func AdamicUnmarshalOptions(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var document json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		return err
	}
	return checkKeyCase(document, reflect.TypeOf(target), "")
}
