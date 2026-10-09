# cohere for VS Code

Format on save with cohere, replacing the Prettier extension. A save runs
`cohere --fix --format --stdin-filepath <file>` on the buffer and applies what comes back, which is
exactly what `cohere --fix --format <file>` would write for the same text: the same lint fixes, the
same printers, the same options and ignore layers. The editor and the gate cannot disagree.

The edit is applied before the save lands, through VS Code's formatter API, so the document does not
reload and undo works as usual.

## What a save does

- **Formats and fixes:** the buffer comes back fixed and formatted.
- **Leaves alone, silently:** a file cohere declines comes back as typed. That covers a type with no
  printer, a file outside the format scope, and source that does not parse; the gate reports a typo,
  the save does not.
- **Shows an error:** a broken formatter, a project cohere cannot load, or a missing binary shows an
  error, and the buffer saves as typed.

## Which cohere

The same one `s c` runs, found in Structure's order:

1. `COHERE_BINARY`.
2. The project's `node_modules/.bin/cohere`.
3. `cohere` on `PATH`.

A save runs from the file's own directory, so cohere finds the nearest project, as `s c` would run
from there.

## Install

Install **cohere** from the Extensions view: it's published to the Visual Studio Marketplace for VS
Code, and to Open VSX for Cursor, VSCodium and Windsurf. Both update it for you. The extension is
versioned with cohere itself, so cohere 1.0.0 pairs with extension 1.0.0.

A `.vsix` from a release installs the same way:

```sh
code --install-extension cohere-1.0.0.vsix
```

Then set it as the formatter (`s doctor` writes these):

```json
"editor.defaultFormatter": "system-inc.cohere",
"editor.formatOnSave": true
```

The extension runs cohere and does not bundle it. Install cohere in the project with
`pnpm add -D @system-inc/cohere`. When no cohere is found, the first save in a project says so, names
where it looked, and offers to copy that command. The file saves as typed.

## Test

```sh
node --test cohere-format.test.js                                  # resolution and failure handling
COHERE_BINARY=/path/to/cohere node --test cohere-format.test.js     # plus a real save against the gate
```

The extension itself is tested in a real VS Code, from its `.vsix`, in a profile with nothing else in
it (macOS, VS Code in `/Applications`; it opens windows while it runs):

```sh
bash ../../.github/scripts/package-vscode-extension.sh 1.0.0 /tmp/vsix
node test-vscode/run.js /tmp/vsix/cohere-1.0.0.vsix ~/Projects/ahra "$(command -v cohere)"
```

It saves a misformatted file in the project and requires the bytes on disk to equal what
`cohere --fix --format` writes for the same text. Then it saves a file where no cohere can be found
and requires it saved as typed.
