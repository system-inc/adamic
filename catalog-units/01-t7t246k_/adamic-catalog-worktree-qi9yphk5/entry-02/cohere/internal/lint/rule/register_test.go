package rule

import (
	"strings"
	"testing"
)

// The duplicate panic is the whole reason this collector is safe to fan out against, so it is
// proven to fire rather than assumed to. A guard nobody has seen fail is indistinguishable from one
// that cannot.
func TestRegisterRejectsADuplicateName(t *testing.T) {
	t.Parallel()
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("registering two rules under one name did not panic, so the second would have " +
				"silently replaced the first and the config would address only one of them")
		}
		message, isString := recovered.(string)
		if !isString || !strings.Contains(message, "two rules answer to") {
			t.Fatalf("panicked for some other reason: %v", recovered)
		}
	}()

	Register(Registration{Rule: Rule{Name: "test-duplicate-probe"}})
	Register(Registration{Rule: Rule{Name: "test-duplicate-probe"}})
}

// An empty name is unaddressable by the config, so it is refused at registration rather than
// discovered as a rule that never runs.
func TestRegisterRejectsAnEmptyName(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("registering a rule with no name did not panic")
		}
	}()

	Register(Registration{Rule: Rule{Name: ""}})
}

// Registered sorts by name, which is what lets two runs be diffed against each other.
func TestRegisteredIsSortedByName(t *testing.T) {
	t.Parallel()
	Register(
		Registration{Rule: Rule{Name: "test-order-zulu"}},
		Registration{Rule: Rule{Name: "test-order-alpha"}},
	)

	var seen []string
	for _, registration := range Registered() {
		if strings.HasPrefix(registration.Rule.Name, "test-order-") {
			seen = append(seen, registration.Rule.Name)
		}
	}

	if len(seen) != 2 {
		t.Fatalf("expected both probe rules back, got %v", seen)
	}
	if seen[0] != "test-order-alpha" || seen[1] != "test-order-zulu" {
		t.Fatalf("expected name order, got %v", seen)
	}
}
