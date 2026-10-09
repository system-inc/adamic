// Proves the provenance `npm publish --provenance` will attach, without publishing anything.
//
//     node check-npm-provenance.js <staged packages directory>
//     node check-npm-provenance.js dist
//
// A dry run cannot show it: npm makes provenance inside libnpmpublish's publish, which
// `npm publish --dry-run` never calls (lib/commands/publish.js, read in npm 11). So this calls what npm
// calls there, from the npm on this runner. generateProvenance signs a statement for the packed dispatcher
// through Sigstore with this run's identity, the same signing a publish does, and verifyProvenance is the
// check npm itself runs on a provenance bundle. Nothing is uploaded except the signing's entry in
// Sigstore's public transparency log, which every signing makes.
//
// Then it holds what the registry will hold: the statement names this repository, this workflow and this
// commit; every staged package's repository field names this repository, which the registry requires
// before it accepts provenance; and the bundle refuses a tarball with one byte flipped, and a statement
// re-pointed at that tarball.
'use strict';

const childProcess = require('node:child_process');
const crypto = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

function fail(message) {
    process.stderr.write('check-npm-provenance: ' + message + '\n');
    process.exit(1);
}

// npmModule loads a module from inside the npm installed on this machine, by file, since npm publishes
// none of these as an API. A layout that moved fails here by name rather than as a missing check.
function npmModule(relativePath) {
    const npmRoot = path.join(childProcess.execFileSync('npm', ['root', '-g'], { encoding: 'utf8' }).trim(), 'npm');
    const file = path.join(npmRoot, 'node_modules', relativePath);
    if (!fs.existsSync(file)) {
        fail(file + ' is not in this npm, so its provenance code has moved; find where generateProvenance lives now');
    }
    return require(file);
}

function sha512Hex(file) {
    return crypto.createHash('sha512').update(fs.readFileSync(file)).digest('hex');
}

// statementOf reads the in-toto statement a bundle signs.
function statementOf(bundle) {
    return JSON.parse(Buffer.from(bundle.dsseEnvelope.payload, 'base64').toString('utf8'));
}

// requireRefusals holds a bundle that verifies to refusing the two tampers it exists to catch: the package
// changed after signing, and the statement edited to name the changed package. The second is refused by the
// signature, not by the digest comparison, which is the half that shows the signature is checked at all.
async function requireRefusals(verifyProvenance, subject, bundleFile, tarball) {
    await verifyProvenance(subject, bundleFile);

    const workspace = fs.mkdtempSync(path.join(os.tmpdir(), 'provenance-tamper-'));
    const tampered = path.join(workspace, path.basename(tarball));
    fs.copyFileSync(tarball, tampered);
    const contents = fs.readFileSync(tampered);
    contents[contents.length >> 1] ^= 0x01;
    fs.writeFileSync(tampered, contents);
    const tamperedSubject = { name: subject.name, digest: { sha512: sha512Hex(tampered) } };
    if (tamperedSubject.digest.sha512 === subject.digest.sha512) {
        fail('flipping a byte did not change the tarball digest, so the tamper below would test nothing');
    }

    let refused = false;
    try {
        await verifyProvenance(tamperedSubject, bundleFile);
    } catch (error) {
        refused = true;
        console.log('ok   a tarball with one byte flipped is refused: ' + error.message);
    }
    if (!refused) {
        fail('the provenance verified for a tarball with a flipped byte');
    }

    const bundle = JSON.parse(fs.readFileSync(bundleFile, 'utf8'));
    const statement = statementOf(bundle);
    statement.subject[0].digest.sha512 = tamperedSubject.digest.sha512;
    bundle.dsseEnvelope.payload = Buffer.from(JSON.stringify(statement)).toString('base64');
    const repointed = path.join(workspace, 'repointed.sigstore.json');
    fs.writeFileSync(repointed, JSON.stringify(bundle));

    refused = false;
    try {
        await verifyProvenance(tamperedSubject, repointed);
    } catch (error) {
        refused = true;
        console.log('ok   a statement re-pointed at the tampered tarball is refused: ' + error.message);
    }
    if (!refused) {
        fail('a statement edited to name the tampered tarball verified, so the signature is not being checked');
    }
}

