#!/usr/bin/env python3
"""Seats: the 26 causes of the 58 unheld firm walls (#bcbp175, firm-walls-plan-1600.md), read from first stops.

Usage: causes.py [file ...]   (no file: stdin)
Reads firststops.py's lines, `FS pass file unit kind ordinal finding-kind where reason`, tab-separated, from a kit's
own output or a job's log (anything before `FS<tab>` on a line is ignored, so a log prefix does no harm). Every other
line is skipped. A unit is keyed by file, unit name, kind and ordinal, as firststops.py and units-compare.py key it.
When both passes are read, a unit's first stop is taken from each, and a unit whose two passes disagree is flagged.

For each cause U1 to U26 it prints the units whose first stop's reason the cause's pattern finds (`re.search`, the
plan's section 4 patterns as written), each with its file, unit and first-stop site; the count of those units; and
which of the 58 walls' sites (the first link's, J-nsy6d40, by `compiler/<path>:<line>:<column>`) are among those first
stops, each marked with the cause it was walled under. It ends with one line per cause,
`CAUSE U<n> units=<k> sites=<s>`, where k is the units counted and s the walls' sites among their first stops.
Exits 1 when it reads no FS line, so an unread log never passes as an empty table.
"""
import collections, re, sys

# The plan's section 4 patterns, exactly.
patterns = {
    "U1": r"^a value without nominal ancestry seen as ",
    "U2": r"\(tsc relates a method's parameters both ways\)",
    "U3": r"^a checked cast from .*(which needs Array\.isArray to tell|which needs a typeof test before its view)",
    "U4": r"^a checked view of the readonly array .*'s elements over ",
    "U5": r"^a checked view of .* over .*, which reads the field .* narrower$",
    "U6": r"^a cast that makes a viewed readonly array writable$",
    "U7": r"^a structural relation deeper than 64 type pairs while checking (members typed any|void function results)$",
    "U8": r"^an assignment used as a value (whose target write type|without a known target write type)",
    "U9": r"^a void-returning function at the .* site",
    "U10": r"^a read of a void value, as a returned value$",
    "U11": r"^an indexed mutation without a closed const literal origin$",
    "U12": r"^a spread argument .*'s type argument can be inferred from, judged against its constraint",
    "U13": r"^a checked predicate result read by a ConditionalExpression inside a condition",
    "U14": r"^a method in object destructuring$",
    "U15": r"^var$",
    "U16": r"^in when a possibly absent spread has no property presence descriptors$",
    "U17": r"^a case whose type differs from the switch's$",
    "U18": r"^a call whose argument for a never parameter isn't a rulings\.md:76 never read$",
    "U19": r"^an overloaded function as a value$",
    "U20": r"run-time key in a program with a class getter that does more than",
    "U21": r"^\?\. to a .*, which would be .* \| undefined$",
    "U22": r"^JSON\.parse: its result's type can't be proven from the text$",
    "U23": r"^a conditional whose branches have different types$",
    "U24": r"^a .* seen as a .* \(a tuple is held as an object, not an array, so far",
    "U25": r"^a \w+ callback that doesn't return a boolean$",
    "U26": r"JSON\.stringify recursive array types$",
}
matchers = {cause: re.compile(pattern) for cause, pattern in patterns.items()}

# The 58 walls by cause, the plan's section 1 (re-parsed from firstlink-row1-J-nsy6d40.md on #bcbp175, each checked
# present there). Sites are under `compiler/`, as the first link writes them.
walls = {
    "U1": ["factory/utilities.ts:1287:27", "factory/utilities.ts:1300:27", "factory/utilities.ts:1318:27",
           "factory/utilities.ts:1332:27", "factory/utilities.ts:1350:27", "factory/utilities.ts:1371:27",
           "factory/utilities.ts:1399:34"],
    "U2": ["factory/parenthesizerRules.ts:75:9", "factory/nodeFactory.ts:778:9", "factory/emitHelpers.ts:156:9",
           "checker.ts:1687:42", "tsbuildPublic.ts:906:22"],
    "U3": ["emitter.ts:1146:12", "builder.ts:1489:32", "commandLineParser.ts:3012:28", "commandLineParser.ts:3299:87",
           "watchPublic.ts:374:27", "parser.ts:9681:47", "transformers/generators.ts:3048:35"],
    "U4": ["builder.ts:407:46"],
    "U5": ["builder.ts:2262:9"],
    "U6": ["utilities.ts:3796:13", "parser.ts:9494:39", "transformers/utilities.ts:600:12", "transformers/ts.ts:1384:13"],
    "U7": ["transformers/declarations.ts:701:98", "expressionToTypeNode.ts:685:47", "expressionToTypeNode.ts:797:89",
           "expressionToTypeNode.ts:828:89", "binder.ts:589:38"],
    "U8": ["binder.ts:984:25", "program.ts:1140:28", "builder.ts:2504:31", "factory/nodeFactory.ts:7497:37"],
    "U9": ["watch.ts:700:56"],
    "U10": ["utilities.ts:10859:65", "watch.ts:1194:28"],
    "U11": ["commandLineParser.ts:3642:17", "commandLineParser.ts:3814:13"],
    "U12": ["utilities.ts:9044:33"],
    "U13": ["utilities.ts:3035:63", "utilities.ts:3498:43"],
    "U14": ["sys.ts:599:5", "sys.ts:1007:5"],
    "U15": ["scanner.ts:1043:5", "scanner.ts:2629:9"],
    "U16": ["hostErrors.ts:3:59", "hostErrors.ts:8:59"],
    "U17": ["semver.ts:380:13"],
    "U18": ["utilities.ts:2405:42"],
    "U19": ["scanner.ts:3740:64"],
    "U20": ["debug.ts:408:27"],
    "U21": ["tracing.ts:243:46"],
    "U22": ["utilitiesPublic.ts:757:44"],
    "U23": ["core.ts:1801:12"],
    "U24": ["utilitiesPublic.ts:721:18"],
    "U25": ["scanner.ts:3017:107"],
    "U26": ["tracing.ts:187:19"],
}
wallCause = {"compiler/" + site: cause for cause, sites in walls.items() for site in sites}
assert len(wallCause) == sum(len(sites) for sites in walls.values()) == 58, "the 58 walls, each once"
assert list(patterns) == list(walls) == [f"U{n}" for n in range(1, 27)]

