// The command's import-graph boundary. Inputs must come from a real compiler program, not from
// tsconfig roots or a text search for imports. Module resolution is deliberately a separate door.
import { compareCodePoints } from '../formatfiles/golang.ts';
import { absoluteFrom } from './location.ts';
import { Scope } from './scope.ts';
export type GraphFile = { name: string; path: string };
export type ResolvedImport = { from: string; toward: string };
export type Closure = { files: GraphFile[]; within: boolean };
export class ImportGraph {
    files: GraphFile[];
    edges: ResolvedImport[];
    directory: string;
    sensitive: boolean;
    constructor(files: GraphFile[], edges: ResolvedImport[], directory: string, sensitive: boolean) {
        this.files = files;
        this.edges = edges;
        this.directory = directory;
        this.sensitive = sensitive;
    }
    pathFor(name: string): string {
        const path = absoluteFrom(this.directory, name.replaceAll('\\', '/'));
        return this.sensitive ? path : path.toLowerCase();
    }
    dependentClosure(seeds: GraphFile[]): Closure {
        const emptyFiles: GraphFile[] = [];
        const emptyPaths: string[] = [];
        if(seeds.length === 0) return { files: emptyFiles, within: true };
        const ours = new Map<string, boolean>();
        for(const file of this.files) ours.set(file.path, true);
        const importers = new Map<string, string[]>();
        for(const edge of this.edges) {
            if(!ours.has(edge.from) || edge.toward === '') continue;
            const toward = this.pathFor(edge.toward);
            if(!ours.has(toward)) continue;
            const emptyRow: string[] = [];
            const row = importers.get(toward) ?? emptyRow;
            row.push(edge.from);
            importers.set(toward, row);
        }
        const seen = new Map<string, boolean>();
        let frontier: string[] = [];
        for(const seed of seeds) {
            if(seen.has(seed.path)) continue;
            seen.set(seed.path, true);
            frontier.push(seed.path);
        }
        while(frontier.length > 0) {
            const next: string[] = [];
            for(const current of frontier) {
                for(const importer of importers.get(current) ?? emptyPaths) {
                    if(seen.has(importer)) continue;
                    seen.set(importer, true);
                    if(seen.size > 500) return { files: emptyFiles, within: false };
                    next.push(importer);
                }
            }
            frontier = next;
        }
        const files: GraphFile[] = [];
        for(const file of this.files) if(seen.has(file.path)) files.push(file);
        return { files, within: true };
    }
}
export type NarrowedGraph = { scope: Scope; files: GraphFile[]; note: string };
export function narrowToClosure(graph: ImportGraph, scope: Scope, population: GraphFile[]): NarrowedGraph {
    if(scope.everything) return { scope: scope.copy(), files: population, note: '' };
    const seeds: GraphFile[] = [];
    for(const file of population) if(scope.includes(absoluteFrom(graph.directory, file.name))) seeds.push(file);
    const closure = graph.dependentClosure(seeds);
    if(!closure.within) {
        const whole = new Scope();
        whole.everything = true;
        return {
            scope: whole,
            files: population,
            note: `note: ${scope.requestDescription} reaches more than 500 files through imports, so the whole tree is checked\n`,
        };
    }
    const narrowed = scope.copy();
    narrowed.dependentCount = closure.files.length - seeds.length;
    return { scope: narrowed, files: closure.files, note: '' };
}
export function rescopeAfterRebuild(
    rebuilt: string[],
    scoped: string[],
    everything: boolean,
): { files: string[]; lost: string[] } {
    const empty: string[] = [];
    if(everything) return { files: rebuilt, lost: empty };
    const wanted = new Map<string, boolean>();
    for(const name of scoped) wanted.set(name, false);
    const files: string[] = [];
    for(const name of rebuilt) {
        if(wanted.has(name)) {
            wanted.set(name, true);
            files.push(name);
        }
    }
    const lost: string[] = [];
    for(const [name, found] of wanted) if(!found) lost.push(name);
    lost.sort(compareCodePoints);
    return { files, lost };
}
