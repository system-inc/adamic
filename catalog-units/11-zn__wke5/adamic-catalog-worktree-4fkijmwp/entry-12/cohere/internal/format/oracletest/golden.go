// Package oracletest keeps what the embedded Prettier answered for a test's fixtures in a checked-in golden
// file, so the test compares bytes and runs no JavaScript (#nxgt2ca).
//
// The live oracle formatted every fixture under every option set on every run, and yaml, css and markdown
// spent minutes of the module's test time in it. Its answers change only when the bundles do, so they are
// recorded once, keyed by everything an answer depends on, and stamped with the bundles' digest. A golden
// recorded from other bundles fails the test loudly, naming the command that records it again, rather
// than comparing the native printers against an oracle nothing loads any more. Recording is its own run:
//
//	COHERE_UPDATE_ORACLE=1 go test ./internal/format/yaml
//
// The checks themselves are untouched: every fixture is still compared under every option set, against
// the same bytes the oracle gives.
package oracletest

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/prettier"
)

// UpdateVariable, set to anything, records the oracle's answers anew instead of reading them.
const UpdateVariable = "COHERE_UPDATE_ORACLE"

// ErrNotRecorded is an answer asked for that the golden does not hold: the fixtures changed since it was
// recorded.
var ErrNotRecorded = errors.New("this answer was never recorded")

// Answer is one recorded answer: what the oracle printed, or the error it gave.
type Answer struct {
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// record is a golden file's contents.
type record struct {
	Digest  string            `json:"digest"`
	Answers map[string]Answer `json:"answers"`
}

// Golden is one test's recorded answers.
type Golden struct {
	t       testing.TB
	path    string
	digest  string
	update  bool
	mutex   sync.Mutex
	answers map[string]Answer
}

// Open reads the golden at testdata/oracle/<name>.json.gz for the bundles this build embeds. When recording,
// it starts empty and writes what the test asked for when the test ends.
func Open(t testing.TB, name string) *Golden {
	t.Helper()
	digest, err := BundlesDigest()
	if err != nil {
		t.Fatal(err)
	}
	golden := &Golden{
		t:       t,
		path:    filepath.Join("testdata", "oracle", name+".json.gz"),
		digest:  digest,
		update:  os.Getenv(UpdateVariable) != "",
		answers: map[string]Answer{},
	}
	if golden.update {
		t.Cleanup(golden.write)
		return golden
	}
	answers, err := read(golden.path, digest)
	if err != nil {
		t.Fatalf("%v: record them with %s=1 go test -run '^%s$' in this package", err, UpdateVariable, t.Name())
	}
	golden.answers = answers
	return golden
}

// read is a golden's answers, refused when the bundles that recorded them are not these.
func read(path string, digest string) (map[string]Answer, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%s holds no recorded oracle answers (%v)", path, err)
	}
	defer file.Close()
	unzipped, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("%s is not a golden: %v", path, err)
	}
	var contents record
	if err := json.NewDecoder(unzipped).Decode(&contents); err != nil {
		return nil, fmt.Errorf("%s is not a golden: %v", path, err)
	}
	if contents.Digest != digest {
		return nil, fmt.Errorf("%s was recorded from Prettier bundles %s, and this build embeds %s", path, contents.Digest, digest)
	}
	return contents.Answers, nil
}

// write saves what this run recorded. Keys are sorted by the encoder, so recording twice from the same
// bundles writes the same bytes.
func (golden *Golden) write() {
	if golden.t.Failed() {
		golden.t.Logf("not recording %s, since the test failed", golden.path)
		return
	}
	if err := os.MkdirAll(filepath.Dir(golden.path), 0o755); err != nil {
		golden.t.Errorf("recording %s: %v", golden.path, err)
		return
	}
	file, err := os.Create(golden.path)
	if err != nil {
		golden.t.Errorf("recording %s: %v", golden.path, err)
		return
	}
	defer file.Close()
	zipped, _ := gzip.NewWriterLevel(file, gzip.BestCompression)
	encoder := json.NewEncoder(zipped)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(record{Digest: golden.digest, Answers: golden.answers}); err != nil {
		golden.t.Errorf("recording %s: %v", golden.path, err)
		return
	}
	if err := zipped.Close(); err != nil {
		golden.t.Errorf("recording %s: %v", golden.path, err)
		return
	}
	golden.t.Logf("recorded %d oracle answers in %s", len(golden.answers), golden.path)
}

// Key is an answer's key: every input it depends on, each prefixed with its length so no two inputs can
// run together into the same key.
func Key(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		fmt.Fprintf(hash, "%d\x00%s\x00", len(part), part)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// Answer is the recorded answer for key. When recording, compute is asked, and what it says is recorded,
// error and all. An answer the golden never recorded fails the test and returns ErrNotRecorded.
func (golden *Golden) Answer(key string, compute func() (string, error)) (string, error) {
	if golden.update {
		output, err := compute()
		answer := Answer{Output: output}
		if err != nil {
			answer = Answer{Error: err.Error()}
		}
		golden.mutex.Lock()
		golden.answers[key] = answer
		golden.mutex.Unlock()
		return output, err
	}
	golden.mutex.Lock()
	answer, recorded := golden.answers[key]
	golden.mutex.Unlock()
	if !recorded {
		golden.t.Errorf("an oracle answer %s never recorded was asked for, so the fixtures changed: record them again with %s=1",
			golden.path, UpdateVariable)
		return "", ErrNotRecorded
	}
	if answer.Error != "" {
		return "", errors.New(answer.Error)
	}
	return answer.Output, nil
}

// Engine answers as prettier.New(options) would, from the golden. The real engine is built only when
// recording, and only once.
func (golden *Golden) Engine(options formatoptions.Options) *Engine {
	return &Engine{golden: golden, options: options}
}

// Engine is a prettier.Engine's Format, answered from a golden.
type Engine struct {
	golden  *Golden
	options formatoptions.Options
	once    sync.Once
	engine  *prettier.Engine
	err     error
}

// Format is the answer prettier.Engine.Format gives for fileName and text under the engine's options.
func (engine *Engine) Format(fileName string, text string) (string, error) {
	return engine.golden.Answer(Key("engine", fmt.Sprintf("%+v", engine.options), fileName, text), func() (string, error) {
		engine.once.Do(func() { engine.engine, engine.err = prettier.New(engine.options) })
		if engine.err != nil {
			return "", engine.err
		}
		return engine.engine.Format(fileName, text)
	})
}

var bundlesDigest struct {
	once   sync.Once
	digest string
	err    error
}

// BundlesDigest is the digest of the Prettier bundles this build embeds, read once.
func BundlesDigest() (string, error) {
	bundlesDigest.once.Do(func() {
		bundles, err := prettier.Bundles()
		if err != nil {
			bundlesDigest.err = err
			return
		}
		bundlesDigest.digest, bundlesDigest.err = prettier.DigestBundles(bundles.Files)
	})
	return bundlesDigest.digest, bundlesDigest.err
}
