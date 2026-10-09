'use strict';

// The save, without VS Code: find cohere the way the Structure CLI does, hand it the buffer, and read
// back what `cohere --fix --format` writes for the file. Kept apart from extension.js so it can be run
// and tested with plain node (cohere-format.test.js).

const NodeChildProcess = require('node:child_process');
const NodeFileSystem = require('node:fs');
const NodeModule = require('node:module');
const NodePath = require('node:path');

// The host the extension runs on, as Node spells it, which is also how the platform packages are named.
const currentHost = { platform: process.platform, arch: process.arch };

function cohereBinaryFileName(host) {
    return host.platform === 'win32' ? 'cohere.exe' : 'cohere';
}

// The dispatcher's own override. Read here, and removed from the environment cohere runs in.
const cohereBinaryOverrideVariable = 'COHERE_BINARY';

// What a project runs to install cohere, said wherever cohere was not found.
const installCommand = 'pnpm add -D @system-inc/cohere';

// The arguments a save runs with: the gate's own, with no engine named, so a save formats with whatever
// cohere's default is.
const saveArguments = ['--fix', '--format'];

/*
 * Finds the cohere binary for a project, or returns undefined having found nothing.
 *
 * The order is Structure's (libraries/structure/command-line/Structure.ts, resolveCohereBinary), so
 * the editor and `s c` run the same binary:
 *
 *   1. COHERE_BINARY, an absolute path, which is the cohere dispatcher's own override. A broken one is
 *      a failure rather than a fallback, because whoever set it named the binary they want.
 *   2. The project's own install, at the project root or above it. On macOS and Linux that is the
 *      launcher, node_modules/.bin/cohere. On Windows it is the platform package's cohere.exe; see
 *      projectPlatformBinary.
 *   3. A cohere on PATH.
 */
function resolveCohereBinary(projectDirectory, environment = process.env, host = currentHost) {
    const overridePath = environment[cohereBinaryOverrideVariable];
    if(overridePath !== undefined && overridePath !== '') {
        return NodeFileSystem.existsSync(overridePath) ? overridePath : undefined;
    }
    if(host.platform === 'win32') {
        const installed = projectPlatformBinary(projectDirectory, host);
        if(installed !== undefined) {
            return installed;
        }
    }
    return cohereCandidatePaths(projectDirectory, environment, host).find(function(candidate) {
        return NodeFileSystem.existsSync(candidate);
    });
}

/*
 * The project's cohere on Windows: the platform package's own cohere.exe, found the way the package's
 * Node launcher finds it.
 *
 * node_modules/.bin holds no .exe on Windows. npm and pnpm write `cohere`, `cohere.cmd` and `cohere.ps1`
 * shims there for the launcher script, and Node refuses to spawn a .cmd without a shell since
 * CVE-2024-27980, which would put the file path through cmd's quoting. So this finds
 * @system-inc/cohere at the project or above, and asks require.resolve, from that package's real
 * directory, for @system-inc/cohere-win32-<arch>, as the launcher does from inside it. The real
 * directory matters under pnpm, where the platform package sits beside cohere's real directory, not
 * beside the link in the project's node_modules.
 */
function projectPlatformBinary(projectDirectory, host = currentHost) {
    for(let directory = projectDirectory; ; directory = NodePath.dirname(directory)) {
        const manifest = NodePath.join(directory, 'node_modules', '@system-inc', 'cohere', 'package.json');
        if(NodeFileSystem.existsSync(manifest)) {
            try {
                const platformManifest = NodeModule.createRequire(NodeFileSystem.realpathSync(manifest)).resolve(
                    `@system-inc/cohere-${host.platform}-${host.arch}/package.json`,
                );
                const binary = NodePath.join(NodePath.dirname(platformManifest), 'bin', cohereBinaryFileName(host));
                return NodeFileSystem.existsSync(binary) ? binary : undefined;
            }
            catch {
                return undefined;
            }
        }
        if(NodePath.dirname(directory) === directory) {
            return undefined;
        }
    }
}

