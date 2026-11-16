async function main(): Promise<void> {
  const inner = Promise.resolve(3);
  console.log(await Promise.resolve(inner));
}
void main();
