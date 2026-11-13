# setCustomPollingValues: keep the enum immutable

JavaScript-changing, task baobab. Adamic ruled road (1) on November 13, 2026;
rulings.md:249 refuses writes into enum objects. Numeric interval customization
is deliberately no longer applied. The default intervals remain 250/500/2000.

No previous adapter pins this complete owner. Searches of stage3/adapt for
setCustomPollingValues, setCustomLevels, setLevel and getCustomPollingBasedLevels
find only timer-closures witness references. Adaptations 70/71 pin other views;
40 pins Error handling, filesystem and timer slots elsewhere in sys.ts. This
new adaptation pins the entire function before and after, with exact LF text
and SHA-256 in owners.json. It preserves CRLF and makes no write on replay.

Line-by-line behavior of the complete function:

- The signature, missing-getEnvironmentVariable guard and bare return stay the
  same. The only caller, sys.ts system initialization, ignores its void result.
- pollingIntervalChanged previously called setCustomLevels with the enum,
  parsed the three interval variables, mutated each member using the custom
  value or old member, and returned whether any nonempty variable existed.
  It now calls !!getCustomLevels for the same variables, preserving exactly
  that boolean and the reads/conversions without passing or writing the enum.
- The pollingChunkSize and unchangedPollThresholds assignment lines are
  unchanged: each calls getCustomPollingBasedLevels and falls back to its
  current table. Both bindings are still assigned on every configured call.
- getLevel still forms the uppercase suffixed environment name and reads it.
- getCustomLevels still creates an optional fresh Partial<Levels>, visits
  Low, Medium and High in order, and returns the object or undefined.
- Its setCustomLevel still reads the variable, tests nonempty text, creates
  the fresh object on demand and stores Number(customLevel) in that object.
- setCustomLevels and its nested setLevel are removed. Their only caller was
  the interval line. The removed lines fetched custom levels, called setLevel
  three times, returned true/false and wrote levels[level]. No other call exists.
- getCustomPollingBasedLevels is unchanged: it parses its own variable family,
  rebuilds when the interval flag or its custom levels exist, spreads custom
  levels over the defaults, and calls createPollingIntervalBasedLevels. The
  table keys now consistently use the unchanged enum's default intervals.

No reader, public type or other upstream source file is changed. The complete
before and after function text is retained in owners.json. Proof and raw
observations are on typescript/polling-no-enum-write-evidence.
