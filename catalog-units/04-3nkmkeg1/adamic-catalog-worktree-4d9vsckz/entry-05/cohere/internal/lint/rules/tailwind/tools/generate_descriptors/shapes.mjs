/*
 * The arbitrary-value shapes the sweep probes every root with.
 *
 * The point of this file is coverage of the *type* space, not of the value space. The value space
 * is unbounded and sampling it is not a strategy; the claim under test is that the reading is a
 * function of the inferred data type, so the shapes are chosen to hit every type in Tailwind's
 * predicate table from several directions, plus the places where inference deliberately declines to
 * answer.
 *
 * Four groups, and each earns its place:
 *
 *   Typed hints (`[color:...]`) bypass inference and name the type outright. They are the only way
 *   to reach a type whose bare syntax another type wins first: `bg-[50%]` infers `position`, so
 *   `percentage` is only reachable for `bg` as `bg-[percentage:...]`.
 *
 *   Bare shapes are what authors actually write, and they are what tests inference itself rather
 *   than the hint parser.
 *
 *   `var()` shapes matter because `inferDataType` returns null for anything starting with `var(`
 *   (verified against the installed bundle). A root's `var()` reading is therefore its *fallback*
 *   reading, and the fallback is a distinct thing the descriptor has to carry.
 *
 *   Nonsense shapes are the control on the other three. `[anything]` types as nothing, so whatever
 *   it reads is the fallback too, and a root where nonsense and `var()` disagree is a root where
 *   the model is wrong.
 */

// Every type name in Tailwind's predicate table, used both as a hint prefix and as the enumeration
// the sweep partitions by.
export const DataTypeNames = [
    'color',
    'length',
    'percentage',
    'ratio',
    'number',
    'integer',
    'url',
    'position',
    'bg-size',
    'line-width',
    'image',
    'family-name',
    'generic-name',
    'absolute-size',
    'relative-size',
    'angle',
    'vector',
];

// One representative value per type, used to build the typed hints. Chosen so the value is also
// independently valid for the type it is hinted as, which keeps a hint from being the only thing
// making the probe work.
const representativeByType = {
    'color': ['red', 'blue', 'rebeccapurple', '#fff', '#ffff', '#112233', '#11223344', 'rgb(1_2_3)', 'rgb(1_2_3_/_0.5)', 'rgba(1,2,3,0.5)', 'hsl(1_2%_3%)', 'hsla(1,2%,3%,0.5)', 'hwb(1_2%_3%)', 'lab(1_2_3)', 'lch(1_2_3)', 'oklab(0.5_0.1_0.2)', 'oklch(0.5_0.1_20)', 'color(display-p3_1_0_0)', 'color-mix(in_oklab,red,blue)', 'light-dark(red,blue)', 'transparent', 'currentcolor', 'canvastext', 'buttonface', 'accentcolor'],
    'length': ['3px', '2rem', '0', '1.5em', '10vh', '10vw', '100vmax', '100vmin', '0.25ch', '2ex', '1lh', '1rlh', '5cm', '5mm', '5Q', '5in', '5pc', '5pt', '-3px', '.5rem', '1e2px', '2dvh', '2svw', '2lvh', '3cqw', '3cqi', '3cqmin'],
    'percentage': ['50%', '0%', '100%', '12.5%', '-25%', '.5%', '1e2%'],
    'ratio': ['16/9', '1/1', '4/3', '2.35/1', '0/1'],
    'number': ['600', '0', '1', '-3', '1.5', '2e3', '.5', '-0.25', '1e-2'],
    'integer': ['3', '0', '42', '-7', '1000'],
    'url': ['url(a.png)', 'url("b.svg")', "url('c.gif')", 'url(data:image/png;base64,iVBOR)'],
    'position': ['center', 'top', 'bottom', 'left', 'right', 'top_left', 'bottom_right', '3px_4px', '50%_50%', 'left_3px_top_4px', 'center_center'],
    'bg-size': ['cover', 'contain', 'auto', '10px_20px', 'auto_50%', '100%_auto'],
    'line-width': ['thin', 'medium', 'thick', '2px', '1', '0', 'thin_thick'],
    'image': ['linear-gradient(red,blue)', 'radial-gradient(red,blue)', 'conic-gradient(red,blue)', 'repeating-linear-gradient(red,blue)', 'repeating-radial-gradient(red,blue)', 'repeating-conic-gradient(red,blue)', 'linear-gradient(to_right,red_0%,blue_100%)', 'image-set("a.png"_1x)', 'cross-fade(url(a.png)_50%)', 'element(#a)', 'image(url(a.png))', 'url(a.png),url(b.png)'],
    'family-name': ['Arial', '"Helvetica_Neue"', 'Inter,sans-serif', "'Times_New_Roman'", 'Foo,Bar,Baz'],
    'generic-name': ['serif', 'sans-serif', 'monospace', 'cursive', 'fantasy', 'system-ui', 'ui-serif', 'ui-sans-serif', 'ui-monospace', 'ui-rounded', 'math', 'emoji', 'fangsong'],
    'absolute-size': ['xx-small', 'x-small', 'small', 'medium', 'large', 'x-large', 'xx-large', 'xxx-large'],
    'relative-size': ['larger', 'smaller'],
    'angle': ['45deg', '0.5turn', '1rad', '100grad', '-90deg', '0deg'],
    'vector': ['1_2_3', '0_0_1', '1_0_0', '0.5_0.5_0.5'],
};

