// A spread whose elements have another representation from the array's, arrays into objects that have a
// length, stops as push stops it (#p1qw4kf).
const sized: { readonly length: number }[] = [{ length: 1 }];
const lists: number[][] = [[1], [2, 3]];
sized.splice(1, 0, ...lists);
console.log(String(sized.length));
