const children: number[] = [1];
const arguments_: number[] = [2, 3];
children.push(...arguments_);
console.log(children.join(','));
