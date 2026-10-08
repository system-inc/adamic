interface SourceFile { readonly name: string; readonly statements: readonly number[] }
interface DiagnosticRelatedInformation { file: SourceFile | undefined; start: number | undefined; length: number | undefined }
interface DiagnosticWithLocation { file: SourceFile; start: number; length: number }
const file: SourceFile = { name: 'source'.repeat(2), statements: [1] };
const diagnostic: DiagnosticWithLocation = { file, start: 0, length: 1 };
function move(view: DiagnosticRelatedInformation, start: number | undefined): void { view.start = start; }
move(diagnostic, 7);
console.log(diagnostic.start.toString());
