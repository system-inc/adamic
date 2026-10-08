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

func TestNodeStartupCommonJSRefusals(t *testing.T) {
	for _, source := range []string{
		`import type {} from 'node:process'; const fs=require('node:fs');`,
		`import type {} from 'node:process'; const name='node:fs'; const fs: typeof import('node:fs')=require(name);`,
		`import type {} from 'node:process'; const fs: typeof import('node:fs')=require('source-map-support');`,
		`import type {} from 'node:process'; const fs: typeof import('node:fs')=require('node:fs'); console.log(String(fs===fs));`,
	} {
		_, err := lowerSource(t, source)
		if err == nil {
			t.Fatal("unsupported CommonJS request or namespace observation was accepted")
		}
	}
}

func TestNodeStartupStdoutReflectionRefused(t *testing.T) {
	for _, source := range []string{
		`import process from 'node:process'; Object.keys(process.stdout);`,
		`import process from 'node:process'; const stream=process.stdout; JSON.stringify(stream);`,
	} {
		_, err := lowerSource(t, source)
		if err == nil {
			t.Fatal("opaque stdout reflection was accepted")
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
