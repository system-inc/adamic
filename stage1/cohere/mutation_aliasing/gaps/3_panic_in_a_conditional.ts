// panic as one arm of a conditional expression.
import { panic } from 'adamic';

type GapType = 'LoopCarriedInversion';
const gaps: GapType[] = ['LoopCarriedInversion'];
for(const gap of gaps) {
    console.log(gap === 'LoopCarriedInversion' ? 'loop-carried-inversion' : panic(`an unnamed gap: ${gap}`));
}
