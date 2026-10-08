class Box { value = 4; }
const box = new Box();
const alias = box;
box.value = 5;
alias.value = null!;
console.log('before');
console.log(`${box.value}`);
