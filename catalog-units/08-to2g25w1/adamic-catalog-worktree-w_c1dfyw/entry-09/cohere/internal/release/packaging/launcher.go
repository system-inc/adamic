package release

import (
	"strings"
)

// DispatcherLauncher returns the Node script published as the `cohere` command.
//
// The launcher is Node rather than Go, which is a deliberate reversal of the obvious design. A Go
// dispatcher would have to be cross-compiled per platform, which makes it a seventh platform
// package rather than one package everyone installs — and the whole point of the dispatcher package
// is that a consumer runs one `pnpm add -D @system-inc/cohere` and gets the right binary. Node is
// already present by construction (this is an npm install), and `require.resolve` asks the package
// manager where a package actually is, which is more reliable than any filesystem walk: it
// understands pnpm's isolated store, npm's hoisting, and yarn's layouts without having to model any
// of them.
//
// The script embeds here rather than living as a checked-in `.js` file so that the platform names,
// the package prefix, and the binary names have exactly one definition. A launcher whose idea of the
// package name drifted from the manifest's would resolve nothing on every machine, and that failure
// looks identical to an unsupported platform.
func DispatcherLauncher() string {
	replacements := []string{
		"__PLATFORM_PACKAGE_PREFIX__", PlatformPackagePrefix,
		"__PACKAGE_SCOPE__", PackageScope,
		"__OVERRIDE_VARIABLE__", BinaryOverrideVariable,
		"__CHECKSUMS_FILE_NAME__", ChecksumsFileName,
	}
	return strings.NewReplacer(replacements...).Replace(launcherSource)
}