/*
 * The fixed places cohere is looked for after the override, in order: the launcher in node_modules/.bin
 * at the project and each directory above it, then each PATH directory. On Windows the project's install
 * is found by projectPlatformBinary instead, and PATH is searched for cohere.exe, which a global npm
 * install never puts there: it is for a binary someone placed on PATH themselves.
 */
function cohereCandidatePaths(projectDirectory, environment = process.env, host = currentHost) {
    const candidates = [];
    if(host.platform !== 'win32') {
        for(let directory = projectDirectory; ; directory = NodePath.dirname(directory)) {
            candidates.push(NodePath.join(directory, 'node_modules', '.bin', cohereBinaryFileName(host)));
            if(NodePath.dirname(directory) === directory) {
                break;
            }
        }
    }
    for(const pathDirectory of (environment.PATH ?? '').split(NodePath.delimiter)) {
        if(pathDirectory !== '') {
            candidates.push(NodePath.resolve(pathDirectory, cohereBinaryFileName(host)));
        }
    }
    return candidates;
}

/*
 * What a save says when no cohere was found: that the file was not formatted, where cohere was looked
 * for, and the command that installs it. Structure's `s c` says the same when it finds none
 * (reportUnresolvableCohereBinary), so the two read alike.
 */
function cohereNotFoundMessage(projectDirectory, environment = process.env) {
    const overridePath = environment[cohereBinaryOverrideVariable];
    if(overridePath !== undefined && overridePath !== '') {
        return `cohere was not found, so this file was not formatted: ${cohereBinaryOverrideVariable} names ${overridePath}, which does not exist.`;
    }
    return `cohere was not found, so this file was not formatted. Looked in ${projectDirectory}'s node_modules/.bin and above, then PATH. Install it in the project with: ${installCommand}`;
}

// The environment cohere runs in: this one without the override. The dispatcher honors the same
// variable, so handing it down to the binary it names makes a dispatcher resolve to itself and exec in
// a loop. Resolution has already happened by then.
function environmentWithoutCohereOverride(environment = process.env) {
    const childEnvironment = { ...environment };
    delete childEnvironment[cohereBinaryOverrideVariable];
    return childEnvironment;
}

/*
 * Runs one save: `cohere --fix --format --stdin-filepath <file>` with the buffer on stdin.
 *
 * Resolves to { text } on success, which is the buffer unchanged when cohere declines the file (no
 * printer for its type, outside the format scope, source that does not parse), or to { error } when
 * cohere failed: a broken formatter, a project it cannot load, a binary that would not start. The
 * caller shows the error and saves the buffer as typed.
 *
 * It runs from the file's own directory, so cohere finds the nearest project the way `s c` run there
 * would, nested repositories included.
 */
function formatWithCohere(binary, filePath, text, cancellation) {
    return new Promise(function(resolve) {
        const child = NodeChildProcess.spawn(binary, [...saveArguments, '--stdin-filepath', filePath], {
            cwd: NodePath.dirname(filePath),
            env: environmentWithoutCohereOverride(),
        });
        const stdout = [];
        const stderr = [];
        child.stdout.on('data', function(chunk) {
            stdout.push(chunk);
        });
        child.stderr.on('data', function(chunk) {
            stderr.push(chunk);
        });
        child.on('error', function(error) {
            resolve({ error: `could not run ${binary}: ${error.message}` });
        });
        child.on('close', function(code, signal) {
            if(signal !== null) {
                resolve({ error: `cohere stopped on ${signal}` });
                return;
            }
            if(code !== 0) {
                resolve({ error: Buffer.concat(stderr).toString('utf8').trim() || `cohere exited ${code}` });
                return;
            }
            resolve({ text: Buffer.concat(stdout).toString('utf8') });
        });
        if(cancellation !== undefined) {
            cancellation(function() {
                child.kill();
            });
        }
        child.stdin.on('error', function() {
            // A child that exits before reading all of stdin closes the pipe; its exit says why.
        });
        child.stdin.end(text, 'utf8');
    });
}

module.exports = {
    cohereNotFoundMessage,
    environmentWithoutCohereOverride,
    formatWithCohere,
    installCommand,
    resolveCohereBinary,
    saveArguments,
};
