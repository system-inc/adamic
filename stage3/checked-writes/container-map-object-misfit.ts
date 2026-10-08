interface Diagnostic { file: { readonly name: string } | undefined; start: number | undefined }
interface DiagnosticWithLocation { file: { readonly name: string }; start: number }
const narrow: Map<string, DiagnosticWithLocation> = new Map<string, DiagnosticWithLocation>();
const good: DiagnosticWithLocation = { file: { name: 'new'.repeat(2) }, start: 1 };
const bad: Diagnostic = { file: undefined, start: undefined };
function store(values: Map<string, Diagnostic>, value: Diagnostic): void { values.set('key', value); }
store(narrow, bad);
console.log(narrow.size.toString());
