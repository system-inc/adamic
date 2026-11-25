// A spread whose elements have another representation from the array's, strings into objects that have a
// length, stops as push stops it (#p1qw4kf).
const sized: { readonly length: number }[] = [{ length: 1 }];
const names = ['ab', 'cde'];
sized.splice(1, 0, ...names);
console.log(String(sized.length));
