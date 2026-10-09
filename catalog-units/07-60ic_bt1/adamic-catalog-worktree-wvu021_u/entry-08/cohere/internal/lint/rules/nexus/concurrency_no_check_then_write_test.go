package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const concurrencyNoCheckThenWriteFile = "source/Subject.ts"

// concurrencyNoCheckThenWriteNodeTypes models `@types/node` 26 as installed in ahra on 2026-10-02:
// the file system and `path` declared under `node:` module names, with `fs.promises` re-exported
// from the promise module.
const concurrencyNoCheckThenWriteNodeTypes = `
declare module 'node:fs' {
    export function existsSync(path: string): boolean;
    export function accessSync(path: string, mode?: number): void;
    export function statSync(path: string, options?: { throwIfNoEntry?: boolean }): { size: number } | undefined;
    export function writeFileSync(path: string, data: string | Uint8Array, options?: string | { flag?: string; encoding?: string }): void;
    export function copyFileSync(source: string, destination: string, mode?: number): void;
    export function renameSync(oldPath: string, newPath: string): void;
    export function createWriteStream(path: string, options?: string | { flags?: string }): { end(): void };
    export function mkdirSync(path: string, options?: { recursive?: boolean }): void;
    export const constants: { COPYFILE_EXCL: number };
    export * as promises from 'node:fs/promises';
}
declare module 'node:fs/promises' {
    export function access(path: string, mode?: number): Promise<void>;
    export function stat(path: string): Promise<{ size: number }>;
    export function writeFile(path: string, data: string | Uint8Array, options?: string | { flag?: string }): Promise<void>;
    export function copyFile(source: string, destination: string, mode?: number): Promise<void>;
    export function rename(oldPath: string, newPath: string): Promise<void>;
    export function mkdir(path: string, options?: { recursive?: boolean }): Promise<number>;
}
declare module 'node:path' {
    namespace path {
        function join(...paths: string[]): string;
        function resolve(...paths: string[]): string;
    }
    export = path;
}
`

// concurrencyNoCheckThenWriteOlderNodeTypes models `@types/node` 20, which declares the API under `fs`
// and re-exports it as `node:fs`, the other way round from 26.
const concurrencyNoCheckThenWriteOlderNodeTypes = `
declare module 'fs' {
    export function existsSync(path: string): boolean;
    export function writeFileSync(path: string, data: string | Uint8Array): void;
}
declare module 'node:fs' {
    export * from 'fs';
}
declare module 'path' {
    namespace path {
        interface PlatformPath {
            join(...paths: string[]): string;
        }
    }
    const path: path.PlatformPath;
    export = path;
}
declare module 'node:path' {
    import path = require('path');
    export = path;
}
`

// concurrencyNoCheckThenWriteSharpPackage models sharp 0.35's ESM declarations as installed in ahra.
var concurrencyNoCheckThenWriteSharpPackage = map[string]string{
	"node_modules/sharp/package.json": `{"name": "sharp", "version": "0.35.3", "types": "index.d.ts"}`,
	"node_modules/sharp/index.d.ts": `
export interface Sharp {
    webp(options?: { quality?: number }): Sharp;
    toFile(fileOut: string): Promise<{ size: number }>;
}
declare function sharp(input?: string): Sharp;
export default sharp;
`,
}