/*
 * Shapes that type as nothing, and shapes that inference declines to type.
 *
 * `var(...)` is not nonsense: it is a legitimate value that `inferDataType` refuses to inspect, by
 * an explicit early return. It is separated from nonsense because a root can plausibly treat the
 * two differently, and if it does, that is a finding rather than noise.
 */
const varShapes = [
    'var(--a)',
    'var(--a,red)',
    'var(--a,3px)',
    'var(--a,50%)',
    'var(--a,var(--b))',
    'var(--spacing-4)',
    'var(--color-red-500)',
    '--a',
    '--spacing-4',
];

const nonsenseShapes = [
    'anything',
    'foo_bar',
    '"quoted"',
    "'single'",
    'a-b-c',
    '???',
    '',
    '_',
    'FooBar',
    '123abc',
    'red;blue',
    '{}',
    'attr(data-x)',
    'env(safe-area-inset-top)',
    'inherit',
    'initial',
    'unset',
    'revert',
    'revert-layer',
    'none',
    'auto',
    'normal',
    'visible',
    'hidden',
    'flex',
    'grid',
    'block',
    'start',
    'end',
    'both',
    'infinite',
    'ease-in-out',
    'cubic-bezier(0.4,0,0.2,1)',
    'steps(4,end)',
    'counter(x)',
    'toggle(a,b)',
    '3px_solid_red',
    '1px_2px_3px_red',
    'inset_0_0_red',
    'translateX(3px)',
    'rotate(45deg)',
    'matrix(1,0,0,1,0,0)',
    'blur(3px)',
    'drop-shadow(0_0_red)',
    'polygon(0_0,100%_0,100%_100%)',
    'circle(50%)',
    'path("M0_0")',
    'span_3',
    '1_/_3',
    'minmax(0,1fr)',
    'repeat(3,1fr)',
    'fit-content(3rem)',
    '1fr',
    '3fr',
    'max-content',
    'min-content',
    'stretch',
    '"a_b"_"c_d"',
    '\\2014',
    '@media',
    '&:hover',
    '--custom',
    '--a',
];

/*
 * Shapes that exercise the math-function branch.
 *
 * `inferDataType` treats a value containing `calc(`, `min(`, `clamp(` and friends as satisfying
 * `number`, `percentage`, `length` and `ratio` simultaneously, so which of those wins is decided by
 * the order of the root's own type list. That makes math the sharpest test of the model: a root
 * whose reading for `calc(...)` does not match its reading for the first math-satisfying type in its
 * list is a root where reading is not a function of type.
 */
