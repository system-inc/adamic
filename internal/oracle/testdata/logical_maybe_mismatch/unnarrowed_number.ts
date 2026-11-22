// narrowed_number.ts without the early return: the && is false | number | undefined, a union, and
// the field's pair is boxed into it, so it lowers and agrees with Node: absent for 0, present for 7
// (#thwgcax).
interface Holder { kind: string; value: number | undefined }
function isReady(holder: Holder): boolean { return holder.kind === "ready"; }
function read(holder: Holder): void {
    if (isReady(holder) && holder.value) console.log("present");
    else console.log("absent");
}
read({ kind: "ready", value: 0 });
read({ kind: "ready", value: 7 });
