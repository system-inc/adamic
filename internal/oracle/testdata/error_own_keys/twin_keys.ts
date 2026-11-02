// Object.keys of a plain object through the same view, in a program with no Error and no catch.
const o: { message: string } = { message: 'x' };
console.log(`${Object.keys(o).length}`);
console.log(Object.keys(o).join(','));
