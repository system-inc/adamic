// narrowed_number.ts's twin over a boolean | undefined field narrowed the same way: the field's pair
// is already the false | undefined result's, so the right fits and the program agrees with Node,
// printing absent (#thwgcax).
interface Holder { kind: string; value: boolean | undefined }
function isReady(holder: Holder): boolean { return holder.kind === "ready"; }
function read(holder: Holder): void {
    if (holder.value !== undefined) return;
    if (isReady(holder) && holder.value) console.log("present");
    else console.log("absent");
}
read({ kind: "ready", value: undefined });
