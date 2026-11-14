"use strict";
const fs = require("node:fs");
const output = process.env.ENV_PRESENCE_COUNTS;
if (!output) throw new Error("ENV_PRESENCE_COUNTS required");
globalThis.__envPresence1965 = function(sys) {
    const row = {pid: process.pid, truthy: !!sys, function: !!sys && typeof sys.getEnvironmentVariable === "function"};
    fs.appendFileSync(output, JSON.stringify(row) + "\n");
};
