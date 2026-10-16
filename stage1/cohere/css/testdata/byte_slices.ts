import { utf8Length } from 'adamic';
import { ByteView } from '../byte_view.ts';
import { quote } from '../../selector/nodes.ts';

// Includes BMP and astral code points, combining marks, line separators, and
// byte cuts inside each code point. Negative and out-of-range cuts are included.
const inputs: string[] = ['', 'ASCII\t\n', 'aé€😀z', 'é\u2028中'];
for(const text of inputs) {
    const view = new ByteView(text);
    const length = utf8Length(text);
    for(let start = -length - 2; start <= length + 2; start++) {
        console.log(`${view.prefixLength(start)}`);
        for(let end = -length - 2; end <= length + 2; end++) {
            console.log(quote(view.slice(start, end)));
        }
    }
}
