'use strict';

// The extension's gate in a real VS Code, from its .vsix, in a profile with nothing else in it.
//
//     node test-vscode/run.js <path.vsix> <project folder> <cohere binary> [<project with cohere installed>]
//     node test-vscode/run.js dist/cohere-1.0.0.vsix ~/Projects/ahra ~/.local/bin/cohere /tmp/installed
//
// Two cases, each in a fresh profile: the user data and extensions directories are new and empty, the
// .vsix is installed with VS Code's own `--install-extension`, and the only settings are the two the
// README tells a person to set, plus the TypeScript override explained in freshProfile. VS Code is then started in its extension test mode, which runs
// suite.js inside the extension host and exits with its result.
//
//   present  in <project folder>, where cohere is found: a save must write what the gate writes.
//   absent   in an empty directory, with PATH reduced to the system's and no COHERE_BINARY: the save
//            must leave the file as typed.
//   schema   when the fourth argument is given, a project with @system-inc/cohere installed from its npm
//            package: a CohereSettings.json naming the installed schema must have a misspelled key flagged.
//
// macOS only, with VS Code in /Applications. It opens visible windows while it runs.

const NodeChildProcess = require('node:child_process');
const NodeFileSystem = require('node:fs');
const NodeOs = require('node:os');
const NodePath = require('node:path');

const application = '/Applications/Visual Studio Code.app';
const cli = NodePath.join(application, 'Contents', 'Resources', 'app', 'bin', 'code');
const electron = NodePath.join(application, 'Contents', 'MacOS', 'Code');

const [vsix, projectFolder, cohereBinary, installedProject] = process.argv.slice(2).map(function(argument) {
    return NodePath.resolve(argument);
});
if(cohereBinary === undefined) {
    console.error('usage: node test-vscode/run.js <path.vsix> <project folder> <cohere binary>');
    process.exit(2);
}

function freshProfile() {
    const root = NodeFileSystem.realpathSync(NodeFileSystem.mkdtempSync(NodePath.join(NodeOs.tmpdir(), 'cohere-vscode-gate-')));
    const userData = NodePath.join(root, 'user-data');
    const extensions = NodePath.join(root, 'extensions');
    NodeFileSystem.mkdirSync(NodePath.join(userData, 'User'), { recursive: true });
    NodeFileSystem.mkdirSync(extensions);
    NodeFileSystem.writeFileSync(NodePath.join(userData, 'User', 'settings.json'), JSON.stringify({
        'editor.defaultFormatter': 'system-inc.cohere',
        'editor.formatOnSave': true,
        // A project's .vscode/settings.json outranks the user's, and ahra's still names Prettier until
        // `s doctor` rewrites it. A language's own setting outranks any setting for every language, at
        // any scope, so this stands in for the project's for the probe's language.
        '[typescript]': { 'editor.defaultFormatter': 'system-inc.cohere' },
    }, null, 4));
    const install = NodeChildProcess.spawnSync(cli, ['--user-data-dir', userData, '--extensions-dir', extensions, '--install-extension', vsix], {
        encoding: 'utf8',
    });
    if(install.status !== 0) {
        throw new Error(`installing ${vsix} failed: ${install.stdout}${install.stderr}`);
    }
    return { root, userData, extensions };
}

function runCase(name, folder, environment) {
    const profile = freshProfile();
    const result = NodeChildProcess.spawnSync(electron, [
        folder,
        '--user-data-dir', profile.userData,
        '--extensions-dir', profile.extensions,
        `--extensionDevelopmentPath=${NodePath.join(__dirname, 'host')}`,
        `--extensionTestsPath=${NodePath.join(__dirname, 'suite.js')}`,
        '--disable-workspace-trust',
        '--skip-welcome',
        '--skip-release-notes',
        '--new-window',
    ], {
        encoding: 'utf8',
        env: { ...environment, COHERE_GATE_CASE: name, COHERE_GATE_FOLDER: folder, COHERE_GATE_EXTENSIONS: profile.extensions },
        timeout: 900000,
    });
    const output = `${result.stdout}${result.stderr}`;
    const said = output.split('\n').filter(function(line) {
        return line.includes('cohere gate:') || line.includes('AssertionError') || line.includes('Error:');
    });
    console.log(said.join('\n'));
    if(result.status !== 0) {
        console.error(`${name}: VS Code exited ${result.status}${result.signal !== null ? ` on ${result.signal}` : ''}`);
        console.error(output.slice(-4000));
        return false;
    }
    return true;
}

const withCohere = { ...process.env, COHERE_GATE_BINARY: cohereBinary, COHERE_GATE_PROBE: 'CohereVscodeGateProbe.ts' };
delete withCohere.COHERE_BINARY;
delete withCohere.ELECTRON_RUN_AS_NODE;

const empty = NodeFileSystem.realpathSync(NodeFileSystem.mkdtempSync(NodePath.join(NodeOs.tmpdir(), 'cohere-vscode-absent-')));
const withoutCohere = {
    HOME: process.env.HOME,
    PATH: '/usr/bin:/bin:/usr/sbin:/sbin',
    TMPDIR: process.env.TMPDIR,
    COHERE_GATE_PROBE: 'Probe.ts',
};

const present = runCase('present', projectFolder, withCohere);
const absent = runCase('absent', empty, withoutCohere);
const schema = installedProject === undefined ? true : runCase('schema', installedProject, withoutCohere);
process.exit(present && absent && schema ? 0 : 1);
