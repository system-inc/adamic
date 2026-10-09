package benchresults

import (
	"path/filepath"
	"strings"
	"testing"
)

// Every committed record is the honest account of its own runs, and is named for itself.
func TestEveryCommittedRecordValidates(t *testing.T) {
	t.Parallel()
	if _, err := Read(filepath.Join("..", "..", Directory)); err != nil {
		t.Fatal(err)
	}
}

// honestRecord is a record that validates: two rounds at an idle floor of 90%, with Cold quiet once, Warm
// quiet twice and Edit never. Its loads are recorded and decide nothing.
func honestRecord() Record {
	commit := strings.Repeat("a", 40)
	runs := []Run{
		{Mode: ModeWarm, Round: 1, VerdictSeconds: 0.11, SettledSeconds: 0.2, EngineSeconds: 0.08, LoadBefore: 6, LoadAfter: 6, IdleBefore: 96, IdleAfter: 95, Cache: "replay"},
		{Mode: ModeEdit, Round: 1, VerdictSeconds: 0.7, SettledSeconds: 0.9, EngineSeconds: 0.65, LoadBefore: 3, LoadAfter: 5, IdleBefore: 95, IdleAfter: 70, Cache: "files 3891/3976"},
		{Mode: ModeCold, Round: 1, VerdictSeconds: 1.8, SettledSeconds: 2.1, EngineSeconds: 1.75, LoadBefore: 3.5, LoadAfter: 4, IdleBefore: 92, IdleAfter: 91, Cache: "off"},
		{Mode: ModeWarm, Round: 2, VerdictSeconds: 0.1, SettledSeconds: 0.2, EngineSeconds: 0.07, LoadBefore: 1, LoadAfter: 1.5, IdleBefore: 97, IdleAfter: 97, Cache: "replay"},
		{Mode: ModeEdit, Round: 2, VerdictSeconds: 0.68, SettledSeconds: 0.8, EngineSeconds: 0.62, LoadBefore: 2, LoadAfter: 2, IdleBefore: 80, IdleAfter: 93, Cache: "files 3891/3976"},
		{Mode: ModeCold, Round: 2, VerdictSeconds: 1.9, SettledSeconds: 2.2, EngineSeconds: 1.85, LoadBefore: 2, LoadAfter: 2.5, IdleBefore: 89.9, IdleAfter: 95, Cache: "off"},
	}
	for index := range runs {
		// Round 2 needed a second prime, as #547dhjz makes a round do, and the record says so.
		runs[index].Primes = runs[index].Round
		runs[index].Quiet = runs[index].IdleBefore >= 90 && runs[index].IdleAfter >= 90
	}
	measured := []Mode{ModeCold, ModeWarm, ModeEdit}
	return Record{
		Schema:        SchemaVersion,
		RecordedAt:    "2026-10-05T10:00:00Z",
		Cohere:        Cohere{Commit: commit, Version: "cohere dev"},
		Project:       Project{Name: "ahra", Commit: strings.Repeat("b", 40)},
		Machine:       Machine{Model: "Mac16,5", Cores: 16},
		IdleFloor:     90,
		SettleSeconds: 60,
		Rounds:        2,
		MeasuredModes: measured,
		Band:          10,
		Edit:          "modules/tasks/TasksSearchQuery.ts",
		Runs:          runs,
		Modes:         Summarize(measured, runs),
	}
}

// loadSchemaRecord is a schema 1 record that validates by its own rules: quiet by loads against a ceiling of
// 4, every round primed, all three modes measured.
func loadSchemaRecord() Record {
	record := honestRecord()
	record.Schema, record.IdleFloor, record.LoadCeiling, record.MeasuredModes, record.Band = loadSchemaVersion, 0, 4, nil, 0
	for index := range record.Runs {
		record.Runs[index].IdleBefore, record.Runs[index].IdleAfter = 0, 0
		record.Runs[index].Quiet = record.Runs[index].LoadBefore <= 4 && record.Runs[index].LoadAfter <= 4
	}
	record.Modes = Summarize(loadSchemaModes, record.Runs)
	return record
}

