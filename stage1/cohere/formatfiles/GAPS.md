# What cohere's format walk needed that Adamic doesn't have yet

The port beside this file is cohere's `internal/format/formatfiles` (enumerate.go): the walk that decides which files a tree offers the formatter, `Enumerate`, with `NestedRepositoriesBelow`, `HasOwnRepository`, `NestedRepositoryContaining` and `repositoryBoundaryBetween`, written as 0.1 Adamic over the file system 'adamic' opened this afternoon (`readDirectory`, `fileStatus`, `readTextFile`, on cloud/collections-fs, merged into this branch at 1b10123). It is stage 1's fifth slice, and the first to read a real disk. Git's ignore rules are the first slice's port, `../gitignore`, reading a working tree from disk (disk.ts), which that slice's GAPS.md waited for: "a WorkingTree read from disk can take the in-memory one's place, and nothing else changes". Nothing else changed: the gitignore port is imported as it is.

## Why this slice

Ahra asked for a fifth slice that walks directories or reads files, throws and catches, or uses Set.

- **format/formatfiles** (chosen): a real walk, the one `cohere format` runs over every tree. It lists directories in Go's order, tells a symbolic link from what it points to, prunes ignored directories rather than filtering what's under them, refuses a nested repository, and reads ignore files as it enters each directory, through cohere's gitignore matcher, which stage 1 already has. Its tests build trees on disk with two helpers the test here calls itself.
- **Exceptions** (not chosen): the walk is the larger test of what's new. cohere's Go has no exceptions, so a port that throws would be shaped by the language rather than by the Go; the slices so far have each carried the Go's errors as values, and that held again here.
- **Set** (not chosen as the reason): nothing in this walk is a set in the Go.

What the port leaves out, and why: `Enumerate` starts with `formatoptions.Resolve(root)`, which walks up from the root reading `CohereSettings.json` files as JSON, and 0.1 has no JSON. The port is given Resolve's answer (the house ignore list, whether one was declared, and the settings file governing the root); the test's Go side calls Resolve on each tree and writes its answer into the cases. And the `ignorePatterns` layer matches lint's globs (`internal/lint/configuration`), which this slice doesn't carry: a tree whose settings name none is ported whole (the layer is counted, and covers nothing), and one whose settings name any is left out of the cases and counted (5 of cohere's 18 trees; the port would panic on one, saying so).

## 1. No `lstat`, and no `readlink`

Not a stage 0 gap: a gap in 'adamic''s file system. `fileStatus` follows symbolic links, and says of one it can't follow what it would say of nothing: `no such file` for a link to nothing, `failed` for a link round a loop (ELOOP). Go's `os.Lstat` sees the link itself. The walk needs that: a link to nothing is still a file the walk counts (`Walked`, `SymbolicLinks`), a `.git` that is a link to nothing makes the gitignore matcher refuse a directory as a nested repository, and a `.prettierignore` that is one is refused at the root.

There's no smallest program: nothing in 'adamic' makes a symbolic link, so a program can't build the case it needs. The test's trees are what show it.

**Around it:** a name its directory lists (`readDirectory`) that `fileStatus` can't follow, for any reason but permission denied, is a symbolic link that can't be followed (`lstat`, disk.ts). Permission denied is the directory refusing the look at all, which refuses `os.Lstat` too. And since there is no `readlink`, a link's gitignore entry names a stand-in sibling whose entry is what `fileStatus` and `readTextFile` find through the link, which is what `os.Stat` and `os.ReadFile` find.

The port's first try at this treated only `no such file` as a link, and the cases found it: a generated tree held `deep/linked -> linked`, a link to itself, which `fileStatus` calls `failed`. All three of the port's runs agreed with each other and not with Go cohere (`walked 9`, Go `walked 10`). Both rules are mutants in the test now, and both are caught.

What `lstat` and `readlink` in 'adamic' would remove: the extra `readDirectory` of a link's directory, and the stand-in targets.

## What lowered as written

Everything else, on the first try but one. Stage 0 refused the driver's resolution object, `{ ..., ignorePatterns: [] }`, as an array of never; that was the object having no declared type, so the checker itself typed `[]` as `never[]` (the values slice's gap 5 is the same root), and typing it as a `Resolution` lowers. What went through: two walks written as Go defines them, `filepath.Walk` and `filepath.WalkDir`, recursive, passing a callback (a closure over the enumeration, its maps and its scopes) and a `WalkStep` union back up; a class with mutable counters and lists; `Map<string, number>` counts and a `Map<string, Matcher>` of entered scopes, written as the walk goes; the gitignore slice's `Matcher` and `Patterns` driven over a working tree whose `ReadonlyMap` this slice keeps filling as the walk enters directories; `readDirectory`, `fileStatus` and `readTextFile` on real paths, names with spaces, newlines, tabs, quotes, DEL, accents decomposed and composed, and characters outside the Basic Multilingual Plane included; and a sort with a comparator.

The cycle finder had nothing to say: a `DiskTree` holds a map of entries, which hold strings; the scopes map holds matchers, which hold the tree; nothing reaches back.

## Go's library, written as Go answers it (golang.ts)

