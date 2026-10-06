# Remaining valid Date refusals

These programs pass stock tsc 6.0.3 under the runner's strict settings and run on
Node with TZ=UTC. Adamic refuses them. No language change was made to remove
these limitations. The full commands and outputs are in remaining-probes.log.

| Program | Node output | Adamic refusal |
|---|---|---|
| gaps/detached.ts | 0 | A method read as a value loses its receiver |
| gaps/prototype.ts | false | Date.prototype is not an own object field |
| gaps/generic-json.ts | custom | Generic ToPrimitive and dynamic toISOString lookup are unbuilt |
| gaps/caught-iso.ts | RangeError | Native Date validation is a panic; Node can catch its exception |

The first two reflect existing language restrictions on detached methods and
observable prototypes. Replacing those observations in a test would remove what
the test checks. The last two require behavior beyond a prelude or source-style
adaptation. Date.now(), Date(), and new Date() with no arguments also remain
refused for the documented nondeterminism reason.

The one remaining original harness restriction is
built-ins/Date/prototype/toJSON/invoke-result.js. Its first assertion passes a
nullable string to the primitive harness; its second assertion borrows toJSON
onto a custom object and expects an object result. The generic-json probe shows
why removing only that first harness diagnostic cannot make the complete test
lower.

The coercion programs from the earlier investigation are not listed as language
gaps: stock tsc rejects them with the same diagnostic, so their refusals are
correct. See report.md and triage.json for the complete Date audit.
