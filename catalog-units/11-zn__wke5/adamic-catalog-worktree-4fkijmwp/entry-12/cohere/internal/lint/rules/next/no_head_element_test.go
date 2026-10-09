package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestNoHeadElementReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			name:     "paired head",
			fileName: "pages/Index.tsx",
			source:   `export const C = () => <head><title>x</title></head>;`,
		},
		{
			name:     "self closing head",
			fileName: "components/Shell.tsx",
			source:   `export const C = () => <div><head /></div>;`,
		},
		{
			// The exemption is a substring test on the whole path, so a directory merely named
			// something-app does not exempt. Pinned because tightening it to a segment test would
			// silently change which files run.
			name:     "a directory whose name ends in app",
			fileName: "src/myapp2/Shell.tsx",
			source:   `export const C = () => <head><title>x</title></head>;`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoHeadElement, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoHeadElement.Id)
		})
	}
}

func TestNoHeadElementIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// The App Router owns the document shell, so upstream skips the whole file.
			name:     "inside an app directory",
			fileName: "app/(os-layout)/Page.tsx",
			source:   `export const C = () => <head><title>x</title></head>;`,
		},
		{
			name:     "app directory deeper in the path",
			fileName: "projects/www/app/Shell.tsx",
			source:   `export const C = () => <head><title>x</title></head>;`,
		},
		{
			// The Head component is the replacement the rule pushes toward and must never itself
			// report. Capitalisation is the whole difference, so this is the case a text-only
			// comparison would get wrong.
			name:     "the next/head component",
			fileName: "components/Shell.tsx",
			source: `import Head from 'next/head';
export const C = () => <Head><title>x</title></Head>;`,
		},
		{
			name:     "no head at all",
			fileName: "components/Shell.tsx",
			source:   `export const C = () => <div><h1>x</h1></div>;`,
		},
		{
			name:     "member expression name",
			fileName: "components/Shell.tsx",
			source:   `export const C = () => <Foo.head />;`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoHeadElement, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}
