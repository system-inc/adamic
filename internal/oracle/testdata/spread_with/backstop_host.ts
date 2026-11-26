// A host value, process.memoryUsage()'s, through { readonly rss: number }: a Node type in the closed world the
// spread's view may hold, so checkLibrarySpreads refuses the spread at compile time, on both backends (#k7g6f9h's
// storage pins). The import brings in Node's types, as node_process_handle_asserted.ts does.
import 'node:process';
function copy(source: { readonly rss: number }): void {
    const made = { ...source, k: 1 };
    console.log(typeof made.rss);
}
copy({ rss: 2 });
copy(process.memoryUsage());