func concurrencyNoCheckThenWriteSource(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

func concurrencyNoCheckThenWriteRun(t *testing.T, nodeTypes string, sourceText string) rule_testing.Result {
	t.Helper()
	files := map[string]string{
		"types/node.d.ts":               nodeTypes,
		concurrencyNoCheckThenWriteFile: sourceText,
	}
	for name, contents := range concurrencyNoCheckThenWriteSharpPackage {
		files[name] = contents
	}
	return rule_testing.RunTypedFiles(t, ConcurrencyNoCheckThenWrite, files, concurrencyNoCheckThenWriteFile)
}

// concurrencyNoCheckThenWriteExpect asserts the findings point at exactly these checks, in order,
// with the message naming the write.
func concurrencyNoCheckThenWriteExpect(t *testing.T, result rule_testing.Result, sourceText string, wantSpans []string, wantWrite string) {
	t.Helper()
	wantIds := make([]string, 0, len(wantSpans))
	for range wantSpans {
		wantIds = append(wantIds, concurrencyNoCheckThenWriteId)
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	for index, diagnostic := range result.Diagnostics {
		reported := sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != wantSpans[index] {
			t.Fatalf("finding %d points at\n%s\nwant\n%s", index, reported, wantSpans[index])
		}
		if diagnostic.Message.Description != concurrencyNoCheckThenWriteMessage(wantWrite).Description {
			t.Fatalf("message is %q, want the one naming `%s`", diagnostic.Message.Description, wantWrite)
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
}

// concurrencyNoCheckThenWriteMediaGeneration models `persistArtifact` in
// `modules/intelligence/IntelligenceMediaGenerationApi.ts` in ahra: before the fix in 0f33c3f0
// (`const candidate` in the loop body, an `access` loop, then a plain `writeFile`) and after it (the
// file as it stands, `:121`, `flag: 'wx'` retried on `EEXIST`). The real before-file read the
// directory and extension as `input.outputDirectory` and `input.extension`, which the rule declines
// (see TestConcurrencyNoCheckThenWriteMissesPropertyReads); here they are read into `const`s first,
// the one change, so the fixture proves the rest of the shape.
func concurrencyNoCheckThenWriteMediaGeneration(fixed bool) string {
	header := []string{
		"import * as NodeFileSystemPromises from 'node:fs/promises';",
		"import * as NodePath from 'node:path';",
		"export async function persistArtifact(input: { outputDirectory: string; bytes: Uint8Array; extension: string }): Promise<string> {",
		"    await NodeFileSystemPromises.mkdir(input.outputDirectory, { recursive: true });",
		"    const outputDirectory = input.outputDirectory;",
		"    const extension = input.extension;",
	}
	if !fixed {
		return concurrencyNoCheckThenWriteSource(append(header,
			"    let fileIndex = 0;",
			"    while(true) {",
			"        const candidate = NodePath.join(outputDirectory, `${fileIndex}.${extension}`);",
			"        try {",
			"            await NodeFileSystemPromises.access(candidate);",
			"            fileIndex++;",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"    }",
			"    const outputPath = NodePath.join(outputDirectory, `${fileIndex}.${extension}`);",
			"    await NodeFileSystemPromises.writeFile(outputPath, input.bytes);",
			"    return outputPath;",
			"}",
		)...)
	}
	return concurrencyNoCheckThenWriteSource(append(header,
		"    let fileIndex = 0;",
		"    let outputPath: string | undefined;",
		"    while(outputPath === undefined) {",
		"        const candidatePath = NodePath.join(outputDirectory, `${fileIndex}.${extension}`);",
		"        try {",
		"            await NodeFileSystemPromises.writeFile(candidatePath, input.bytes, { flag: 'wx' });",
		"            outputPath = candidatePath;",
		"        }",
		"        catch(error) {",
		"            if((error as { code?: string }).code !== 'EEXIST') {",
		"                throw error;",
		"            }",
		"            fileIndex++;",
		"        }",
		"    }",
		"    return outputPath;",
		"}",
	)...)
}

// concurrencyNoCheckThenWriteKlingVideo models `modules/kling/KlingApi.ts:253` in ahra on 2026-10-02:
// the `access` loop, then a whole network download, then the plain `writeFile`. Fixed, it is the same
// claim loop as the media generation fix, with the download moved before it.
func concurrencyNoCheckThenWriteKlingVideo(fixed bool) string {
	header := []string{
		"import * as NodeFileSystemPromises from 'node:fs/promises';",
		"import * as NodePath from 'node:path';",
		"declare function fetchWithCookies(url: string): Promise<{ ok: boolean; arrayBuffer(): Promise<ArrayBuffer> }>;",
		"export async function saveVideo(outputDirectory: string, firstUrl: string): Promise<string> {",
		"    await NodeFileSystemPromises.mkdir(outputDirectory, { recursive: true });",
		"",
		"    // Find next available index",
		"    let index = 0;",
	}
	if !fixed {
		return concurrencyNoCheckThenWriteSource(append(header,
			"    while(true) {",
			"        try {",
			"            await NodeFileSystemPromises.access(NodePath.join(outputDirectory, `${index}.mp4`));",
			"            index++;",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"    }",
			"",
			"    const outputPath = NodePath.join(outputDirectory, `${index}.mp4`);",
			"    const response = await fetchWithCookies(firstUrl);",
			"    if(!response.ok) {",
			"        throw new Error('Failed to download video');",
			"    }",
			"",
			"    const buffer = new Uint8Array(await response.arrayBuffer());",
			"    await NodeFileSystemPromises.writeFile(outputPath, buffer);",
			"    return outputPath;",
			"}",
		)...)
	}
	return concurrencyNoCheckThenWriteSource(append(header,
		"    const response = await fetchWithCookies(firstUrl);",
		"    if(!response.ok) {",
		"        throw new Error('Failed to download video');",
		"    }",
		"    const buffer = new Uint8Array(await response.arrayBuffer());",
		"    let outputPath: string | undefined;",
		"    while(outputPath === undefined) {",
		"        const candidatePath = NodePath.join(outputDirectory, `${index}.mp4`);",
		"        try {",
		"            await NodeFileSystemPromises.writeFile(candidatePath, buffer, { flag: 'wx' });",
		"            outputPath = candidatePath;",
		"        }",
		"        catch(error) {",
		"            if((error as { code?: string }).code !== 'EEXIST') {",
		"                throw error;",
		"            }",
		"            index++;",
		"        }",
		"    }",
		"    return outputPath;",
		"}",
	)...)
}

// concurrencyNoCheckThenWriteKlingImages models `modules/kling/KlingApi.ts:383`: the same loop inside
// the `urls.map(async ...)` callback that downloads every image at once, so the race is between
// siblings of one call.
const concurrencyNoCheckThenWriteKlingImages = `import * as NodeFileSystemPromises from 'node:fs/promises';
import * as NodePath from 'node:path';
declare function fetchWithCookies(url: string): Promise<{ ok: boolean; arrayBuffer(): Promise<ArrayBuffer> }>;
export async function saveImages(urls: string[], outputDirectory: string): Promise<string[]> {
    await NodeFileSystemPromises.mkdir(outputDirectory, { recursive: true });
    const savedPaths: string[] = [];
    const downloads = urls.map(async (url, urlIndex) => {
        let index = urlIndex;
        while(true) {
            try {
                await NodeFileSystemPromises.access(NodePath.join(outputDirectory, ` + "`${index}.png`" + `));
                index++;
            }
            catch {
                break;
            }
        }

        const outputPath = NodePath.join(outputDirectory, ` + "`${index}.png`" + `);
        const response = await fetchWithCookies(url);
        if(!response.ok) throw new Error('Failed to download image');

        const buffer = new Uint8Array(await response.arrayBuffer());
        await NodeFileSystemPromises.writeFile(outputPath, buffer);
        savedPaths.push(outputPath);
    });
    await Promise.all(downloads);
    return savedPaths;
}
`

// concurrencyNoCheckThenWriteOpenAiCommandLine models `modules/openai/OpenAiCommandLineInterface.ts:144`:
// the loop, then `copyFile` from the generator's temporary file.
const concurrencyNoCheckThenWriteOpenAiCommandLine = `import * as NodeFileSystemPromises from 'node:fs/promises';
import * as NodePath from 'node:path';
export async function save(outputDirectory: string, imageIndex: number, localPath: string): Promise<string> {
    let fileIndex = imageIndex;
    while(true) {
        try {
            await NodeFileSystemPromises.access(NodePath.join(outputDirectory, ` + "`${fileIndex}.png`" + `));
            fileIndex++;
        }
        catch {
            break;
        }
    }
    const outputPath = NodePath.join(outputDirectory, ` + "`${fileIndex}.png`" + `);
    await NodeFileSystemPromises.copyFile(localPath, outputPath);
    const { size: byteSize } = await NodeFileSystemPromises.stat(outputPath);
    return outputPath + byteSize;
}
`

// concurrencyNoCheckThenWriteArtApi models `modules/art/ArtApi.ts:231` in ahra on 2026-10-02: the
// synchronous `do ... while(existsSync(...))` for the next video name, then `writeFileSync`. Fixed,
// the loop claims the name with `flag: 'wx'`.
func concurrencyNoCheckThenWriteArtApi(fixed bool) string {
	header := []string{
		"import * as NodeFileSystem from 'node:fs';",
		"import * as NodePath from 'node:path';",
		"export function saveVideo(dayDirectory: string, imageFilename: string, videoBuffer: Uint8Array): string {",
		"    const metadataPath = NodePath.join(dayDirectory, imageFilename.replace('.png', '.json'));",
		"    if(!NodeFileSystem.existsSync(metadataPath)) {",
		"        return '';",
		"    }",
		"",
		"    // Find next available video index for this image",
		"    const baseWithoutExtension = imageFilename.replace('.png', '');",
		"    let videoIndex = 1;",
		"    let videoFilename: string;",
	}
	if !fixed {
		return concurrencyNoCheckThenWriteSource(append(header,
			"    do {",
			"        videoFilename = `${baseWithoutExtension}-v${videoIndex.toString().padStart(2, '0')}.mp4`;",
			"        videoIndex++;",
			"    } while(NodeFileSystem.existsSync(NodePath.join(dayDirectory, videoFilename)));",
			"",
			"    const videoPath = NodePath.join(dayDirectory, videoFilename);",
			"    NodeFileSystem.writeFileSync(videoPath, videoBuffer);",
			"    return videoFilename;",
			"}",
		)...)
	}
	return concurrencyNoCheckThenWriteSource(append(header,
		"    while(true) {",
		"        videoFilename = `${baseWithoutExtension}-v${videoIndex.toString().padStart(2, '0')}.mp4`;",
		"        try {",
		"            NodeFileSystem.writeFileSync(NodePath.join(dayDirectory, videoFilename), videoBuffer, { flag: 'wx' });",
		"            break;",
		"        }",
		"        catch(error) {",
		"            if((error as { code?: string }).code !== 'EEXIST') {",
		"                throw error;",
		"            }",
		"            videoIndex++;",
		"        }",
		"    }",
		"    return videoFilename;",
		"}",
	)...)
}

// concurrencyNoCheckThenWriteArtLibrary models `modules/art/ArtLibrary.ts:65`: the same loop, then
// `copyFileSync` into the library. Fixed, the copy refuses an existing file with `COPYFILE_EXCL`.
func concurrencyNoCheckThenWriteArtLibrary(fixed bool) string {
	copy := "    NodeFileSystem.copyFileSync(sourcePath, libraryPath);"
	if fixed {
		copy = "    NodeFileSystem.copyFileSync(sourcePath, libraryPath, NodeFileSystem.constants.COPYFILE_EXCL);"
	}
	return concurrencyNoCheckThenWriteSource(
		"import * as NodeFileSystem from 'node:fs';",
		"import * as NodePath from 'node:path';",
		"export function addToLibrary(artLibrary: string, sourcePath: string, dateString: string, slug: string): string {",
		"    // Find next available number",
		"    const dayDirectory = NodePath.join(artLibrary, dateString);",
		"    if(!NodeFileSystem.existsSync(dayDirectory)) {",
		"        NodeFileSystem.mkdirSync(dayDirectory, { recursive: true });",
		"    }",
		"",
		"    let number = 1;",
		"    let filename: string;",
		"    do {",
		"        filename = `${dateString}-${slug}-${number.toString().padStart(2, '0')}.png`;",
		"        number++;",
		"    } while(NodeFileSystem.existsSync(NodePath.join(dayDirectory, filename)));",
		"",
		"    const libraryPath = NodePath.join(dayDirectory, filename);",
		"",
		"    // Copy file",
		copy,
		"    return libraryPath;",
		"}",
	)
}

// concurrencyNoCheckThenWritePhiSocialLibrary models `modules/phi/social/PhiSocialLibrary.ts:210`:
// the check reads `${baseName}.webp` directly and the write reads it through two `const`s, and the
// write is sharp's `toFile`. Fixed, the encode goes to a buffer that is written with `flag: 'wx'`.
func concurrencyNoCheckThenWritePhiSocialLibrary(fixed bool) string {
	write := "    await sharp(sourcePath).webp({ quality: 90 }).toFile(libraryPath);"
	if fixed {
		write = "    NodeFileSystem.writeFileSync(libraryPath, encoded, { flag: 'wx' });"
	}
	return concurrencyNoCheckThenWriteSource(
		"import * as NodeFileSystem from 'node:fs';",
		"import * as NodePath from 'node:path';",
		"import sharp from 'sharp';",
		"export async function addToLibrary(dayDirectory: string, sourcePath: string, dateString: string, slug: string, encoded: Uint8Array): Promise<string> {",
		"    let number = 1;",
		"    let baseName: string;",
		"    do {",
		"        baseName = `${dateString}-${slug}-${number.toString().padStart(2, '0')}`;",
		"        number++;",
		"    } while(NodeFileSystem.existsSync(NodePath.join(dayDirectory, `${baseName}.webp`)));",
		"",
		"    const webpFilename = `${baseName}.webp`;",
		"    const libraryPath = NodePath.join(dayDirectory, webpFilename);",
		"    void encoded;",
		write,
		"    return libraryPath;",
		"}",
	)
}

// The real sites, before and after. Each before-shape is one finding on the check; each after-shape,
// written the way the fix in `IntelligenceMediaGenerationApi.ts` writes it, is silent.
func TestConcurrencyNoCheckThenWriteRealSites(t *testing.T) {
	t.Parallel()

	accessKling := "NodeFileSystemPromises.access(NodePath.join(outputDirectory, `${index}.mp4`))"
	cases := []struct {
		name      string
		before    string
		after     string
		wantSpan  string
		wantWrite string
	}{
		{"IntelligenceMediaGenerationApi persistArtifact", concurrencyNoCheckThenWriteMediaGeneration(false), concurrencyNoCheckThenWriteMediaGeneration(true),
			"NodeFileSystemPromises.access(candidate)", "writeFile"},
		{"KlingApi video", concurrencyNoCheckThenWriteKlingVideo(false), concurrencyNoCheckThenWriteKlingVideo(true), accessKling, "writeFile"},
		{"ArtApi video name", concurrencyNoCheckThenWriteArtApi(false), concurrencyNoCheckThenWriteArtApi(true),
			"NodeFileSystem.existsSync(NodePath.join(dayDirectory, videoFilename))", "writeFileSync"},
		{"ArtLibrary copy", concurrencyNoCheckThenWriteArtLibrary(false), concurrencyNoCheckThenWriteArtLibrary(true),
			"NodeFileSystem.existsSync(NodePath.join(dayDirectory, filename))", "copyFileSync"},
		{"PhiSocialLibrary sharp encode", concurrencyNoCheckThenWritePhiSocialLibrary(false), concurrencyNoCheckThenWritePhiSocialLibrary(true),
			"NodeFileSystem.existsSync(NodePath.join(dayDirectory, `${baseName}.webp`))", "toFile"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := concurrencyNoCheckThenWriteRun(t, concurrencyNoCheckThenWriteNodeTypes, testCase.before)
			concurrencyNoCheckThenWriteExpect(t, result, testCase.before, []string{testCase.wantSpan}, testCase.wantWrite)
			rule_testing.ExpectClean(t, concurrencyNoCheckThenWriteRun(t, concurrencyNoCheckThenWriteNodeTypes, testCase.after))
		})
	}
}

// The two other real shapes: the loop inside a parallel `map` callback, and a `copyFile` write.
func TestConcurrencyNoCheckThenWriteRealSitesWithoutAFixedCopy(t *testing.T) {
	t.Parallel()

	result := concurrencyNoCheckThenWriteRun(t, concurrencyNoCheckThenWriteNodeTypes, concurrencyNoCheckThenWriteKlingImages)
	concurrencyNoCheckThenWriteExpect(t, result, concurrencyNoCheckThenWriteKlingImages,
		[]string{"NodeFileSystemPromises.access(NodePath.join(outputDirectory, `${index}.png`))"}, "writeFile")

	result = concurrencyNoCheckThenWriteRun(t, concurrencyNoCheckThenWriteNodeTypes, concurrencyNoCheckThenWriteOpenAiCommandLine)
	concurrencyNoCheckThenWriteExpect(t, result, concurrencyNoCheckThenWriteOpenAiCommandLine,
		[]string{"NodeFileSystemPromises.access(NodePath.join(outputDirectory, `${fileIndex}.png`))"}, "copyFile")
}

// concurrencyNoCheckThenWriteShape is one function body over a directory, a counter `index` and bytes,
// with Node's API imported every way the rule must resolve.
func concurrencyNoCheckThenWriteShape(lines ...string) string {
	return concurrencyNoCheckThenWriteSource(append([]string{
		"import * as NodeFileSystem from 'node:fs';",
		"import * as NodeFileSystemPromises from 'node:fs/promises';",
		"import { access, writeFile } from 'node:fs/promises';",
		"import * as NodePath from 'node:path';",
		"declare function delay(milliseconds: number): Promise<void>;",
		"declare function later(callback: () => void): void;",
		"export async function save(directory: string, bytes: Uint8Array, options: { flag: string }, items: string[]): Promise<void> {",
		"    let index = 0;",
	}, append(lines, "}")...)...)
}

func TestConcurrencyNoCheckThenWriteFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		lines     []string
		wantSpan  string
		wantWrite string
	}{
		{"a while condition, then writeFileSync", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) {",
			"        index++;",
			"    }",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
		}, "NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))", "writeFileSync"},
		{"a for condition, with the counter in the incrementor", []string{
			"    for(; NodeFileSystem.existsSync(directory + '/' + index + '.png'); index += 1) {}",
			"    NodeFileSystem.writeFileSync(directory + '/' + index + '.png', bytes, 'utf8');",
		}, "NodeFileSystem.existsSync(directory + '/' + index + '.png')", "writeFileSync"},
		{"named imports of the promise API", []string{
			"    for(;;) {",
			"        try {",
			"            await access(`${directory}/${index}.png`);",
			"            index++;",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"    }",
			"    await writeFile(`${directory}/${index}.png`, bytes, { flag: 'w' });",
		}, "access(`${directory}/${index}.png`)", "writeFile"},
		{"fs.promises through the callback module", []string{
			"    while(true) {",
			"        try {",
			"            await NodeFileSystem.promises.stat(NodePath.join(directory, `${index}.png`));",
			"            ++index;",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"    }",
			"    await NodeFileSystem.promises.rename(NodePath.join(directory, 'temporary'), NodePath.join(directory, `${index}.png`));",
		}, "NodeFileSystem.promises.stat(NodePath.join(directory, `${index}.png`))", "rename"},
		{"a synchronous throwing check, then a stream", []string{
			"    while(true) {",
			"        try {",
			"            NodeFileSystem.accessSync(NodePath.join(directory, `${index}.log`));",
			"            index++;",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"    }",
			"    NodeFileSystem.createWriteStream(NodePath.join(directory, `${index}.log`), { flags: 'a' }).end();",
		}, "NodeFileSystem.accessSync(NodePath.join(directory, `${index}.log`))", "createWriteStream"},
		{"the write in a branch after the loop", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    if(bytes.length > 0) {",
			"        NodeFileSystem.renameSync(NodePath.join(directory, 'temporary'), NodePath.join(directory, `${index}.png`));",
			"    }",
		}, "NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))", "renameSync"},
		{"the loop in a branch, the write after it", []string{
			"    if(bytes.length > 0) {",
			"        while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    }",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
		}, "NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))", "writeFileSync"},
		{"a counter written again inside the loop before the check, the write after the loop", []string{
			"    while(true) {",
			"        index++;",
			"        try {",
			"            await access(NodePath.join(directory, `${index}.png`));",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"    }",
			"    await writeFile(NodePath.join(directory, `${index}.png`), bytes);",
		}, "access(NodePath.join(directory, `${index}.png`))", "writeFile"},
		{"an await between the loop and the write, a write in a nested loop after it", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    await delay(10);",
			"    for(const item of items) {",
			"        NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), item);",
			"    }",
		}, "NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))", "writeFileSync"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceText := concurrencyNoCheckThenWriteShape(testCase.lines...)
			result := concurrencyNoCheckThenWriteRun(t, concurrencyNoCheckThenWriteNodeTypes, sourceText)
			concurrencyNoCheckThenWriteExpect(t, result, sourceText, []string{testCase.wantSpan}, testCase.wantWrite)
		})
	}
}

