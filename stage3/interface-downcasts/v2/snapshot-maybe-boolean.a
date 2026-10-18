interface Base { readonly kind: number }
interface View extends Base { readonly value: boolean | string | undefined }
function read(value: boolean | undefined): void {
    const source = { kind: 1, value: 42 };
    const object = { ...source, value };
    const base: Base = object;
    const view = base as View;
    console.log(`${view.value === true} / ${view.value}`);
}
read(true);
read(false);
read(undefined);
