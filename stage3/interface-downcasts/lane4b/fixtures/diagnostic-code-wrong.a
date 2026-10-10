interface Chain { readonly messageText: string; readonly code: number; }
interface Base { readonly kind: string; }
interface Box extends Base { readonly kind: 'Box'; readonly messageText: string | Chain; }
function read(value: Box): string {
 const member = value.messageText;
 return typeof member === 'string' ? member : `${member.code}`;
}
const concrete0 = { kind: 'Box', messageText: { messageText: 'nested', code: false } };
const source0: Base = concrete0;
console.log(read(source0 as Box));
