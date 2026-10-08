import type { Weak } from 'adamic';
interface Box { readonly label: string; }
interface Holder { value: Weak<Box>; }
let held: Box | undefined = { label: 'live'.repeat(2) };
const holder: Holder = { value: held };
function clear(slot: Holder): void { slot.value = undefined; }
if (holder.value !== undefined) {
    clear(holder);
    console.log('before');
    console.log(`${holder.value! === undefined}`);
}
