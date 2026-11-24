'use strict';
// #bbssdcy, family 63, rulings 323, 324 and 497.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const sites = [
    {
        file: "program.ts",
        owner: "computeCommonSourceDirectoryOfFilenames",
        before: "commonPathComponents.length = i;",
        after: "commonPathComponents.splice(i);",
        snapshot: `export function computeCommonSourceDirectoryOfFilenames(fileNames: readonly string[], currentDirectory: string, getCanonicalFileName: GetCanonicalFileName): string {
    let commonPathComponents: string[] | undefined;
    const failed = forEach(fileNames, sourceFile => {
        // Each file contributes into common source file path
        const sourcePathComponents = getNormalizedPathComponents(sourceFile, currentDirectory);
        sourcePathComponents.pop(); // The base file name is not part of the common directory path

        if (!commonPathComponents) {
            // first file
            commonPathComponents = sourcePathComponents;
            return;
        }

        const n = Math.min(commonPathComponents.length, sourcePathComponents.length);
        for (let i = 0; i < n; i++) {
            if (getCanonicalFileName(commonPathComponents[i]!) !== getCanonicalFileName(sourcePathComponents[i]!)) {
                if (i === 0) {
                    // Failed to find any common path component
                    return true;
                }

                // New common path found that is 0 -> i-1
                commonPathComponents.length = i;
                break;
            }
        }

        // If the sourcePathComponents was shorter than the commonPathComponents, truncate to the sourcePathComponents
        if (sourcePathComponents.length < commonPathComponents.length) {
            commonPathComponents.length = sourcePathComponents.length;
        }
    });

    // A common path can not be found when paths span multiple drives on windows, for example
    if (failed) {
        return "";
    }

    if (!commonPathComponents) { // Can happen when all input files are .d.ts files
        return currentDirectory;
    }

    return getPathFromPathComponents(commonPathComponents);
}`.replaceAll("\n", "\r\n"),
    },
    {
        file: "sys.ts",
        owner: "pollWatchedFileQueue",
        before: "queue.length = definedValueCopyToIndex;",
        after: "queue.splice(definedValueCopyToIndex);",
        snapshot: `function pollWatchedFileQueue<T extends WatchedFileWithIsClosed>(
    host: { getModifiedTime: NonNullable<System["getModifiedTime"]>; },
    queue: (T | undefined)[],
    pollIndex: number,
    chunkSize: number,
    callbackOnWatchFileStat?: (watchedFile: T, pollIndex: number, fileChanged: boolean) => void,
) {
    let definedValueCopyToIndex = pollIndex;
    // Max visit would be all elements of the queue
    for (let canVisit = queue.length; chunkSize && canVisit; nextPollIndex(), canVisit--) {
        const watchedFile = queue[pollIndex];
        if (!watchedFile) {
            continue;
        }
        else if (watchedFile.isClosed) {
            queue[pollIndex] = undefined;
            continue;
        }

        // Only files polled count towards chunkSize
        chunkSize--;
        const fileChanged = onWatchedFileStat(watchedFile, getModifiedTime(host, watchedFile.fileName));
        if (watchedFile.isClosed) {
            // Closed watcher as part of callback
            queue[pollIndex] = undefined;
            continue;
        }

        callbackOnWatchFileStat?.(watchedFile, pollIndex, fileChanged);
        // Defragment the queue while we are at it
        if (queue[pollIndex]) {
            // Copy this file to the non hole location
            if (definedValueCopyToIndex < pollIndex) {
                queue[definedValueCopyToIndex] = watchedFile;
                queue[pollIndex] = undefined;
            }
            definedValueCopyToIndex++;
        }
    }

    // Return next poll index
    return pollIndex;

    function nextPollIndex() {
        pollIndex++;
        if (pollIndex === queue.length) {
            if (definedValueCopyToIndex < pollIndex) {
                // There are holes from definedValueCopyToIndex to end of queue, change queue size
                queue.length = definedValueCopyToIndex;
            }
            pollIndex = 0;
            definedValueCopyToIndex = 0;
        }
    }
}`.replaceAll("\n", "\r\n"),
    },
];

function plan(file, text, site) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `parse error: ${file}`);
    const owners = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === site.owner);
    assert.equal(owners.length, 1, `expected one owner: ${site.owner}`);

    const owner = owners[0];
    const snapshot = owner.getText(source);
    const after = site.snapshot.replace(site.before, site.after);
    assert(snapshot === site.snapshot || snapshot === after, `unreviewed owner: ${site.owner}`);

    const alreadyApplied = snapshot === after;
    const statementText = alreadyApplied ? site.after : site.before;
    const statements = [];
    function visit(node) {
        if (ts.isExpressionStatement(node) && node.getText(source) === statementText) {
            statements.push(node);
        }
        ts.forEachChild(node, visit);
    }
    visit(owner);
    assert.equal(statements.length, 1, `expected one truncation statement: ${site.owner}`);
    if (alreadyApplied) return text;

    const statement = statements[0];
    return text.slice(0, statement.getStart(source)) + site.after + text.slice(statement.end);
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const plans = sites.map(site => {
        const file = path.join(tree, 'src/compiler', site.file);
        const before = fs.readFileSync(file, 'utf8');
        return {file, before, after: plan(file, before, site)};
    });
    const edits = plans.filter(({before, after}) => before !== after);
    assert(edits.length === 0 || edits.length === plans.length, 'partially applied truncations');
    for (const {file, before} of edits) {
        assert.equal(fs.readFileSync(file, 'utf8'), before, `concurrent edit: ${file}`);
    }
    for (const {file, after} of edits) {
        fs.writeFileSync(file, after);
    }
    return {files: edits.length, edits: edits.length, javascriptChanging: true};
}

module.exports = {adapt};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
