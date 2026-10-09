'use strict';

// The VS Code side of the save: a document formatter that runs cohere, so `editor.formatOnSave` with
// this extension as `editor.defaultFormatter` formats the buffer before the save lands. The edit is
// applied to the document in memory, so nothing reloads and undo keeps working.

const NodePath = require('node:path');
const Vscode = require('vscode');
const { cohereNotFoundMessage, formatWithCohere, installCommand, resolveCohereBinary } = require('./cohere-format.js');

const copyInstallCommand = 'Copy install command';

function activate(context) {
    // Projects already told cohere is missing in this window. The message is the same on every save,
    // so it is shown once rather than on each; every save still leaves the file unformatted.
    const toldNotFound = new Set();

    const provider = {
        async provideDocumentFormattingEdits(document, _options, token) {
            if(document.uri.scheme !== 'file' || document.isUntitled) {
                return [];
            }
            const filePath = document.uri.fsPath;
            const workspaceFolder = Vscode.workspace.getWorkspaceFolder(document.uri);
            const projectDirectory = workspaceFolder !== undefined ? workspaceFolder.uri.fsPath : NodePath.dirname(filePath);

            const binary = resolveCohereBinary(projectDirectory);
            if(binary === undefined) {
                if(!toldNotFound.has(projectDirectory)) {
                    toldNotFound.add(projectDirectory);
                    // Not awaited: the save goes ahead as typed while the message is up.
                    Vscode.window.showErrorMessage(cohereNotFoundMessage(projectDirectory), copyInstallCommand).then(function(choice) {
                        if(choice === copyInstallCommand) {
                            Vscode.env.clipboard.writeText(installCommand);
                        }
                    });
                }
                return [];
            }

            const text = document.getText();
            const result = await formatWithCohere(binary, filePath, text, function(stop) {
                token.onCancellationRequested(stop);
            });
            if(token.isCancellationRequested) {
                return [];
            }
            if(result.error !== undefined) {
                // A broken formatter is shown; the buffer saves as typed. A file cohere declines never
                // reaches here: it comes back unchanged with no error.
                Vscode.window.showErrorMessage(`cohere could not format ${NodePath.basename(filePath)}: ${result.error}`);
                return [];
            }
            if(result.text === text) {
                return [];
            }
            const wholeDocument = new Vscode.Range(document.positionAt(0), document.positionAt(text.length));
            return [Vscode.TextEdit.replace(wholeDocument, result.text)];
        },
    };

    context.subscriptions.push(
        Vscode.languages.registerDocumentFormattingEditProvider({ scheme: 'file' }, provider),
    );
}

function deactivate() {}

module.exports = { activate, deactivate };