// launcherSource is the dispatcher script, with placeholders filled in by DispatcherLauncher.
//
// It is CommonJS rather than an ES module because a `bin` script has no package.json of its own to
// declare a type, and a `.js` file inside a package without `"type": "module"` is CommonJS. Writing
// it as ESM would work until someone added a type field, which is a trap rather than a design.
const launcherSource = `#!/usr/bin/env node
'use strict';

// The cohere launcher. It finds the binary for this platform and hands the process over to it.
//
// The one rule this file exists to hold: a missing binary exits non-zero naming the platform. It
// never falls through to a bare command name, never resolves a binary built for another platform,
// and never exits zero having run nothing. The gate cohere replaces printed a green checkmark over
// zero files for days because a resolver did exactly that, and an empty result is indistinguishable
// from a clean tree.
//
// The second rule: a binary that differs from the checksum this package was released with is never
// run. A CI that installs cohere executes it on every commit, which is a trust extended to us, so
// what it runs has to be what the release built.

const childProcess = require('node:child_process');
const crypto = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

const platformPackagePrefix = '__PLATFORM_PACKAGE_PREFIX__';
const packageScope = '__PACKAGE_SCOPE__';
const overrideVariable = '__OVERRIDE_VARIABLE__';
const checksumsFileName = '__CHECKSUMS_FILE_NAME__';

// The prefix is the dispatcher's own name and a hyphen.
const dispatcherPackage = platformPackagePrefix.slice(0, -1);

// Node's process.arch and process.platform are already npm's spelling, which is the same spelling
// the platform packages publish under. No translation belongs here.
const platformPackage = platformPackagePrefix + process.platform + '-' + process.arch;
const binaryFileName = process.platform === 'win32' ? 'cohere.exe' : 'cohere';

function fail(message) {
    process.stderr.write('cohere: ' + message + '\n');
    process.exit(1);
}

// resolveOverride validates an explicitly requested binary.
//
// Every failure is fatal rather than a fallback. Someone who sets this has stated which binary they
// want to run, so quietly running a different one would hand them results they would read as their
// own build's. A stale export in a shell profile is exactly how that happens.
function resolveOverride(requested) {
    const resolved = path.resolve(requested);
    let information;
    try {
        information = fs.statSync(resolved);
    } catch {
        fail(overrideVariable + ' points at ' + resolved + ', and there is nothing there.');
    }
    if (information.isDirectory()) {
        fail(overrideVariable + ' points at ' + resolved + ', which is a directory rather than a binary.');
    }
    return resolved;
}

// resolvePlatformPackage asks the package manager where the platform package is.
//
// require.resolve is used rather than a walk up node_modules because it is the package manager's
// own answer: it understands pnpm's isolated store, npm's hoisting, and yarn's layouts without this
// file having to model any of them. The package.json is resolved rather than the package itself
// because the platform packages have no entry point — they are a binary and a manifest.
function resolvePlatformPackage() {
    let manifestPath;
    try {
        manifestPath = require.resolve(platformPackage + '/package.json');
    } catch {
        return null;
    }
    return path.dirname(manifestPath);
}

// verifyPlatformPackage checks each file the platform package ships against the checksums this
// dispatcher was published with, and runs nothing when one differs.
//
// The dispatcher pins its platform packages to its own exact version, so its SHA256SUMS describe the
// only binaries it is ever installed beside. A binary that differs changed after the release built
// it: a corrupted download or cache, a mirror serving something else, or a compromised registry.
// Each is refused, naming the file, and none falls back to running the binary anyway.
//
// Hashing the binaries costs tens of milliseconds, which every run would pay, so a passing verdict is
// remembered in the user cache. It holds each file's device, inode, size, and modification and change
// times, beside the hash it was checked against, and is honored only when all of them still match. Any
// write to a file moves its change time, which no unprivileged process can set back, and a replaced file
// has a new inode, so a remembered verdict only ever describes the bytes it was reached on. The files are
// stat'ed before they are hashed, so a write landing during the hash leaves a verdict that no longer
// matches. This guards what reaches the machine, not against someone who can already write to
// node_modules: they could edit this file instead.
function verifyPlatformPackage(packageDirectory) {
    const checksumsPath = path.join(__dirname, '..', checksumsFileName);
    let checksums;
    try {
        checksums = fs.readFileSync(checksumsPath, 'utf8');
    } catch {
        fail(
            checksumsPath + ' is missing, so ' + platformPackage + ' cannot be checked against the release.\n' +
            '  The ' + dispatcherPackage + ' package is incomplete — reinstall it.',
        );
    }

    // Lines name files from the directory holding the packages, as "cohere-darwin-arm64/bin/cohere":
    // the package name without its scope, which is how node_modules lays a scope out.
    const directoryName = platformPackage.slice(packageScope.length + 1);
    const expected = [];
    for (const line of checksums.split('\n')) {
        const match = /^([0-9a-f]{64}) {2}(.+)$/.exec(line);
        if (match !== null && match[2].startsWith(directoryName + '/')) {
            expected.push({ hash: match[1], file: path.join(packageDirectory, match[2].slice(directoryName.length + 1)) });
        }
    }
    if (expected.length === 0) {
        fail(
            checksumsPath + ' lists no binary for ' + platformPackage + ', so it cannot be checked against the release.\n' +
            '  Reinstall ' + dispatcherPackage + '.',
        );
    }

    const identities = expected.map((entry) => {
        let information;
        try {
            information = fs.statSync(entry.file, { bigint: true });
        } catch {
            fail(
                platformPackage + ' is installed but ' + entry.file + ', which the release shipped, is missing.\n' +
                '  The package is incomplete — reinstall it.',
            );
        }
        return [entry.hash, information.dev, information.ino, information.size, information.mtimeNs, information.ctimeNs].join(' ');
    });
    const verdict = identities.join('\n');
    const verdictPath = rememberedVerdictPath(packageDirectory);
    if (verdictPath !== null && readQuietly(verdictPath) === verdict) {
        return;
    }

    for (const entry of expected) {
        let actual;
        try {
            actual = crypto.createHash('sha256').update(fs.readFileSync(entry.file)).digest('hex');
        } catch (error) {
            fail('could not read ' + entry.file + ' to check it against the release: ' + error.message);
        }
        if (actual !== entry.hash) {
            fail(
                entry.file + ' does not match the checksum ' + dispatcherPackage + ' was released with, so it was not run.\n' +
                '  expected ' + entry.hash + '\n' +
                '  found    ' + actual + '\n' +
                '  It changed after the release built it: a corrupted download or cache, or tampering.\n' +
                '  Reinstall ' + platformPackage + '. To run a build of your own on purpose, set ' + overrideVariable + '.',
            );
        }
    }

    if (verdictPath !== null) {
        rememberQuietly(verdictPath, verdict);
    }
}

// rememberedVerdictPath is where a passing verdict for one installed package is kept, in the same user
// cache directory Go's os.UserCacheDir names, or null when this machine has none.
function rememberedVerdictPath(packageDirectory) {
    let cacheDirectory;
    try {
        if (process.platform === 'win32') {
            cacheDirectory = process.env.LOCALAPPDATA;
        } else if (process.platform === 'darwin') {
            cacheDirectory = path.join(os.homedir(), 'Library', 'Caches');
        } else {
            const configured = process.env.XDG_CACHE_HOME;
            cacheDirectory = configured && path.isAbsolute(configured) ? configured : path.join(os.homedir(), '.cache');
        }
    } catch {
        return null;
    }
    if (!cacheDirectory) {
        return null;
    }
    const key = crypto.createHash('sha256').update(packageDirectory).digest('hex');
    return path.join(cacheDirectory, 'cohere', 'verified-binaries', key);
}

function readQuietly(file) {
    try {
        return fs.readFileSync(file, 'utf8');
    } catch {
        return null;
    }
}

// rememberQuietly keeps a verdict, written beside its place and renamed in, so two runs at once never
// read half of one. A cache that cannot be written costs the next run a hash and nothing else, so
// failures here are ignored: the binary was already checked.
function rememberQuietly(file, verdict) {
    const temporary = file + '.' + process.pid + '.tmp';
    try {
        fs.mkdirSync(path.dirname(file), { recursive: true });
        fs.writeFileSync(temporary, verdict);
        fs.renameSync(temporary, file);
    } catch {
        try {
            fs.unlinkSync(temporary);
        } catch {
            // Never written.
        }
    }
}

function resolveBinary() {
    // An override is not checked: whoever sets it has named the binary they mean to run, which is
    // usually one they built, and no release checksum describes it.
    const requested = (process.env[overrideVariable] || '').trim();
    if (requested) {
        return resolveOverride(requested);
    }

    const packageDirectory = resolvePlatformPackage();
    if (packageDirectory === null) {
        fail(
            'no binary for ' + process.platform + '/' + process.arch + '.\n' +
            '  ' + platformPackage + ' is not installed.\n' +
            '  If your package manager skipped optional dependencies, reinstall without that flag.\n' +
            '  To point at a local build instead, set ' + overrideVariable + '.',
        );
    }

    const candidate = path.join(packageDirectory, 'bin', binaryFileName);
    if (!fs.existsSync(candidate)) {
        // The package resolved but its binary is absent. This is a packaging failure rather than an
        // unsupported platform, and it is named separately so nobody goes hunting the wrong bug.
        fail(
            platformPackage + ' is installed but its binary is missing at ' + candidate + '.\n' +
            '  The package is incomplete — reinstall it.',
        );
    }

    verifyPlatformPackage(packageDirectory);
    return candidate;
}

const binary = resolveBinary();
const forwarded = process.argv.slice(2);

// Replace this process where the platform allows it, so the binary inherits the terminal directly
// and its exit status is the one the caller sees, with no layer in between able to lose it. Losing
// a non-zero exit is how a gate goes quietly green, which is the failure this whole tool exists to
// make impossible.
if (process.platform !== 'win32' && typeof process.execve === 'function') {
    try {
        process.execve(binary, [binary, ...forwarded], process.env);
    } catch {
        // Older Node, or a platform that will not execve. The spawn below is the real path.
    }
}

const result = childProcess.spawnSync(binary, forwarded, { stdio: 'inherit' });

if (result.error) {
    fail('could not run ' + binary + ': ' + result.error.message);
}

// A process killed by a signal reports a null status. Exiting zero there would report success for a
// run that was terminated, so the conventional 128 + signal encoding is used instead. This is the
// case the obvious implementation gets wrong, because null is falsy and reads as "fine".
if (result.status === null) {
    const signalNumber = os.constants.signals[result.signal] || 0;
    process.exit(signalNumber ? 128 + signalNumber : 1);
}

process.exit(result.status);
`