// `@types/node` 20 declares the API under `fs` and re-exports it as `node:fs`; the same import still
// resolves to Node's declaration.
func TestConcurrencyNoCheckThenWriteFiresOnOlderNodeTypes(t *testing.T) {
	t.Parallel()

	sourceText := concurrencyNoCheckThenWriteSource(
		"import * as NodeFileSystem from 'node:fs';",
		"import * as NodePath from 'node:path';",
		"export function save(directory: string, bytes: Uint8Array): void {",
		"    let index = 0;",
		"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
		"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
		"}",
	)
	result := concurrencyNoCheckThenWriteRun(t, concurrencyNoCheckThenWriteOlderNodeTypes, sourceText)
	concurrencyNoCheckThenWriteExpect(t, result, sourceText,
		[]string{"NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))"}, "writeFileSync")
}

func TestConcurrencyNoCheckThenWriteStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"a wait for a lock to go away, no counter", []string{
			"    const lockPath = NodePath.join(directory, 'lock');",
			"    while(NodeFileSystem.existsSync(lockPath)) {",
			"        await delay(100);",
			"    }",
			"    NodeFileSystem.writeFileSync(lockPath, bytes);",
		}},
		{"the write claims with wx", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes, { flag: 'wx' });",
		}},
		{"options the source does not show", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes, options);",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes, { ...options });",
		}},
		{"a stream opened with wx", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.log`))) index++;",
			"    NodeFileSystem.createWriteStream(NodePath.join(directory, `${index}.log`), { flags: 'wx' }).end();",
		}},
		{"a copy with a mode", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    NodeFileSystem.copyFileSync('source.png', NodePath.join(directory, `${index}.png`), NodeFileSystem.constants.COPYFILE_EXCL);",
		}},
		{"a write to a different path", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.json`), bytes);",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}`, 'png'), bytes);",
		}},
		{"the counter moves between the loop and the write", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    index++;",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
		}},
		{"the counter is written from a nested function", []string{
			"    later(() => { index = 0; });",
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
		}},
		{"the write in a callback after the loop", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    later(() => NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes));",
		}},
		{"the write before the loop", []string{
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
		}},
		{"the write in the other branch", []string{
			"    if(bytes.length > 0) {",
			"        while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    }",
			"    else {",
			"        NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
			"    }",
		}},
		{"the write past an enclosing loop", []string{
			"    for(const item of items) {",
			"        while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"        void item;",
			"    }",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
		}},
		{"the loop stops on a name that exists", []string{
			"    while(!NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
		}},
		{"a check that does not throw on a missing name", []string{
			"    while(true) {",
			"        try {",
			"            NodeFileSystem.statSync(NodePath.join(directory, `${index}.png`), { throwIfNoEntry: false });",
			"            index++;",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"    }",
			"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
		}},
		{"a promise check that is not awaited", []string{
			"    while(true) {",
			"        try {",
			"            access(NodePath.join(directory, `${index}.png`));",
			"            index++;",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"    }",
			"    await writeFile(NodePath.join(directory, `${index}.png`), bytes);",
		}},
		{"a catch that does more than break", []string{
			"    while(true) {",
			"        try {",
			"            await access(NodePath.join(directory, `${index}.png`));",
			"            index++;",
			"        }",
			"        catch {",
			"            index++;",
			"            break;",
			"        }",
			"    }",
			"    await writeFile(NodePath.join(directory, `${index}.png`), bytes);",
		}},
		{"a try with a finally", []string{
			"    while(true) {",
			"        try {",
			"            await access(NodePath.join(directory, `${index}.png`));",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"        finally {",
			"            index++;",
			"        }",
			"    }",
			"    await writeFile(NodePath.join(directory, `${index}.png`), bytes);",
		}},
		{"something else in the try that can throw", []string{
			"    while(true) {",
			"        try {",
			"            await access(NodePath.join(directory, `${index}.png`));",
			"            await delay(10);",
			"            index++;",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"    }",
			"    await writeFile(NodePath.join(directory, `${index}.png`), bytes);",
		}},
		{"a loop-body const read before the counter moved", []string{
			"    while(true) {",
			"        const candidate = NodePath.join(directory, `${index}.png`);",
			"        index++;",
			"        try {",
			"            await access(candidate);",
			"        }",
			"        catch {",
			"            break;",
			"        }",
			"    }",
			"    await writeFile(NodePath.join(directory, `${index}.png`), bytes);",
		}},
		{"a const from before the loop, spelled out on the other side", []string{
			"    const first = NodePath.join(directory, `${index}.png`);",
			"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.png`))) index++;",
			"    NodeFileSystem.writeFileSync(first, bytes);",
		}},
		{"a path read through a property", []string{
			"    while(NodeFileSystem.existsSync(NodePath.join(options.flag, `${index}.png`))) index++;",
			"    NodeFileSystem.writeFileSync(NodePath.join(options.flag, `${index}.png`), bytes);",
		}},
		{"path.resolve, which reads the working directory", []string{
			"    while(NodeFileSystem.existsSync(NodePath.resolve(directory, `${index}.png`))) index++;",
			"    NodeFileSystem.writeFileSync(NodePath.resolve(directory, `${index}.png`), bytes);",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := concurrencyNoCheckThenWriteRun(t, concurrencyNoCheckThenWriteNodeTypes, concurrencyNoCheckThenWriteShape(testCase.lines...))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// A project's own `existsSync` is not Node's, whatever its name.
func TestConcurrencyNoCheckThenWriteStaysSilentOnALookalikeCheck(t *testing.T) {
	t.Parallel()

	sourceText := concurrencyNoCheckThenWriteSource(
		"import * as NodeFileSystem from 'node:fs';",
		"import * as NodePath from 'node:path';",
		"declare function existsSync(path: string): boolean;",
		"export function save(directory: string, bytes: Uint8Array): void {",
		"    let index = 0;",
		"    while(existsSync(NodePath.join(directory, `${index}.png`))) index++;",
		"    NodeFileSystem.writeFileSync(NodePath.join(directory, `${index}.png`), bytes);",
		"}",
	)
	rule_testing.ExpectClean(t, concurrencyNoCheckThenWriteRun(t, concurrencyNoCheckThenWriteNodeTypes, sourceText))
}

// A `toFile` that is not sharp's is not matched, even in a file that imports sharp.
func TestConcurrencyNoCheckThenWriteStaysSilentOnALookalikeToFile(t *testing.T) {
	t.Parallel()

	sourceText := concurrencyNoCheckThenWriteSource(
		"import * as NodeFileSystem from 'node:fs';",
		"import * as NodePath from 'node:path';",
		"import sharp from 'sharp';",
		"declare const image: { toFile(path: string): Promise<void> };",
		"export async function save(directory: string): Promise<void> {",
		"    let index = 0;",
		"    while(NodeFileSystem.existsSync(NodePath.join(directory, `${index}.webp`))) index++;",
		"    await image.toFile(NodePath.join(directory, `${index}.webp`));",
		"    void sharp;",
		"}",
	)
	rule_testing.ExpectClean(t, concurrencyNoCheckThenWriteRun(t, concurrencyNoCheckThenWriteNodeTypes, sourceText))
}

// The real before-file of `persistArtifact`, as it stood before 0f33c3f0, is a missed finding by
// decision: its path reads `input.outputDirectory` and `input.extension`, and a property can be
// changed by anything else holding `input` during the `await` between the check and the write, so the
// rule cannot prove the two paths equal. It fires once the fields are read into locals (the fixture
// above). Reported beside the measurement as the gap between the brief's nine sites and eight.
func TestConcurrencyNoCheckThenWriteMissesPropertyReads(t *testing.T) {
	t.Parallel()

	sourceText := concurrencyNoCheckThenWriteSource(
		"import * as NodeFileSystemPromises from 'node:fs/promises';",
		"import * as NodePath from 'node:path';",
		"export async function persistArtifact(input: { outputDirectory: string; buffer: Uint8Array; extension: string }): Promise<string> {",
		"    await NodeFileSystemPromises.mkdir(input.outputDirectory, { recursive: true });",
		"",
		"    let fileIndex = 0;",
		"    while(true) {",
		"        const candidate = NodePath.join(input.outputDirectory, `${fileIndex}.${input.extension}`);",
		"        try {",
		"            await NodeFileSystemPromises.access(candidate);",
		"            fileIndex++;",
		"        }",
		"        catch {",
		"            break;",
		"        }",
		"    }",
		"    const outputPath = NodePath.join(input.outputDirectory, `${fileIndex}.${input.extension}`);",
		"    await NodeFileSystemPromises.writeFile(outputPath, input.buffer);",
		"    return outputPath;",
		"}",
	)
	rule_testing.ExpectClean(t, concurrencyNoCheckThenWriteRun(t, concurrencyNoCheckThenWriteNodeTypes, sourceText))
}
