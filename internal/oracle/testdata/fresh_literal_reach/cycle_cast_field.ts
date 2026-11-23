interface Named { readonly name: string }
interface Box { item: Named | undefined }
const box: Box = { item: undefined };
box.item = { name: "f", back: box } as Named;
console.log(box.item.name);
