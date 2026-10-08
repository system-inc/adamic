function f(): void {
 let value = 4;
 const clear = (): void => { value = undefined!; };
 value = 5;
 clear();
 console.log('before');
 console.log(`${value}`);
}
f();
