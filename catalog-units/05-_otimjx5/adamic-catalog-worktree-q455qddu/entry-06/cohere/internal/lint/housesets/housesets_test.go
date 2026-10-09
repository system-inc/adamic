package housesets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/types/program"
)

// Detection is from the program, per file (#j6p9t1g): each case is a project on disk, built the way a run
// builds it, so the imports and the JSX fact are the compiler's own.

func buildProject(t *testing.T, files map[string]string) (string, *program.Graph) {
	t.Helper()
	root := t.TempDir()
	if _, written := files["tsconfig.json"]; !written {
		files["tsconfig.json"] = `{"compilerOptions": {"jsx": "preserve", "strict": true, "noEmit": true, "skipLibCheck": true}, "include": ["**/*"]}`
	}
	for name, contents := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json"), CurrentDirectory: root})
	if err != nil {
		t.Fatal(err)
	}
	return root, graph
}

func TestReactIsDetectedFromImportsAndJsxNotFromPackageJson(t *testing.T) {
	t.Parallel()
	root, graph := buildProject(t, map[string]string{
		// react is a dependency, and that alone applies nothing.
		"package.json":             `{"dependencies": {"react": "19.2.0", "next": "16.0.0"}, "scripts": {"react": "echo react"}}`,
		"source/Imports.tsx":       "import { useState } from 'react';\nexport function useCount() { return useState(0); }\n",
		"source/Dom.ts":            "import { createRoot } from 'react-dom/client';\nexport const make = createRoot;\n",
		"source/JsxOnly.tsx":       "export function Badge() { return <span>new</span>; }\n",
		"source/NoJsx.tsx":         "export const answer = 42;\n",
		"server/index.ts":          "export function handle(request: Request): Response { return new Response(request.url); }\n",
		"source/Reactive.ts":       "import { reactive } from './reactive-store';\nexport const store = reactive;\n",
		"source/reactive-store.ts": "export const reactive = 1;\n",
	})
	detection := Detect(graph.ProjectFiles(), root, graph.Program.Host().FS(), nil, "")

	want := map[string]bool{"source/Imports.tsx": true, "source/Dom.ts": true, "source/JsxOnly.tsx": true}
	for _, file := range graph.ProjectFiles() {
		relative, _ := filepath.Rel(root, file.FileName().AsString())
		if detection.ReactFiles[file.FileName().AsString()] != want[relative] {
			t.Errorf("%s: react=%v, want %v", relative, detection.ReactFiles[file.FileName().AsString()], want[relative])
		}
	}
	// next is in package.json and imported nowhere, so no file is Next's.
	if len(detection.NextFiles) != 0 {
		t.Errorf("next applied with no import of it: %v", detection.NextFiles)
	}
}

func TestNextIsDetectedFromImportsAndItsOwnFilesOnlyInANextProgram(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"app/layout.tsx":                 "import type { Metadata } from 'next';\nexport const metadata: Metadata = {};\n",
		"app/page.tsx":                   "export default function Page() { return <main />; }\n",
		"app/api/health/route.ts":        "export function GET() { return new Response('ok'); }\n",
		"middleware.ts":                  "export function middleware() {}\n",
		"source/Link.tsx":                "import Link from 'next/link';\nexport const Home = () => <Link href=\"/\">home</Link>;\n",
		"server/worker.ts":               "export const port = 8080;\n",
		"node_modules/next/index.d.ts":   "export interface Metadata { title?: string }\n",
		"node_modules/next/link.d.ts":    "declare const Link: (properties: { href: string; children?: unknown }) => unknown;\nexport default Link;\n",
		"node_modules/next/package.json": `{"name": "next", "types": "index.d.ts"}`,
	}
	root, graph := buildProject(t, files)
	detection := Detect(graph.ProjectFiles(), root, graph.Program.Host().FS(), nil, "")
	want := map[string]bool{"app/layout.tsx": true, "app/page.tsx": true, "app/api/health/route.ts": true, "middleware.ts": true, "source/Link.tsx": true}
	for _, file := range graph.ProjectFiles() {
		relative, _ := filepath.Rel(root, file.FileName().AsString())
		if strings.HasPrefix(relative, "node_modules") {
			continue
		}
		if detection.NextFiles[file.FileName().AsString()] != want[relative] {
			t.Errorf("%s: next=%v, want %v", relative, detection.NextFiles[file.FileName().AsString()], want[relative])
		}
	}

	// The control: the same app/ directory in a program that never imports next is just a directory.
	delete(files, "app/layout.tsx")
	delete(files, "source/Link.tsx")
	root, graph = buildProject(t, files)
	if detection := Detect(graph.ProjectFiles(), root, graph.Program.Host().FS(), nil, ""); len(detection.NextFiles) != 0 {
		t.Errorf("an app/ directory with no import of next applied cohere:next: %v", detection.NextFiles)
	}
}

func TestTailwindIsDetectedFromTheStylesheetAndSkippedByNameWithoutOne(t *testing.T) {
	t.Parallel()
	root, graph := buildProject(t, map[string]string{
		"index.ts":        "export const x = 1;\n",
		"app/globals.css": "@import \"tailwindcss\";\n",
	})
	detection := Detect(graph.ProjectFiles(), root, graph.Program.Host().FS(), nil, "")
	if detection.TailwindEntryPoint != filepath.Join(root, "app", "globals.css") {
		t.Errorf("entry %q, want app/globals.css", detection.TailwindEntryPoint)
	}

	root, graph = buildProject(t, map[string]string{
		"index.ts":        "export const x = 1;\n",
		"app/globals.css": "body { margin: 0; }\n",
	})
	detection = Detect(graph.ProjectFiles(), root, graph.Program.Host().FS(), nil, "")
	if detection.TailwindEntryPoint != "" || !strings.Contains(detection.TailwindSkipped, "app/globals.css does not import tailwindcss") {
		t.Errorf("a plain stylesheet applied tailwind: %+v", detection)
	}

	root, graph = buildProject(t, map[string]string{"index.ts": "export const x = 1;\n"})
	detection = Detect(graph.ProjectFiles(), root, graph.Program.Host().FS(), nil, "")
	if detection.TailwindEntryPoint != "" || !strings.Contains(detection.TailwindSkipped, "no Tailwind stylesheet") {
		t.Errorf("no stylesheet must skip tailwind by name: %+v", detection)
	}
}
