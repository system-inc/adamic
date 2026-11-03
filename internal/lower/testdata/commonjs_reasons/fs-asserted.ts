// A literal behind an assertion is the literal it wraps, so a require of fs keeps the generic
// reason, never the computed path's throw.
import 'node:process';
require("fs" as string);
