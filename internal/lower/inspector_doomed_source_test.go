package lower

// Exact TypeScript 6.0.3 sys.ts:1645-1666, including source columns.
const inspectorEnableCPUProfiler = `        function enableCPUProfiler(path: string, cb: () => void) {
            if (activeSession) {
                cb();
                return false;
            }
            const inspector: typeof import("inspector") = require("inspector");
            if (!inspector || !inspector.Session) {
                cb();
                return false;
            }
            const session = new inspector.Session();
            session.connect();

            session.post("Profiler.enable", () => {
                session.post("Profiler.start", () => {
                    activeSession = session;
                    profilePath = path;
                    cb();
                });
            });
            return true;
        }
`
