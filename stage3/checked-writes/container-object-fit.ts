interface Diagnostic { file: { readonly name: string } | undefined; start: number | undefined }
interface DiagnosticWithLocation { file: { readonly name: string }; start: number }
const narrow: DiagnosticWithLocation[] = [];
const good: DiagnosticWithLocation = { file: { name: 'new'.repeat(2) }, start: 1 };
const bad: Diagnostic = { file: undefined, start: undefined };
function store(values: Diagnostic[], value: Diagnostic): void { values.push(value); }
store(narrow, good);
console.log(narrow.length.toString());
