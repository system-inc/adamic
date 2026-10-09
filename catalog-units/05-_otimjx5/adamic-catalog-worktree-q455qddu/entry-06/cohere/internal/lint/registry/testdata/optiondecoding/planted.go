// Package optiondecoding is planted for TestTheOptionDecodingGuardSeesALooseDecodeAndAnUntaggedField:
// each function is one shape the guard must report or must leave alone, named for which.
package optiondecoding

import (
	"bytes"
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
)

type plantedOptions struct {
	Tagged     bool `json:"tagged"`
	Untagged   bool
	Skipped    bool `json:"-"`
	unexported bool
	Nested     plantedNested `json:"nested"`
	plantedEmbedded
}

type plantedEmbedded struct {
	Promoted bool `json:"promoted"`
}

type plantedNested struct {
	Deep int
}

type plantedStrictOptions struct {
	AlsoUntagged bool
}

// plantedSelfDecoding decides its own shape, so the guard does not look inside it; the decode in its
// UnmarshalJSON is found on its own.
type plantedSelfDecoding struct {
	Inside bool
}

func (decoded *plantedSelfDecoding) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Inside bool `json:"inside"`
	}
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return err
	}
	decoded.Inside = wire.Inside
	return nil
}

func decodeLoose(raw []byte) (plantedOptions, error) {
	var options plantedOptions
	err := json.Unmarshal(raw, &options)
	return options, err
}

func decodeLooseNested(raw []byte) ([]map[string]plantedNested, error) {
	var options []map[string]plantedNested
	err := json.Unmarshal(raw, &options)
	return options, err
}

func decodeWithDecoder(raw []byte) (plantedOptions, error) {
	var options plantedOptions
	err := json.NewDecoder(bytes.NewReader(raw)).Decode(&options)
	return options, err
}

func decodeLooseAnonymous(raw []byte) (int, error) {
	var options struct {
		Count int `json:"count"`
	}
	err := json.Unmarshal(raw, &options)
	return options.Count, err
}

func decodeScalars(raw []byte) (string, []json.RawMessage, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil, nil
	}
	var list []json.RawMessage
	err := json.Unmarshal(raw, &list)
	return "", list, err
}

func decodeStrict(raw []byte) (plantedStrictOptions, error) {
	var options plantedStrictOptions
	err := rule.UnmarshalOptions(raw, &options)
	return options, err
}

func decodeSelf(raw []byte) (plantedSelfDecoding, error) {
	var options plantedSelfDecoding
	err := json.Unmarshal(raw, &options)
	return options, err
}

var _ = []any{decodeLoose, decodeLooseNested, decodeWithDecoder, decodeLooseAnonymous, decodeScalars, decodeStrict, decodeSelf}