const mathShapes = [
    'calc(1px+2px)',
    'calc(100%-2rem)',
    'calc(1*2)',
    'calc(var(--a)*2)',
    'min(1rem,2vw)',
    'max(1rem,2vw)',
    'clamp(1rem,2vw,3rem)',
    'round(1.5px)',
    'round(up,1.5px,1px)',
    'mod(5,2)',
    'rem(5,2)',
    'pow(2,3)',
    'sqrt(4)',
    'hypot(3,4)',
    'log(10)',
    'exp(2)',
    'sin(45deg)',
    'cos(45deg)',
    'tan(45deg)',
    'asin(0.5)',
    'acos(0.5)',
    'atan(0.5)',
    'atan2(1,2)',
];

/*
 * Shapes carrying a modifier, which is a separate axis from the value.
 *
 * `bg-[red]/50` and `text-[16px]/[1.5]` route through the modifier path, and the descriptor model
 * says nothing about modifiers. Including them tests whether it needs to: if a root's reading is
 * unchanged by a modifier, the model is complete without a modifier axis; if it changes, the
 * descriptor needs one and that is a change to Phase 3.
 */
const modifierSuffixes = ['', '/50', '/25', '/[0.5]', '/[var(--a)]', '/[50%]', '/none'];

/*
 * The shape list, built once.
 *
 * Each entry carries the syntax to append to a root and the type the shape is *claimed* to be, or
 * null when nothing is claimed. `claimedType` is not the prediction; it is the label the sweep uses
 * to check whether the engine's own `inferDataType` agrees, which is how a hint that silently fails
 * to parse gets caught instead of being read as evidence.
 */
function buildShapes() {
    const shapes = [];
    const seen = new Set();

    function push(syntax, claimedType, group) {
        if (seen.has(syntax)) return;
        seen.add(syntax);
        shapes.push({ syntax, claimedType, group });
    }

    for (const typeName of DataTypeNames) {
        for (const representative of representativeByType[typeName] ?? []) {
            // Bare: inference decides. The type it lands on may not be `typeName`, and that is the
            // point; `claimedType` is null so the sweep reads the engine rather than this file.
            push('[' + representative + ']', null, 'bare');
            // Hinted: the type is named outright, bypassing inference.
            push('[' + typeName + ':' + representative + ']', typeName, 'hint');
        }
        // Hinted `var()`. The only route to a specific type for a value inference refuses to read.
        push('[' + typeName + ':var(--a)]', typeName, 'hint-var');
    }

    for (const shape of varShapes) push('[' + shape + ']', null, 'var');
    for (const shape of nonsenseShapes) push('[' + shape + ']', null, 'nonsense');
    for (const shape of mathShapes) push('[' + shape + ']', null, 'math');

    // Modifiers, applied to a small spanning subset rather than the cross product, which would be
    // four times the sweep for one extra axis.
    for (const base of ['[red]', '[3px]', '[var(--a)]', '[anything]', '[50%]', '[16/9]', '[url(a.png)]', '[600]']) {
        for (const suffix of modifierSuffixes) {
            if (suffix === '') continue;
            push(base + suffix, null, 'modifier');
        }
    }

    // Bare (non-arbitrary) values, which route through theme lookup rather than the arbitrary path.
    // A descriptor that only predicts arbitrary values predicts the smaller half of the registry.
    for (const bare of ['1', '2', '3', '4', '8', '0', '0.5', '1.5', 'px', 'full', 'auto', 'none', 'sm', 'md', 'lg', 'xl', '2xl', 'xs', 'red-500', 'blue-200', 'black', 'white', 'current', 'transparent', 'inherit', 'initial', 'reverse', 'normal', 'bold', 'thin', 'center', 'start', 'end', 'left', 'right', 'top', 'bottom', 'x', 'y', 'b', 'l', 'r', 't', 'screen', 'min', 'max', 'fit', 'dvh', 'svw', 'lvh', '1/2', '1/3', '2/3', '11/12']) {
        push('-' + bare, null, 'bare-theme');
    }

    // Negative bare values, which route through the negative branch.
    for (const bare of ['1', '2', '4', '8', 'px', 'full', '1/2', '0.5']) push('--' + bare, null, 'negative');

    return shapes;
}

export const Shapes = buildShapes();

// Composes a probe class from a root and a shape. Arbitrary shapes need a dash before the bracket;
// bare shapes already carry their own leading dash.
export function probeClassName(root, shape) {
    return shape.syntax.startsWith('[') ? root + '-' + shape.syntax : root + shape.syntax;
}
