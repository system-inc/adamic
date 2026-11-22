// The checker narrows holder.value to undefined past the early return, so isReady(holder) &&
// holder.value is false | undefined, a boolean | undefined result, while the field is still held as
// its number | undefined pair. That pair can't be held as the result, so lowering stops at the &&
// with both named, where it once emitted C clang refused (#thwgcax). Node prints absent.
interface Holder { kind: string; value: number | undefined }
function isReady(holder: Holder): boolean { return holder.kind === "ready"; }
function read(holder: Holder): void {
    if (holder.value !== undefined) return;
    if (isReady(holder) && holder.value) console.log("present");
    else console.log("absent");
}
read({ kind: "ready", value: undefined });
