'use strict';

// Runs inside a real VS Code's extension host, started by run.js with the cohere .vsix installed into an
// empty profile. It saves files the way a person does and reads the result from disk.
//
// COHERE_GATE_CASE picks the case:
//   present  saves a misformatted probe in COHERE_GATE_FOLDER and requires the bytes on disk to equal
//            what `cohere --fix --format` writes for the same text, and to differ from what was typed.
//   schema   in a project with @system-inc/cohere installed, opens a CohereSettings.json whose $schema
//            names the installed schema, and requires VS Code to flag a misspelled key, to load the
//            schema at all, and to flag nothing in a correct file.
//   absent   saves a file where no cohere can be found and requires it saved as typed. The message is
//            shown by VS Code, which offers no way to read a notification back, so its wording is held
//            by cohere-format.test.js and its look by a person.

const NodeAssert = require('node:assert/strict');
const NodeChildProcess = require('node:child_process');
const NodeFileSystem = require('node:fs');
const NodePath = require('node:path');
const Vscode = require('vscode');

// Misformatted on purpose, with a lint fix in it (`let` that is never reassigned), so a save that only
// formatted, or only fixed, would not match the gate.
const typed = 'export function cohereVscodeGateProbe(): number {\n  let   first = 1\n    return first }\n';

async function saveAsTyped(filePath) {
    NodeFileSystem.writeFileSync(filePath, '');
    const document = await Vscode.workspace.openTextDocument(filePath);
    await Vscode.window.showTextDocument(document);
    const edit = new Vscode.WorkspaceEdit();
    edit.insert(document.uri, new Vscode.Position(0, 0), typed);
    NodeAssert.ok(await Vscode.workspace.applyEdit(edit), 'the edit was not applied');
    const formatter = Vscode.workspace.getConfiguration('editor', document).get('defaultFormatter');
    NodeAssert.equal(formatter, 'system-inc.cohere', `${document.languageId} files here are formatted by ${formatter}`);
    // Cmd+S: the save a person makes, with the active editor's save participants.
    await Vscode.commands.executeCommand('workbench.action.files.save');
    NodeAssert.ok(!document.isDirty, 'the save did not happen');
    return NodeFileSystem.readFileSync(filePath, 'utf8');
}

async function run() {
    const extension = Vscode.extensions.getExtension('system-inc.cohere');
    NodeAssert.ok(extension !== undefined, 'system-inc.cohere is not installed');
    NodeAssert.ok(
        extension.extensionPath.startsWith(process.env.COHERE_GATE_EXTENSIONS),
        `system-inc.cohere was loaded from ${extension.extensionPath}, not from the profile the .vsix was installed into`,
    );
    await extension.activate();

    const folder = process.env.COHERE_GATE_FOLDER;
    const probe = NodePath.join(folder, process.env.COHERE_GATE_PROBE);

    if(process.env.COHERE_GATE_CASE === 'schema') {
        await checkSettingsSchema(folder);
        return;
    }

    if(process.env.COHERE_GATE_CASE === 'present') {
        try {
            const saved = await saveAsTyped(probe);
            NodeAssert.notEqual(saved, typed, 'the save left the file as typed, so nothing formatted it');

            // The gate, on the same text at the same path.
            NodeFileSystem.writeFileSync(probe, typed);
            const environment = { ...process.env };
            delete environment.COHERE_BINARY;
            const gate = NodeChildProcess.spawnSync(process.env.COHERE_GATE_BINARY, ['--fix', '--format', probe], {
                cwd: folder,
                encoding: 'utf8',
                env: environment,
                timeout: 600000,
            });
            NodeAssert.equal(gate.status, 0, gate.stdout + gate.stderr);
            const written = NodeFileSystem.readFileSync(probe, 'utf8');
            NodeAssert.equal(saved, written, 'the save and the gate wrote different bytes');
            console.log(`cohere gate: present: the save wrote ${Buffer.byteLength(saved)} bytes, identical to cohere --fix --format`);
        }
        finally {
            NodeFileSystem.rmSync(probe, { force: true });
        }
        return;
    }

    const saved = await saveAsTyped(probe);
    NodeAssert.equal(saved, typed, 'with no cohere, the file did not save as typed');
    console.log('cohere gate: absent: the file saved as typed');
}

// The diagnostics VS Code's JSON support reports for a settings file, once the schema has had time to load.
async function settingsDiagnostics(filePath, contents) {
    NodeFileSystem.writeFileSync(filePath, contents);
    const document = await Vscode.workspace.openTextDocument(filePath);
    await Vscode.window.showTextDocument(document);
    for(let waited = 0; waited < 15000; waited += 250) {
        await new Promise(function(resolve) {
            setTimeout(resolve, 250);
        });
        const diagnostics = Vscode.languages.getDiagnostics(document.uri);
        if(diagnostics.length > 0) {
            return diagnostics;
        }
    }
    return [];
}

async function checkSettingsSchema(folder) {
    const schemaPath = './node_modules/@system-inc/cohere/schema/CohereSettings.schema.json';
    const settingsPath = NodePath.join(folder, 'CohereSettings.json');
    try {
        const flagged = await settingsDiagnostics(settingsPath, JSON.stringify({ $schema: schemaPath, ignorPatterns: ['dist/**'] }, null, 4) + '\n');
        const messages = flagged.map(function(diagnostic) {
            return diagnostic.message;
        });
        NodeAssert.ok(!messages.some(function(message) {
            return /load|resolve/i.test(message);
        }), `VS Code could not load the installed schema: ${messages.join(' | ')}`);
        NodeAssert.ok(messages.some(function(message) {
            return message.includes('ignorPatterns');
        }), `the misspelled key was not flagged: ${messages.join(' | ') || 'no diagnostics'}`);
        await Vscode.commands.executeCommand('workbench.action.closeActiveEditor');

        // The control: the same file spelled right has nothing to flag, so the flag above came from the
        // schema's rules rather than from anything else about the file.
        const clean = await settingsDiagnostics(settingsPath, JSON.stringify({ $schema: schemaPath, ignorePatterns: ['dist/**'] }, null, 4) + '\n');
        NodeAssert.deepEqual(clean.map(function(diagnostic) {
            return diagnostic.message;
        }), [], 'a correct settings file was flagged');
        console.log(`cohere gate: schema: VS Code flagged the misspelled key (${messages.join(' | ')}) and nothing in the correct file`);
    }
    finally {
        NodeFileSystem.rmSync(settingsPath, { force: true });
    }
}

module.exports = { run };
