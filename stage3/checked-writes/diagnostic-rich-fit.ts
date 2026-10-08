interface SourceFile {
 readonly name: string;
 readonly statements: readonly number[];
 readonly tags: ReadonlyMap<string, string>;
 readonly line: () => number;
}
interface Diagnostic { file: SourceFile | undefined; start: number | undefined; length: number | undefined }
interface DiagnosticWithLocation { file: SourceFile; start: number; length: number }
const first: SourceFile = { name: 'old'.repeat(2), statements: [1], tags: new Map<string, string>(), line: () => 2 };
const second: SourceFile = { name: 'new'.repeat(2), statements: [3, 4], tags: new Map<string, string>(), line: () => 3 };
const diagnostic: DiagnosticWithLocation = { file: first, start: 0, length: 1 };
function store(view: Diagnostic, file: SourceFile | undefined, start: number | undefined, length: number | undefined): void {
 view.file = file; view.start = start; view.length = length;
}
store(diagnostic, second, 3, 4);
console.log(diagnostic.file.name + ' ' + diagnostic.file.line().toString() + ' ' + diagnostic.file.statements.length.toString() + ' ' + diagnostic.start.toString() + ' ' + diagnostic.length.toString());
store(diagnostic, first, 0 / 0, -0);
console.log(diagnostic.start.toString() + ' ' + (1 / diagnostic.length).toString());
