import type { Weak } from 'adamic';
interface Box { readonly label: string; }
interface Holder { value: Weak<Box>; }
let held: Box | undefined = { label: 'live'.repeat(2) };
const holder: Holder = { value: held };
if (holder.value !== undefined) {
    held = undefined;
    console.log('before');
    console.log(holder.value!.label);
}
