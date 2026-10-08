package lower

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestNodeOSSupportedImports(t *testing.T) {
	for _, source := range []string{
		`import os from 'os'; console.log(os.homedir());`,
		`import os from 'node:os'; console.log(os.homedir && os.homedir());`,
		`import * as os from 'node:os'; console.log(os.platform()); console.log(os.EOL); console.log(os.tmpdir());`,
		`import {homedir as home} from 'node:os'; console.log(home());`,
		`import os from 'node:os'; const home=os.homedir; console.log(home()); console.log(String(home===os.homedir));`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNodeOSOtherMembersAreRefused(t *testing.T) {
	t.Parallel()
	// Derive the refusal census from the actual embedded @types/node bytes,
	// including future additions to declarations, not a manually selected list.
	data, err := os.ReadFile("../load/node_types/node_modules/@types/node/os.d.ts")
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, match := range regexp.MustCompile(`(?m)^    (?:function|const|namespace) ([A-Za-z][A-Za-z0-9]*)`).FindAllStringSubmatch(string(data), -1) {
		names[match[1]] = true
	}
	for name := range names {
		if name == "EOL" || name == "platform" || name == "homedir" || name == "tmpdir" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			source := `import * as os from 'node:os'; const unsupported=os.` + name + `;`
			_, err := lowerSource(t, source)
			var refused *NotYet
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "node:os."+name) || !strings.Contains(err.Error(), "outside the pinned TypeScript src inventory") {
				t.Fatalf("want named inventory refusal, got %v", err)
			}
		})
	}
	if len(names) < 20 {
		t.Fatalf("OS declaration census unexpectedly small: %v", names)
	}
}

func TestNodeOSCacheFallbackRemainsRefused(t *testing.T) {
	_, err := lowerSource(t, `import os from 'node:os'; console.log((os.homedir && os.homedir()) || 'fallback');`)
	var refused *NotYet
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "string and a string") {
		t.Fatalf("want generic truthy string-or refusal, got %v", err)
	}
}