Two places where JavaScript's answer differs from Go's, and the port answers as Go does, each with a mutant that answers as JavaScript does and is caught:

- **Lowercasing an extension.** Go's `strings.ToLower` maps each character on its own with Unicode's simple mapping; JavaScript's `toLowerCase` uses the full mapping in context. They differ on U+0130 (İ, which JavaScript makes `i` and a combining dot) and on Σ ending a word (JavaScript's final ς). Both sides read Unicode 17.0 (Go's `unicode.Version` and Node's `process.versions.unicode`, measured), so lowering one character at a time, with U+0130 to `i`, is Go's answer exactly. A file `h.İ` is declined as `.i`, and `i.ΑΣ` as `.ασ`.
- **Ordering strings.** Go's sorted map keys are in UTF-8 byte order, which is code point order; JavaScript's `<` is UTF-16 unit order, which puts a supplementary character (its high surrogate, U+D800 to U+DBFF) before U+E000 to U+FFFF. A tree declining `.ａ` (U+FF41) and `.😀` lists them in Go's order only with `compareCodePoints`.

`readDirectory`'s names come in libuv's order, by their bytes, which is Go's `os.ReadDir` order, so the walks visit entries in the same order without sorting.

## What the cases reach

The 13 trees cohere's tests build with `writeTree` and `settingsTree` from literals (read out of enumerate_test.go with go/ast, built with cohere's own helpers, and copied, links kept as links, to where the port can read them), and 400 generated trees whose pieces start at every edge of what the walk decides on: extensions in every case and with Go's and JavaScript's lowercasing apart, extensions whose UTF-8 and UTF-16 orders differ, a name ending in a dot and one with none, the house list's names and near misses (`x.db-wa`), hidden names, names the ignore files name and negate, names that sort at the edges, names the output escapes (DEL, U+001F, a tab, a newline, a quote, a backslash) and those just inside; nested repositories as a clone's `.git` directory, a submodule's `.git` file, a `.git` link to a directory, and a `.git` link to nothing; symbolic links to a file, to a directory, to nothing, out of the tree, round a loop and to another link; a root `.git` with an exclude file; and, rarely, since each refuses the whole walk, a `.gitignore` that is a link and a `.prettierignore` that is a file or a link to nothing. Each tree is asked whether each of its paths, its root, the root's parent, `/` and a relative path holds a repository, lies in a nested one, or has one between it and the root.

Nine mutants, each caught natively and on Node: JavaScript's lowercasing, UTF-16 order for the declined extensions, a link round a loop taken for nothing (the port's own first mistake), a link to nothing taken for nothing, a trailing dot no extension, a `.git` link to nothing taken for a repository, a `.prettierignore` link to nothing let through, and the quoting's fast path missing DEL or a quote. A tenth, `rel` answering a relative path as though it lay below the root, is left out: it changes an answer only when a directory relative to the process's working directory holds a `.git`, and the two sides run in different working directories, so no case both share can show it.

## Not a gap here: Node's writev at millions of lines

The suppression slice's GAPS.md has Node 24.21.0 failing a `writev` with EINVAL when it writes millions of short lines into a pipe Go reads, at 2.26 million lines there. This test can't reach it: its 413 trees answer in 21,779 lines, and the 3,013-tree run timed below in 162,350.

## Performance, observed (not refusals)

On 3,013 trees (13 of cohere's and 3,000 generated, seed 20261005), every side's output identical to Go cohere's, unsanitized `-O2`, each run three times:

| | Native | Node | Go cohere |
|---|---|---|---|
| the whole run | 2.26 to 2.27 s | 3.56 to 3.65 s | 1.16 to 1.19 s |

Go cohere's time is the same calls, Enumerate and the rest, on the same trees, with every answer formatted into memory, not written out. Native is faster than Node, as in the suppression slice. It was 3.3 s before the driver's `quote` wrote a string with nothing to escape whole (29% of the run under callgrind, visiting every character of every absolute path), and 2.7 s before the gitignore slice's `clean` returned an already-clean path as itself (40% of the run; the gitignore slice's GAPS.md, "Path cleaning", has how, and why Go's own way of writing it was slower here).

Not system calls: on 145 of the trees, strace counted 23,122 calls natively against Go's 40,016 (26,053 of them file status calls, most of them `formatoptions.Resolve` probing for settings files, which the port is given the answer of). What is left is the walk's string work: `clean`'s scans, a third of the instructions, most of it `includes`, which compares at every position with a call to `memcmp`; and building paths a piece at a time.

## Not covered

- Permission errors. The tests run as root, which reads every directory, so a directory the walk can't list or enter, and a link whose target refuses `stat`, aren't among the cases; the port's reading of them (a listed name `fileStatus` refuses with permission denied is not there, as `os.Lstat` would refuse it) is from the documentation, not measured.
- A file that can't be read as an ignore file, and an ignore file larger than 100 MiB (the gitignore slice covers the second in memory).
- File names that aren't UTF-8. Node decodes them, and 'adamic' with it, as UTF-8 with replacement characters; Go keeps their bytes. No case has one.
- Trees whose settings name ignorePatterns globs, and every settings file `Resolve` reads, since the port is given its answer.
- macOS, where names are normalized by the file system and links and listings behave differently.
