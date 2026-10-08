class Box {
 value = 4;
 clear(): void { this.value = undefined!; }
 read(): number { return this.value; }
}
const box = new Box();
box.clear();
console.log('before');
console.log(`${box.read()}`);
