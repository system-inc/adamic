interface Chain { readonly messageText: string; }
interface Base { readonly kind: string; }
interface Box extends Base { readonly kind: 'Box'; readonly messageText: string | Chain; }
function read(value: Box): string {
 const member = value.messageText;
 return typeof member === 'string' ? member : member.messageText;
}
const concrete0 = { kind: 'Box', messageText: 'plain' };
const source0: Base = concrete0;
console.log(read(source0 as Box));
const concrete1 = { kind: 'Box', messageText: { messageText: 'nested' } };
const source1: Base = concrete1;
console.log(read(source1 as Box));