// coldOnlyRecord is a schema 2 record that measured Cold alone and stopped after three rounds, because its
// three quiet runs agreed within 5%. It read no cache, so it primed nothing.
func coldOnlyRecord() Record {
	record := honestRecord()
	record.MeasuredModes, record.Edit, record.Band, record.StoppedEarly, record.Rounds = []Mode{ModeCold}, "", 5, true, 3
	record.Runs = []Run{}
	for round, seconds := range []float64{1.52, 1.55, 1.58} {
		record.Runs = append(record.Runs, Run{Mode: ModeCold, Round: round + 1, VerdictSeconds: seconds, SettledSeconds: seconds + 0.04,
			EngineSeconds: seconds - 0.01, LoadBefore: 4, LoadAfter: 5, IdleBefore: 95, IdleAfter: 94, Quiet: true, Cache: "off"})
	}
	record.Modes = Summarize(record.MeasuredModes, record.Runs)
	return record
}

func TestAnHonestRecordValidates(t *testing.T) {
	t.Parallel()
	record := honestRecord()
	if err := Validate(record); err != nil {
		t.Fatal(err)
	}
	if record.Modes[ModeEdit].Quiet != nil {
		t.Fatalf("Edit had no quiet run and carries a quiet number: %+v", record.Modes[ModeEdit].Quiet)
	}
	if quiet := record.Modes[ModeCold].Quiet; quiet == nil || quiet.Count != 1 || quiet.Best != 1.8 {
		t.Fatalf("Cold's one quiet run gave %+v, want a count of 1 at 1.8s", quiet)
	}
	if quiet := record.Modes[ModeWarm].Quiet; quiet == nil || quiet.Count != 2 {
		t.Fatalf("Warm's two quiet runs, at loads of 6 that decide nothing here, gave %+v", quiet)
	}
	if got := FileName(record); got != "2026-10-05-aaaaaaaaaaaa-mac16-5.json" {
		t.Errorf("named %s", got)
	}
}

// A schema 1 record, as committed before idleness decided, still validates by the loads it was judged by.
func TestALoadSchemaRecordStillValidates(t *testing.T) {
	t.Parallel()
	if err := Validate(loadSchemaRecord()); err != nil {
		t.Fatal(err)
	}
}

// A record may measure Cold alone, read no cache and so prime nothing, and stop early when its quiet runs agree.
func TestAColdOnlyRecordThatStoppedEarlyValidates(t *testing.T) {
	t.Parallel()
	if err := Validate(coldOnlyRecord()); err != nil {
		t.Fatal(err)
	}
}

