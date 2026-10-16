package native

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Stage1ProfileUnit records both the source identity seen by clang and its bytes.
// Split manifests include the shared header and every stable emitted unit.
type Stage1ProfileUnit struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

// Stage1ProfileManifest binds a text profile to the bytes the compiler consumes.
// Toolchain and target checks are conservative: another host regenerates its own profile.
type Stage1ProfileManifest struct {
	Parts             []Stage1ProfileUnit `json:"profile_parts,omitempty"`
	Machine           string              `json:"machine"`
	Version           int                 `json:"version"`
	Units             []Stage1ProfileUnit `json:"emitted_units"`
	CompilerSHA256    string              `json:"compiler_sha256"`
	LinkFlags         []string            `json:"link_flags"`
	TrainingFlags     []string            `json:"training_flags"`
	TrainingLinkFlags []string            `json:"training_link_flags"`
	Runtime           string              `json:"runtime_sha256"`
	Profile           string              `json:"profile_sha256"`
	Training          string              `json:"training_manifest_sha256"`
	Compiler          string              `json:"compiler"`
	Target            string              `json:"target"`
	Flags             []string            `json:"flags"`
}

func profileHash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

// RuntimeFingerprint hashes the embedded snapshot, including every header and file boundary.
func RuntimeFingerprint() (string, error) {
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		return "", err
	}
	return runtimeFingerprint(files), nil
}

