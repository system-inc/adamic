let value = 4;
for (let i = 0; i < 3; i++) {
 value = undefined!;
 value = i;
 console.log(`${value}`);
 value = null!;
 value = i + 10;
 console.log(`${value}`);
}
