interface Named { readonly name: string }
interface Box { item: Named | undefined }
const box: Box = { item: undefined };
const extra = { back: box };
box.item = { ...extra, name: "f" };
console.log(box.item.name);
