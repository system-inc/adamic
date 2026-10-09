'use strict';

// The development extension VS Code's test mode requires. It does nothing: the extension under test is
// the cohere .vsix installed into the profile, and suite.js drives it.
module.exports = { activate() {}, deactivate() {} };
