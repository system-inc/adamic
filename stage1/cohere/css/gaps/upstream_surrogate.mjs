// This is the original library's cut at the second UTF-16 unit, not an
// expectation inferred from the Go port. Run with the pinned npm directory.
import {createRequire} from 'node:module';
const require = createRequire(`${process.argv[2]}/package.json`);
for(const parser of [require('postcss'), require('postcss-scss')]) {
    try { parser.parse('\\😀|a', {map: false}); }
    catch(error) {
        console.log(JSON.stringify({reason:error.reason,endColumn:error.endColumn,endOffset:error.input.endOffset}));
    }
}
