// Canonical boundary observations for Go's own command tests. No Go answers enter this program.
import { programArguments, panic } from 'adamic';
import { parseJson, quoted } from '../config/json.ts';
import { discoveryRoot, discoveryApplies } from './discovery.ts';
import { Location, locateProject } from './location.ts';
import { namedPathsScope, wholeTreeScope, describeEnumeration } from './scope.ts';
import { Enumeration } from '../formatfiles/enumerate.ts';
import { resolveOwnership, childArguments, type Project } from './ownership.ts';
import { compareCodePoints } from '../formatfiles/golang.ts';
import { parseArguments } from './flags.ts';
const document = parseJson(programArguments()[0] ?? panic('missing request'), false);
if(document.error !== '') {
    panic(document.error);
}
function text(key: string): string {
    return document.string(document.get(0, key));
}
function truth(key: string): boolean {
    return text(key) === 'true';
}
function strings(key: string): string[] {
    return document.strings(document.get(0, key));
}
function number(key: string): number {
    return Number(document.node(document.get(0, key)).raw);
}
const kind = text('kind');
if(kind === 'discovery-root') {
    const location = new Location();
    location.root = text('root');
    location.error = text('error');
    console.log(quoted(discoveryRoot(location, text('cwd'))));
}
else if(kind === 'discovery-applies') {
    const given = new Map<string, boolean>();
    for(const name of strings('given')) {
        given.set(name, true);
    }
    console.log(String(discoveryApplies(given, strings('names'))));
}
else if(kind === 'ownership') {
    const projects: Project[] = [];
    const index = document.get(0, 'projects');
    if(document.kind(index) === 'array') {
        for(const child of document.node(index).children) {
            projects.push({
                directory: document.string(document.get(child, 'Directory')),
                engine: document.string(document.get(child, 'Engine')),
                configFile: document.string(document.get(child, 'ConfigFile')),
            });
        }
    }
    const result = resolveOwnership(text('root'), projects, strings('refused'), strings('nested'));
    console.log(quoted(result.error));
    for(const project of result.projects) {
        console.log(`project\t${quoted(project.directory)}\t${quoted(project.engine)}\t${quoted(project.configFile)}`);
    }
    for(const label of result.solutions) {
        console.log(`solution\t${quoted(label)}`);
    }
    for(const label of result.yielding) {
        console.log(`yielding\t${quoted(label)}`);
    }
    const labels: string[] = [];
    for(const label of result.yields.keys()) {
        labels.push(label);
    }
    labels.sort(compareCodePoints);
    for(const label of labels) {
        const yieldTo = result.yields.get(label) ?? panic('missing yield');
        console.log(`yield\t${quoted(label)}`);
        for(const file of yieldTo.files) {
            console.log(`file\t${quoted(file)}`);
        }
        for(const directory of yieldTo.directories) {
            console.log(`directory\t${quoted(directory)}`);
        }
    }
    const shared: string[] = [];
    for(const label of result.sharedDirectory.keys()) {
        shared.push(label);
    }
    shared.sort(compareCodePoints);
    for(const label of shared) {
        console.log(`shared\t${quoted(label)}`);
    }
}
else if(kind === 'child') {
    for(const argument of childArguments(strings('arguments'), text('lintConfig'))) {
        console.log(quoted(argument));
    }
}
else if(kind === 'location') {
    const result = locateProject({
        workingDirectory: text('WorkingDirectory'),
        directory: text('Directory'),
        engine: text('Engine'),
        configFileName: text('ConfigFileName'),
        configFileNameGiven: truth('ConfigFileNameGiven'),
        lintConfigFileName: text('LintConfigFileName'),
        lintConfigFileNameGiven: truth('LintConfigFileNameGiven'),
    });
    console.log(
        [
            quoted(result.error),
            quoted(result.root),
            quoted(result.engine),
            quoted(result.configFileName),
            quoted(result.lintConfigFileName),
            String(result.lintConfigFileNameGiven),
            quoted(result.argumentBase),
            quoted(result.workingDirectory),
            quoted(result.error === '' ? result.rootNote() : ''),
        ].join('\t'),
    );
}
else if(kind === 'flags') {
    const result = parseArguments(text('program'), strings('arguments'));
    console.log(`${result.exitCode}\t${result.stopped}\t${quoted(result.stderr)}`);
    for(const [name, value] of result.values) {
        console.log(`${name}\t${quoted(value)}\t${result.given.has(name)}`);
    }
    for(const name of result.names) {
        console.log(`arg\t${quoted(name)}`);
    }
}
else {
    const initial = kind === 'named' ? namedPathsScope(text('cwd'), text('root'), strings('names')) : wholeTreeScope();
    let result = initial;
    if(kind === 'narrow') {
        initial.everything = truth('everything');
        initial.description = text('description');
        initial.requestDescription = text('request');
        initial.fileNames = strings('files');
        result = initial.narrowTo(strings('population'));
        console.log(quoted(initial.description));
    }
    if(kind === 'enumeration') {
        initial.everything = truth('everything');
        initial.description = text('description');
        initial.requestDescription = text('request');
        initial.fileNames = strings('files');
        const enumeration = new Enumeration(text('root'));
        enumeration.walked = number('walked');
        enumeration.symbolicLinks = number('symbolicLinks');
        for(const name of strings('handled')) {
            enumeration.files.push(name);
        }
        for(const name of strings('nested')) {
            enumeration.nestedRepositories.push(name);
        }
        for(const name of strings('layers')) {
            const fields = name.split('\t');
            enumeration.ignoredByLayer.set(fields[0] ?? '', Number(fields[1] ?? '0'));
        }
        for(const name of strings('declined')) {
            const fields = name.split('\t');
            enumeration.declinedExtensions.set(fields[0] ?? '', Number(fields[1] ?? '0'));
        }
        result = initial.narrowToEnumeration(enumeration);
        console.log(quoted(describeEnumeration(enumeration, number('formattable'))));
        console.log(quoted(initial.description));
    }
    console.log(
        [
            quoted(result.error),
            quoted(result.description),
            String(result.everything),
            String(result.dependentCount),
            quoted(result.requestDescription),
        ].join('\t'),
    );
    for(const name of result.fileNames) {
        console.log(quoted(name));
    }
}