// Each way a record could claim more than its runs measured is refused.
func TestADishonestRecordIsRefused(t *testing.T) {
	t.Parallel()
	for name, mutate := range map[string]func(*Record){
		"a quiet number on a mode with no quiet run": func(record *Record) {
			summary := record.Modes[ModeEdit]
			summary.Quiet = &QuietSummary{Count: 1, Best: 0.68, Median: 0.68, Worst: 0.68}
			record.Modes[ModeEdit] = summary
		},
		"a run marked quiet under the floor": func(record *Record) {
			record.Runs[1].Quiet = true
			record.Modes = Summarize(record.MeasuredModes, record.Runs)
		},
		"a quiet run marked loaded": func(record *Record) {
			record.Runs[0].Quiet = false
			record.Modes = Summarize(record.MeasuredModes, record.Runs)
		},
		"a run quiet by its loads but not its idleness": func(record *Record) {
			record.Runs[5].Quiet = true
			record.Modes = Summarize(record.MeasuredModes, record.Runs)
		},
		"a median the runs do not give": func(record *Record) {
			record.Modes[ModeWarm].Quiet.Median = 0.09
		},
		"a loaded run left out of the loaded list": func(record *Record) {
			summary := record.Modes[ModeEdit]
			summary.LoadedVerdictSeconds = summary.LoadedVerdictSeconds[:1]
			record.Modes[ModeEdit] = summary
		},
		// The first real record (91cf8e39) read 'none' on every replay, since the instrument grepped a footer
		// that had moved behind --verbose (#zrgrk14).
		"a warm run that only took every file from the cache": func(record *Record) { record.Runs[0].Cache = "files 3976/3976" },
		"a replay whose cache read none":                      func(record *Record) { record.Runs[0].Cache = "none" },
		"an edit run that replayed nothing":                   func(record *Record) { record.Runs[1].Cache = "none" },
		"an edit run that replayed everything":                func(record *Record) { record.Runs[1].Cache = "files 3976/3976" },
		"a cold run that read the cache":                      func(record *Record) { record.Runs[2].Cache = "files 3891/3976" },
		"a round whose runs disagree on its primes":           func(record *Record) { record.Runs[1].Primes = 2 },
		"a run with no primes recorded":                       func(record *Record) { record.Runs[0].Primes = 0 },
		"a run whose output was not read":                     func(record *Record) { record.Runs[0].EngineSeconds = 0 },
		"an engine from a modified tree":                      func(record *Record) { record.Cohere.Dirty = true },
		"an abbreviated commit":                               func(record *Record) { record.Cohere.Commit = record.Cohere.Commit[:12] },
		"a missing round": func(record *Record) {
			record.Runs = record.Runs[:5]
			record.Modes = Summarize(record.MeasuredModes, record.Runs)
		},
		"an idle floor of zero":       func(record *Record) { record.IdleFloor = 0 },
		"an idle floor over 100%":     func(record *Record) { record.IdleFloor = 101 },
		"another schema":              func(record *Record) { record.Schema = SchemaVersion + 1 },
		"no measured modes":           func(record *Record) { record.MeasuredModes = nil },
		"a measured mode named twice": func(record *Record) { record.MeasuredModes = append(record.MeasuredModes, ModeCold) },
		"a run of a mode not measured": func(record *Record) {
			record.MeasuredModes = []Mode{ModeWarm, ModeEdit}
			record.Modes = Summarize(record.MeasuredModes, record.Runs)
		},
		"an edit file with Edit not measured": func(record *Record) {
			*record = coldOnlyRecord()
			record.Edit = "modules/tasks/TasksSearchQuery.ts"
		},
		"Edit measured with no edit file": func(record *Record) { record.Edit = "" },
		"an early stop with too few quiet runs": func(record *Record) {
			record.StoppedEarly = true
		},
		"an early stop wider than its band": func(record *Record) {
			*record = coldOnlyRecord()
			record.Band = 2
		},
		"an early stop with no band": func(record *Record) {
			*record = coldOnlyRecord()
			record.Band = 0
		},
		"a cold-only round that primed": func(record *Record) {
			*record = coldOnlyRecord()
			for index := range record.Runs {
				record.Runs[index].Primes = 1
			}
		},
		"an idle run that read the cache": func(record *Record) {
			*record = coldOnlyRecord()
			record.MeasuredModes = []Mode{ModeIdle}
			for index := range record.Runs {
				record.Runs[index].Mode, record.Runs[index].Cache = ModeIdle, "replay"
			}
			record.Modes = Summarize(record.MeasuredModes, record.Runs)
		},
		"a schema 1 run marked quiet over the ceiling": func(record *Record) {
			*record = loadSchemaRecord()
			record.Runs[1].Quiet = true
			record.Modes = Summarize(loadSchemaModes, record.Runs)
		},
		"a schema 1 ceiling of zero": func(record *Record) {
			*record = loadSchemaRecord()
			record.LoadCeiling = 0
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			record := honestRecord()
			mutate(&record)
			if err := Validate(record); err == nil {
				t.Fatal("validated")
			}
			if _, err := Encode(record); err == nil {
				t.Fatal("encoded")
			}
		})
	}
}
