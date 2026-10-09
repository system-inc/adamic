interface Base { readonly kind: number }
interface View extends Base { readonly value: number | string | undefined }
function read(value: number | undefined): void {
    const source = { kind: 1, value: 42 };
    const object = { ...source, value };
    const base: Base = object;
    const view = base as View;
    console.log(`${view.value}`);
}
read(7);
read(0);
read(undefined);
