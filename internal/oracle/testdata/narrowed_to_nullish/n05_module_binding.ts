import { fill, later } from './n05_module_binding/back.ts';
if (later === undefined) {
  fill();
  console.log(String(later));
}
