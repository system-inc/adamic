// Package optionreads is planted for TestTheOptionReadGuardSeesAnUnreadField: each field is one shape
// the guard must report as unread or must see read, named for which.
package optionreads

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
)

type plantedOptions struct {
	Read              bool   `json:"read"`
	NeverMentioned    bool   `json:"neverMentioned"`
	OnlyDefaulted     string `json:"onlyDefaulted"`
	DefaultedThenRead string `json:"defaultedThenRead"`
	OnlyAssigned      int    `json:"onlyAssigned"`
	Skipped           bool   `json:"-"`
	unexported        bool
	Nested            plantedNested       `json:"nested"`
	Entries           []plantedListEntry  `json:"entries"`
	Self              plantedSelfDecoding `json:"self"`
	plantedEmbedded
}

type plantedNested struct {
	DeepRead   bool `json:"deepRead"`
	DeepUnread bool `json:"deepUnread"`
}

type plantedEmbedded struct {
	PromotedRead   bool `json:"promotedRead"`
	PromotedUnread bool `json:"promotedUnread"`
}

type plantedListEntry struct {
	EntryRead   string `json:"entryRead"`
	EntryUnread string `json:"entryUnread"`
}

// plantedSelfDecoding decodes itself, and its field is still one a config sets.
type plantedSelfDecoding struct {
	Mode string
}

func (self *plantedSelfDecoding) UnmarshalJSON(raw []byte) error {
	return json.Unmarshal(raw, &self.Mode)
}

var decodePlanted = rule.DecodeOptionsInto[plantedOptions]()

func resolvePlanted(options plantedOptions) plantedOptions {
	if options.OnlyDefaulted == "" {
		options.OnlyDefaulted = "all"
	}
	if options.DefaultedThenRead == "" {
		options.DefaultedThenRead = "all"
	}
	options.OnlyAssigned = 3
	options.unexported = true
	return options
}

func runPlanted(options plantedOptions) int {
	options = resolvePlanted(options)
	count := 0
	if options.Read && options.Nested.DeepRead && options.PromotedRead && options.DefaultedThenRead == "local" && options.Self.Mode != "" {
		count++
	}
	for _, entry := range options.Entries {
		count += len(entry.EntryRead)
	}
	return count
}
