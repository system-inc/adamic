package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"strings"
	"testing"
)

// V4's .a rule refuses an unproven escape; V2's exact certified controls remain admitted.
func v4UnionReadRefusal(t *testing.T, path, field string) {
	t.Helper()
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "true\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node: %#v", truth)
	}
	_, err := lowered(t, path)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.HasPrefix(refusal.Where, path+":") || refusal.What != "an unproven escaping callable read of "+field || refusal.Fix != "prove the producer parameter and result relation before this read, or call the member directly through the view" {
		t.Fatalf("expected ruled escape refusal with path and fix: %v", err)
	}
}
