async function main(): Promise<void> {
  const values = await Promise.all([Promise.resolve(1), Promise.resolve(2)]);
  console.log(values[0] + values[1]);
}
void main();
