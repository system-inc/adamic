package ir

import "encoding/json"

// JSONDecodeSchema is a graph, so recursive data types need no runtime type expansion.
// Only the type descriptor is shared with Node; its parser and validator are independent.
type JSONDecodeSchema struct {
	Nodes []JSONDecodeNode `json:"nodes"`
	Root  int              `json:"root"`
}
type JSONDecodeNode struct {
	Kind              string            `json:"kind"`
	Of                Type              `json:"of"`
	Expected          string            `json:"expected"`
	Literal           string            `json:"-"`
	LiteralUnits      []uint16          `json:"literalUnits,omitempty"`
	NumberText        string            `json:"numberText,omitempty"`
	Number            float64           `json:"-"`
	Boolean           bool              `json:"boolean,omitempty"`
	Children          []int             `json:"children,omitempty"`
	Fields            []JSONDecodeField `json:"fields"`
	DiscriminantUnits []uint16          `json:"discriminantUnits,omitempty"`
	Discriminant      string            `json:"discriminant,omitempty"`
}

// Every descriptor has a fields array. Empty tuples and objects use [], never null or omission.
func (node JSONDecodeNode) MarshalJSON() ([]byte, error) {
	type descriptor JSONDecodeNode
	value := descriptor(node)
	if value.Fields == nil {
		value.Fields = []JSONDecodeField{}
	}
	return json.Marshal(value)
}

type JSONDecodeField struct {
	Name      string   `json:"name"`
	NameUnits []uint16 `json:"nameUnits,omitempty"`
	Node      int      `json:"node"`
	Optional  bool     `json:"optional,omitempty"`
}
type JSONDecode struct {
	Text   Expression
	Schema JSONDecodeSchema
}

func (JSONDecode) Type() Type { return Object }

// JSONLiteralUnits reads the checker's WTF-8 without replacing lone UTF-16 surrogates.
func JSONLiteralUnits(text string) []uint16 {
	units := []uint16{}
	for i := 0; i < len(text); {
		lead := text[i]
		size := 1
		point := uint32(lead)
		switch {
		case lead < 0x80:
		case lead < 0xe0:
			size = 2
			point = uint32(lead&31)<<6 | uint32(text[i+1]&63)
		case lead < 0xf0:
			size = 3
			point = uint32(lead&15)<<12 | uint32(text[i+1]&63)<<6 | uint32(text[i+2]&63)
		default:
			size = 4
			point = uint32(lead&7)<<18 | uint32(text[i+1]&63)<<12 | uint32(text[i+2]&63)<<6 | uint32(text[i+3]&63)
		}
		i += size
		if point > 0xffff {
			point -= 0x10000
			units = append(units, uint16(0xd800+(point>>10)), uint16(0xdc00+(point&1023)))
		} else {
			units = append(units, uint16(point))
		}
	}
	return units
}