# A where as `path:line:column` or `path(line, column)`; the first is the first link's form.
paren = re.compile(r"^(.*)\((\d+),\s*(\d+)\)$")


def site(where):
    match = paren.match(where)
    return f"{match.group(1)}:{match.group(2)}:{match.group(3)}" if match else where


def wall(where):
    """The wall site a first stop's where names, or None. The census's where is under the adapted tree (`src/compiler/...`)."""
    where = site(where)
    for key in wallCause:
        if where == key or where.endswith("/" + key):
            return key
    return None


lines = []
if len(sys.argv) > 1:
    for name in sys.argv[1:]:
        with open(name, encoding="utf-8", errors="replace") as handle:
            lines.extend(handle.read().splitlines())
else:
    lines = sys.stdin.read().splitlines()

# first[unit][pass] = (finding kind, where, reason)
first = collections.defaultdict(dict)
read = 0
for line in lines:
    at = line.find("FS\t")
    if at < 0 or (at > 0 and not line[at - 1].isspace() and line[at - 1] not in "]|:"):
        continue
    fields = line[at:].split("\t", 8)
    if len(fields) != 9 or fields[1] not in ("entry", "directory"):
        continue
    _, pass_name, file, unit, kind, ordinal, finding_kind, where, reason = fields
    read += 1
    first[(file, unit, kind, ordinal)][pass_name] = (finding_kind, where, reason)

if read == 0:
    print("FAIL causes: no FS line read")
    sys.exit(1)

passes = sorted({pass_name for stops in first.values() for pass_name in stops})
disagree = sorted(unit for unit, stops in first.items() if len(set(stops.values())) > 1)
print(f"  FS lines read {read}, units {len(first)}, passes {', '.join(passes)}; units whose passes disagree {len(disagree)}")
for unit in disagree:
    print(f"    DIFFER {unit[0]} {unit[1]} ({unit[2]} #{unit[3]})")
    for pass_name, stop in sorted(first[unit].items()):
        print(f"        {pass_name}: {stop[0]} {stop[1]} {stop[2][:200]}")

# A unit counts for a cause when any pass's first stop matches; each matching stop is listed.
byCause = collections.defaultdict(list)
causesOf = collections.defaultdict(set)
for unit, stops in sorted(first.items()):
    for pass_name, stop in sorted(stops.items()):
        for cause, matcher in matchers.items():
            if matcher.search(stop[2]):
                byCause[cause].append((unit, pass_name, stop))
                causesOf[unit].add(cause)

summary = []
for cause in patterns:
    hits = byCause[cause]
    units = sorted({unit for unit, _, _ in hits})
    stopSites = sorted({site(stop[1]) for _, _, stop in hits})
    wallSites = sorted({wall(stop[1]) for _, _, stop in hits} - {None})
    print(f"== {cause}  /{patterns[cause]}/")
    for unit, pass_name, stop in hits:
        print(f"    {unit[0]}\t{unit[1]} ({unit[2]} #{unit[3]})\t{pass_name}\tfirst stop {stop[0]} {site(stop[1])}\t{stop[2][:200]}")
    print(f"  units {len(units)}; distinct first-stop sites {len(stopSites)}")
    own = [key for key in wallSites if wallCause[key] == cause]
    print(f"  walls among these first stops {len(wallSites)} (of this cause's own {len(own)} of {len(walls[cause])})")
    for key in wallSites:
        print(f"    wall {key} (walled under {wallCause[key]})")
    summary.append(f"CAUSE {cause} units={len(units)} sites={len(wallSites)}")

overlapping = sorted(unit for unit, causes in causesOf.items() if len(causes) > 1)
print(f"  units matched by more than one cause {len(overlapping)}")
for unit in overlapping:
    print(f"    OVERLAP {unit[0]} {unit[1]} ({unit[2]} #{unit[3]}): {', '.join(sorted(causesOf[unit], key=lambda c: int(c[1:])))}")
print(f"  units matched by no cause {len(set(first) - set(causesOf))}")
for line in summary:
    print(line)
