interface SourceFile { readonly name: string; readonly statements: readonly number[]; readonly tags: ReadonlyMap<string, string>; readonly line: () => number }
interface Diagnostic { file: SourceFile | undefined; start: number | undefined; length: number | undefined }
interface DiagnosticWithLocation { file: SourceFile; start: number; length: number }
const file: SourceFile = { name: 'source'.repeat(2), statements: [1], tags: new Map<string, string>(), line: () => 1 };
const diagnostic: DiagnosticWithLocation = { file, start: 0, length: 1 };
function clear(view: Diagnostic): void { view.file = undefined; }
clear(diagnostic);
console.log('stored');
