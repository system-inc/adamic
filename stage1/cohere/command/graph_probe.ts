import { panic, programArguments, readTextFile } from 'adamic';
import { parseJson, quoted } from '../config/json.ts';
import { Scope } from './scope.ts';
import { ImportGraph, narrowToClosure, rescopeAfterRebuild, type GraphFile, type ResolvedImport } from './graph.ts';
import { runProject } from './process.ts';
const argument = programArguments()[0] ?? panic('missing request');
function request(argument: string): string {
    if(argument.startsWith('{')) return argument;
    const input = readTextFile(argument);
    if(input.kind === 'Error') panic(input.message);
    return input.text;
}
const doc = parseJson(request(argument), false);
if(doc.error !== '') panic(doc.error);
function text(key: string): string {
    return doc.string(doc.get(0, key));
}
function strings(key: string): string[] {
    return doc.strings(doc.get(0, key));
}
function truth(key: string): boolean {
    return doc.node(doc.get(0, key)).raw === 'true';
}
const kind = text('kind');
if(kind === 'process') {
    const result = runProject(
        text('executable'),
        text('root'),
        { directory: text('directory'), engine: text('engine'), configFile: text('configFile') },
        strings('arguments'),
        text('yieldFile'),
    );
    console.log(`${quoted(result.output)}\t${result.exitCode}\t${quoted(result.error)}`);
}
else if(kind === 'rebuild') {
    const result = rescopeAfterRebuild(strings('rebuilt'), strings('scoped'), truth('everything'));
    for(const name of result.files) console.log(`file\t${quoted(name)}`);
    for(const name of result.lost) console.log(`lost\t${quoted(name)}`);
}
else {
    const files: GraphFile[] = [];
    const edges: ResolvedImport[] = [];
    for(const index of doc.node(doc.get(0, 'files')).children)
        files.push({ name: doc.string(doc.get(index, 'name')), path: doc.string(doc.get(index, 'path')) });
    for(const index of doc.node(doc.get(0, 'edges')).children)
        edges.push({ from: doc.string(doc.get(index, 'from')), toward: doc.string(doc.get(index, 'toward')) });
    const graph = new ImportGraph(files, edges, text('directory'), truth('sensitive'));
    const scope = new Scope();
    scope.everything = truth('everything');
    scope.description = text('description');
    scope.requestDescription = text('request');
    for(const name of strings('named')) scope.index.set(name, true);
    const result = narrowToClosure(graph, scope, files);
    console.log(
        `${result.scope.everything}\t${result.scope.dependentCount}\t${quoted(result.note)}\t${scope.dependentCount}`,
    );
    for(const file of result.files) console.log(quoted(file.name));
}
