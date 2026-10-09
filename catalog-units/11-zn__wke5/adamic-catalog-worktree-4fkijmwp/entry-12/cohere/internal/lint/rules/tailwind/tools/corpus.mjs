/*
 * Paths inside a corpus, for the generators (#sycrdr6).
 *
 * A fixture that records a path inside a private checkout spells it "<corpus>:<path in it>", as in
 * "ahra:app/_theme/styles/theme.css", so the committed file reads the same on every machine and the
 * tests resolve it through internal/corpus. A generator takes the same spelling as input and records it
 * as given, so a regenerated fixture stays portable. This is internal/corpus's Locate for the Node
 * side: the corpus's variable, or else the user's corpora file, gives its root.
 */

import * as NodeFileSystem from 'node:fs';
import * as NodeOs from 'node:os';
import * as NodePath from 'node:path';

/*
 * Each corpus's name and the variable that holds its root, a copy of internal/corpus's All, held to it
 * by that package's TestTheGeneratorsNameTheSameCorpora.
 */
const corpusVariables = {
    'ahra': 'COHERE_CORPUS_AHRA',
    'structure': 'COHERE_CORPUS_STRUCTURE',
    'connected': 'COHERE_CORPUS_CONNECTED',
    'prettier-fork': 'COHERE_PRETTIER_FORK',
};

const spellingPattern = /^([a-z][a-z0-9-]+):(.*)$/;

function corporaFilePath() {
    return process.env.COHERE_CORPORA_CONFIG || NodePath.join(NodeOs.homedir(), '.config', 'cohere', 'corpora.json');
}

function corpusRoot(name, spelling) {
    if (!Object.hasOwn(corpusVariables, name)) {
        throw new Error(`${spelling} names the corpus "${name}", which is no corpus; the corpora are ${Object.keys(corpusVariables).join(', ')}`);
    }
    const variable = corpusVariables[name];
    if (process.env[variable]) {
        return NodePath.resolve(process.env[variable]);
    }
    const configPath = corporaFilePath();
    const config = NodeFileSystem.existsSync(configPath) ? JSON.parse(NodeFileSystem.readFileSync(configPath, 'utf8')) : {};
    if (!config[name]) {
        throw new Error(`${spelling} needs the ${name} corpus: set ${variable}, or name "${name}" in ${configPath}`);
    }
    return NodePath.resolve(config[name]);
}

/*
 * The path a generator's argument names, and how a fixture records it. A spelling is recorded as given.
 * Anything else is a path, resolved against the working directory and recorded as it resolved, which is
 * what every generator did before spellings existed.
 */
export function locate(argument) {
    const match = spellingPattern.exec(argument);
    if (!match) {
        const path = NodePath.resolve(argument);
        return { path, recorded: path };
    }
    const [, name, inside] = match;
    return { path: NodePath.resolve(corpusRoot(name, argument), inside), recorded: argument };
}

/*
 * How a fixture records `path`, found inside what `argument` named: in the same corpus when `argument`
 * is a spelling and `path` lies inside that corpus, and otherwise as the path itself.
 */
export function recordBeside(argument, path) {
    const match = spellingPattern.exec(argument);
    if (!match) {
        return path;
    }
    const [, name] = match;
    const inside = NodePath.relative(corpusRoot(name, argument), path);
    if (inside.startsWith('..') || NodePath.isAbsolute(inside)) {
        return path;
    }
    return `${name}:${inside.split(NodePath.sep).join('/')}`;
}