func runtimeFingerprint(files []runtimeFile) string {
	h := sha256.New()
	for _, f := range files {
		fmt.Fprintf(h, "%d:%s%d:", len(f.name), f.name, len(f.contents))
		h.Write(f.contents)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// NewStage1ProfileManifest is called by regeneration after llvm-profdata writes the text file.
func NewStage1ProfileManifest(source string, profile []byte, training string) (Stage1ProfileManifest, error) {
	return newStage1ProfileManifest(source, profile, training, Options{Release: true, Split: true})
}

func newStage1ProfileManifest(source string, profile []byte, training string, options Options) (Stage1ProfileManifest, error) {
	// The text profile has its own content hash. Its local indexed pathname is
	// not a portable build input. Record all other flags from the actual policy.
	options.Profile = ""
	options.ProfileGenerate = false
	snapshot, err := RuntimeFingerprint()
	if err != nil {
		return Stage1ProfileManifest{}, err
	}
	compiler, err := toolchain.Command("clang", "--version").CombinedOutput()
	if err != nil {
		return Stage1ProfileManifest{}, fmt.Errorf("native: compiler identity: %w", err)
	}
	compilerPath, err := toolchain.Path("clang")
	if err != nil {
		return Stage1ProfileManifest{}, err
	}
	file, err := os.Open(compilerPath)
	if err != nil {
		return Stage1ProfileManifest{}, err
	}
	defer file.Close()
	digest := sha256.New()
	if _, err = io.Copy(digest, file); err != nil {
		return Stage1ProfileManifest{}, err
	}
	units := []Stage1ProfileUnit{{Name: "main.c", SHA256: profileHash([]byte(source))}}
	if useSplitSource(source, options) {
		header, parts, err := splitC(source)
		if err != nil {
			return Stage1ProfileManifest{}, err
		}
		units = []Stage1ProfileUnit{{Name: "units.h", SHA256: profileHash([]byte(header))}}
		for _, unit := range parts {
			units = append(units, Stage1ProfileUnit{Name: unit.name, SHA256: profileHash([]byte(unit.source))})
		}
	}
	machine, err := os.Hostname()
	if err != nil {
		return Stage1ProfileManifest{}, err
	}
	trainingOptions := options
	trainingOptions.ProfileGenerate = true
	return Stage1ProfileManifest{
		Version:           3,
		Machine:           machine,
		Units:             units,
		Runtime:           snapshot,
		Profile:           profileHash(profile),
		Training:          training,
		Compiler:          string(compiler),
		CompilerSHA256:    fmt.Sprintf("%x", digest.Sum(nil)),
		Target:            goruntime.GOOS + "-" + goruntime.GOARCH,
		Flags:             sourceFlags(source, options),
		LinkFlags:         LinkFlags(options),
		TrainingFlags:     sourceFlags(source, trainingOptions),
		TrainingLinkFlags: LinkFlags(trainingOptions),
	}, nil
}

func profileMatches(record, current Stage1ProfileManifest) bool {
	return record.Version == current.Version && slices.Equal(record.Units, current.Units) &&
		record.CompilerSHA256 == current.CompilerSHA256 && slices.Equal(record.LinkFlags, current.LinkFlags) &&
		slices.Equal(record.TrainingFlags, current.TrainingFlags) && slices.Equal(record.TrainingLinkFlags, current.TrainingLinkFlags) &&
		record.Runtime == current.Runtime &&
		record.Profile == current.Profile && record.Compiler == current.Compiler && record.Target == current.Target &&
		slices.Equal(record.Flags, current.Flags)
}

func prepareStage1Profile(source string, options Options) (Options, error) {
	if options.ProfileGenerate && options.Profile != "" {
		return options, fmt.Errorf("native: profile generation and use are mutually exclusive")
	}
	if options.profileValidated || options.Profile == "" {
		return options, nil
	}
	// Even an accidentally supplied profile cannot alter a nonshipping lane.
	if !shippedRelease(options) {
		options.Profile = ""
		return options, nil
	}
	original := options.Profile
	fallback := func(reason string) (Options, error) {
		options.Profile = ""
		fmt.Fprintf(os.Stderr, "adamic: stage 1 profile unavailable (%s); using plain ThinLTO\n", reason)
		return options, nil
	}
	data, err := ReadStage1Profile(original)
	if err != nil {
		return fallback("missing profile")
	}
	manifest, err := os.ReadFile(filepath.Join(filepath.Dir(original), "manifest.json"))
	if err != nil {
		return fallback("missing manifest")
	}
	var record Stage1ProfileManifest
	if err = json.Unmarshal(manifest, &record); err != nil {
		return fallback("invalid manifest")
	}
	current, err := newStage1ProfileManifest(source, data, record.Training, options)
	if err != nil {
		return options, err
	}
	if !profileMatches(record, current) {
		return fallback("source, runtime, profile or toolchain changed")
	}
	// Hash-addressed immutable indexed copies make runtime.a's existing flag key sound.
	profdata, err := profileDataTool()
	if err != nil {
		return fallback("llvm-profdata missing")
	}
	version, err := toolchain.Command(profdata, "--version").CombinedOutput()
	if err != nil {
		return fallback("llvm-profdata unavailable")
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return options, err
	}
	directory := filepath.Join(cache, "adamic", "profiles", profileHash(append(append([]byte{}, data...), version...)))
	if err = os.MkdirAll(directory, 0o755); err != nil {
		return options, err
	}
	indexed := filepath.Join(directory, "profile.profdata")
	if _, err = os.Stat(indexed); err != nil {
		temporary, err := os.MkdirTemp(directory, ".convert-")
		if err != nil {
			return options, err
		}
		defer os.RemoveAll(temporary)
		// Convert the already verified snapshot, never reread a mutable caller path.
		text := filepath.Join(temporary, "profile.txt")
		if err = os.WriteFile(text, data, 0o644); err != nil {
			return options, err
		}
		output := filepath.Join(temporary, "profile.profdata")
		if log, err := toolchain.Command(profdata, "merge", text, "-o", output).CombinedOutput(); err != nil {
			_ = log
			return fallback("text profile unsupported by llvm-profdata")
		}
		if err = os.Rename(output, indexed); err != nil {
			return options, err
		}
	}
	options.Profile = indexed
	options.profileValidated = true
	return options, nil
}

func profileDataTool() (string, error) {
	compiler, err := toolchain.Path("clang")
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(compiler); err == nil {
		compiler = resolved
	}
	candidate := filepath.Join(filepath.Dir(compiler), "llvm-profdata")
	if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
		return candidate, nil
	}
	if found, err := toolchain.Path("llvm-profdata"); err == nil {
		return found, nil
	}
	if goruntime.GOOS == "darwin" {
		if output, err := toolchain.Command("xcrun", "--find", "llvm-profdata").Output(); err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	}
	return "", fmt.Errorf("llvm-profdata is unavailable")
}

// CheckStage1ProfileManifest validates a shipped artifact without silently falling
// back. Training is the hash of the committed corpus, not a value from the record.
func CheckStage1ProfileManifest(source string, profile []byte, training string, record Stage1ProfileManifest) error {
	current, err := NewStage1ProfileManifest(source, profile, training)
	if err != nil {
		return err
	}
	if record.Machine == "" || record.Training != training || !profileMatches(record, current) {
		return fmt.Errorf("native: shipped stage 1 profile does not match its inputs")
	}
	return nil
}

// ReadStage1Profile also accepts ordered data shards. Every part and the joined
// bytes are verified before llvm-profdata consumes them. Parts cannot name paths.
func ReadStage1Profile(path string) ([]byte, error) {
	metadata, err := os.ReadFile(filepath.Join(filepath.Dir(path), "manifest.json"))
	if os.IsNotExist(err) {
		return os.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	var record Stage1ProfileManifest
	if err := json.Unmarshal(metadata, &record); err != nil {
		return nil, err
	}
	if len(record.Parts) == 0 {
		return os.ReadFile(path)
	}
	var data []byte
	for index, part := range record.Parts {
		if part.Name != fmt.Sprintf("profile-%02d.txt", index) {
			return nil, fmt.Errorf("native: invalid profile shard name")
		}
		content, err := os.ReadFile(filepath.Join(filepath.Dir(path), part.Name))
		if err != nil {
			return nil, err
		}
		if profileHash(content) != part.SHA256 {
			return nil, fmt.Errorf("native: profile shard bytes changed")
		}
		data = append(data, content...)
	}
	if profileHash(data) != record.Profile {
		return nil, fmt.Errorf("native: joined profile bytes changed")
	}
	return data, nil
}

// Stage1ProfileDataTool resolves the profile converter beside the selected clang.
func Stage1ProfileDataTool() (string, error) { return profileDataTool() }