async function main() {
    const staged = path.resolve(process.argv[2] || '');
    const dispatcher = path.join(staged, 'cohere');
    if (!fs.existsSync(path.join(dispatcher, 'package.json'))) {
        fail('no staged dispatcher at ' + dispatcher);
    }
    for (const variable of ['GITHUB_ACTIONS', 'GITHUB_SERVER_URL', 'GITHUB_REPOSITORY', 'GITHUB_SHA', 'ACTIONS_ID_TOKEN_REQUEST_URL']) {
        if (!process.env[variable]) {
            fail(variable + ' is not set: provenance is signed with a GitHub Actions identity, in a job with id-token: write');
        }
    }
    const repository = process.env.GITHUB_SERVER_URL + '/' + process.env.GITHUB_REPOSITORY;

    // The registry refuses provenance from a repository other than the one a package names, so a
    // repository field that drifted would pass every check here and fail at the real publish.
    let packages = 0;
    for (const entry of fs.readdirSync(staged, { withFileTypes: true })) {
        const manifestFile = path.join(staged, entry.name, 'package.json');
        if (!entry.isDirectory() || !fs.existsSync(manifestFile)) {
            continue;
        }
        const manifest = JSON.parse(fs.readFileSync(manifestFile, 'utf8'));
        if (manifest.repository?.url !== 'git+' + repository + '.git') {
            fail(manifest.name + ' names its repository as ' + manifest.repository?.url + ', and this run is ' + repository + ', so the registry would refuse its provenance');
        }
        packages++;
    }
    console.log('ok   all ' + packages + ' staged packages name ' + repository);

    const workspace = fs.mkdtempSync(path.join(os.tmpdir(), 'provenance-'));
    const packed = childProcess.execFileSync('npm', ['pack', '--silent', '--pack-destination', workspace, dispatcher], { encoding: 'utf8' }).trim().split('\n').pop();
    const tarball = path.join(workspace, packed);

    const manifest = JSON.parse(fs.readFileSync(path.join(dispatcher, 'package.json'), 'utf8'));
    const npa = npmModule('npm-package-arg');
    const { generateProvenance, verifyProvenance } = npmModule('libnpmpublish/lib/provenance.js');

    // The subject exactly as libnpmpublish builds it: the package URL, and the tarball's sha512 in hex.
    const subject = {
        name: npa.toPurl(npa(manifest.name + '@' + manifest.version)),
        digest: { sha512: sha512Hex(tarball) },
    };
    const bundle = await generateProvenance([subject], {});
    const bundleFile = path.join(workspace, 'provenance.sigstore.json');
    fs.writeFileSync(bundleFile, JSON.stringify(bundle));
    console.log('ok   signed provenance for ' + subject.name);

    const definition = statementOf(bundle).predicate.buildDefinition;
    const commit = definition.resolvedDependencies[0].digest.gitCommit;
    if (commit !== process.env.GITHUB_SHA) {
        fail('the provenance names commit ' + commit + ', and this run built ' + process.env.GITHUB_SHA);
    }
    if (definition.externalParameters.workflow.repository !== repository) {
        fail('the provenance names ' + definition.externalParameters.workflow.repository + ', not ' + repository);
    }
    console.log('ok   the provenance names ' + repository + ' at ' + commit + ', built by ' + definition.externalParameters.workflow.path);

    await requireRefusals(verifyProvenance, subject, bundleFile, tarball);
}

module.exports = { requireRefusals, npmModule, sha512Hex };

if (require.main === module) {
    main().catch((error) => fail(error.stack || String(error)));
}
