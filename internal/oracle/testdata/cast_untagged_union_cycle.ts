// A link stored through an untagged union's view into a value of the other member (#axqgdck, Language's y18 on
// #03e5ke4) makes a cycle the program never reads back. Holder's and Named's link hold one type, so the view's arms
// pair and the graph program lowers; the counted build frees it whole, held in graph regions, as Node prints.
interface P { owner: Holder | undefined }
interface Holder { readonly id: number; link: P }
interface Named { readonly name: string; link: P }
function made(which: number): Holder | Named {
    const holder: Holder = { id: 1, link: { owner: undefined } };
    const named: Named = { name: 'b', link: { owner: undefined } };
    return which === 0 ? holder : named;
}
function tie(holder: Holder): void {
    holder.link = { owner: holder };
}
tie(made(0) as Holder);
const tied = made(1);
tie(tied as Holder);
console.log((tied as Named).name);
