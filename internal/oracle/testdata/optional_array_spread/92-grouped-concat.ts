interface Box { items: string[] | undefined }
const box: Box = { items: ["a"] };
function clear(): void { box.items = undefined; }
const more: string[] = ["b"];
if (box.items !== undefined) {
    clear();
    console.log(String((box.items?.concat)(...more)));
}
