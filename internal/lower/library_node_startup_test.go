package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeStartupEnvironmentWritesRefused(t *testing.T) {
	for _, source := range []string{
		`import process from 'node:process'; process.env.ADAMIC_HOST='value';`,
		`import process from 'node:process'; const key='ADAMIC_HOST'; process.env[key]='value';`,
		`import process from 'node:process'; delete process.env.ADAMIC_HOST;`,
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || refused.What != "process.env mutation" {
			t.Fatalf("want environment mutation refused, got %v", err)
		}
	}
}

func TestNodeStartupExitOverloadsRefused(t *testing.T) {
	for _, source := range []string{
		`import process from 'node:process'; process.exit('2');`,
		`import process from 'node:process'; process.exitCode='2';`,
	} {
		_, err := lowerSource(t, source)
		var missing *NotYet
		if !errors.As(err, &missing) || !strings.Contains(missing.What, "non-numeric code") {
			t.Fatalf("want unsupported Node string exit overload refused before emission, got %v", err)
		}
	}
}
