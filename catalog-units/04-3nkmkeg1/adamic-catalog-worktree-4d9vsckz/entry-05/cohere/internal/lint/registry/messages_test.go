package registry

import (
	"testing"

	"github.com/system-inc/cohere/policy"
)

// TestEveryHouseMessageIsRenderedByARegisteredRule: with every rule package linked, each TypeScript
// message in policy/messages has a handle a rule took, each object-phrase option has a claim a rule
// took, and each handle names a registered rule. An entry no rule renders is wording nobody reads, and
// a handle on an unregistered rule renders into nothing.
func TestEveryHouseMessageIsRenderedByARegisteredRule(t *testing.T) {
	t.Parallel()
	registered := map[string]bool{}
	for _, registration := range All() {
		registered[registration.Name] = true
	}
	requested := map[policy.MessageHandle]bool{}
	for _, handle := range policy.RequestedMessages() {
		requested[handle] = true
		if !registered[handle.Rule] {
			t.Errorf("%s takes the message %q and is not a registered rule", handle.Rule, handle.Id)
		}
	}
	held := policy.Messages.MessagesFor(policy.MessageLanguageTypeScript)
	if len(held) == 0 {
		t.Fatal("the catalog holds no TypeScript message, so this test proves nothing")
	}
	for _, handle := range held {
		if !requested[handle] {
			t.Errorf("policy/messages holds %s %q, and no rule renders it", handle.Rule, handle.Id)
		}
	}
	claimed := map[policy.MessageOption]bool{}
	for _, option := range policy.RequestedOptions() {
		claimed[option] = true
	}
	for _, option := range policy.Messages.OptionsFor(policy.MessageLanguageTypeScript) {
		if !claimed[option] {
			t.Errorf("policy/messages holds the option %q of %q in %s %q, and no rule picks it", option.Name, option.Phrase, option.Rule, option.Id)
		}
	}
}
