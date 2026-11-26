// A host value, process.memoryUsage()'s, through { readonly rss: number }: lowering stops it where it's passed,
// before the spread, since a host object seen as a plain object type isn't lowered yet (#k7g6f9h's storage pins).
// checkLibrarySpreads would refuse the spread too, so a host value never reaches the run-time stop. The import
// brings in Node's types, as node_process_handle_asserted.ts does.
import 'node:process';
function copy(source: { readonly rss: number }): void {
    const made = { ...source, k: 1 };
    console.log(typeof made.rss);
}
copy({ rss: 2 });
copy(process.memoryUsage());
