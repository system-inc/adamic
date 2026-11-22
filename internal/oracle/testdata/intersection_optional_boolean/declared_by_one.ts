// tsc's namedEvaluation.ts shape (#tq6yzqy): ExportAssignment declares isExportEquals?: boolean, and
// the member it's intersected with declares nothing by that name. Read present true, present false
// and absent. Node prints: equals true, equals false, equals undefined.
interface Statement {
	kind: number;
}
interface ExportAssignment extends Statement {
	isExportEquals?: boolean;
	expression: string;
}
type Named = ExportAssignment & { readonly label: string };
function probe(value: Named): void {
	console.log(`${value.label} ${String(value.isExportEquals)}`);
}
probe({ kind: 1, isExportEquals: true, expression: "a", label: "equals" });
probe({ kind: 1, isExportEquals: false, expression: "b", label: "equals" });
probe({ kind: 1, expression: "c", label: "equals" });
