package watcher

import (
	"gcs_sync/internal/config"
	"gcs_sync/internal/gsutil"
	"gcs_sync/internal/ignore"
	"gcs_sync/internal/logging"
	"gcs_sync/internal/util"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

type ruleRunner struct {
	rule    config.SyncRule
	srcRoot string
	ign     []*regexp.Regexp
	log     *logrus.Entry
}

// newRuleRunner creates and initializes a new ruleRunner instance.
//
// It sets up a ruleRunner with the provided SyncRule, expanding the source path,
// compiling ignore patterns, and initializing a logger.
//
// Parameters:
//   - rule: A config.SyncRule that defines the synchronization configuration.
//
// Returns:
//   - *ruleRunner: A pointer to the newly created ruleRunner instance.
//   - error: An error if there was a problem compiling the ignore patterns, or nil if successful.
func newRuleRunner(rule config.SyncRule) (*ruleRunner, error) {
	src := util.Expand(rule.Src)
	ign, err := ignore.Compile(src, rule.Ignore)
	if err != nil {
		return nil, err
	}
	return &ruleRunner{
		rule:    rule,
		srcRoot: src,
		ign:     ign,
		log:     logging.L().WithField("rule", src),
	}, nil
}

// run starts the file system watcher and synchronization process for a rule runner.
//
// This function sets up a file system watcher, performs an initial synchronization,
// and then enters a loop to handle file system events and periodic synchronizations.
// It uses a debounce mechanism to avoid excessive synchronizations during rapid file changes.
//
// Parameters:
//   - stop: A receive-only channel of struct{} used to signal when the watcher should stop.
//     When a value is received on this channel, the function will terminate its execution.
//
// Returns:
//   - error: An error if there was a problem setting up or running the watcher,
//     or nil if the watcher was stopped normally via the stop channel.
func (rr *ruleRunner) run(stop <-chan struct{}) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	// watch existing tree
	if err := addRecursive(w, rr.srcRoot); err != nil {
		return err
	}

	// initial sync
	rr.syncOnce("initial")

	// ───────────────────── debounce state ────────────────────────
	var mu sync.Mutex
	var timer *time.Timer
	resetDebounce := func(reason string) {
		mu.Lock()
		defer mu.Unlock()
		if timer == nil {
			timer = time.AfterFunc(rr.rule.DebounceWindow, func() { rr.syncOnce("debounce") })
			rr.log.Debugf("debounce timer started (%s) reason=%s", rr.rule.DebounceWindow, reason)
		} else {
			timer.Reset(rr.rule.DebounceWindow)
		}
	}

	// ───────────────────── polling ticker ────────────────────────
	var ticker *time.Ticker
	if containsDir(rr.rule.Directions, config.RemoteToLocal) || containsDir(rr.rule.Directions, config.Full) {
		ticker = time.NewTicker(rr.rule.RemotePollWindow)
		defer ticker.Stop()
		rr.log.Infof("remote polling enabled (%s)", rr.rule.RemotePollWindow)
	}

	// ───────────────────────── main loop ─────────────────────────
	for {
		select {
		case ev := <-w.Events:
			rr.handleEvent(ev, w)
			resetDebounce(ev.Op.String())

		case err := <-w.Errors:
			rr.log.WithError(err).Warn("watcher error")

		case <-tickerTick(ticker):
			rr.syncOnce("periodic pull")

		case <-stop:
			rr.log.Info("stopping watcher")
			return nil
		}
	}
}

// syncOnce performs a one-time synchronization based on the rule runner's configuration.
//
// This function synchronizes files between local and remote locations according to
// the specified sync directions in the rule. It handles both local-to-remote and
// remote-to-local synchronizations, using the gsutil.RSync function for the actual
// file transfer.
//
// Parameters:
//   - reason: A string describing the reason for this synchronization (e.g., "initial", "debounce").
//     This is used for logging purposes.
//
// The function does not return any value, but it logs the synchronization activities
// and any errors that occur during the process.
func (rr *ruleRunner) syncOnce(reason string) {
	l := rr.log.WithField("reason", reason)
	gsutil.RSync(rr.srcRoot, rr.rule.Dst, true, rr.ign, l)
}

// handleEvent processes a file system event and updates the watcher accordingly.
//
// This function is responsible for handling individual file system events. It checks
// if the event should be ignored based on the ignore patterns, logs the event,
// and adds new directories to the watcher if they are created.
//
// Parameters:
//   - ev: An fsnotify.Event representing the file system event that occurred.
//   - w: A pointer to the fsnotify.Watcher that is monitoring the file system.
//
// The function does not return any value, but it may modify the watcher's state
// by adding new directories to be watched.
func (rr *ruleRunner) handleEvent(ev fsnotify.Event, w *fsnotify.Watcher) {
	rel, _ := filepath.Rel(rr.srcRoot, ev.Name)
	rel = filepath.ToSlash(rel)

	if ignore.Match(rel, rr.ign) {
		rr.log.Debugf("ignored %s %s", ev.Op, rel)
		return
	}
	rr.log.Debugf("event %s %s", ev.Op, rel)

	// if new dir created → watch it too
	if ev.Op&fsnotify.Create != 0 {
		if fi, err := os.Stat(ev.Name); err == nil && fi.IsDir() {
			_ = addRecursive(w, ev.Name)
		}
	}
}

// containsDir checks if a given SyncDirection is present in a slice of SyncDirections.
//
// This function iterates through the provided slice and compares each element
// with the given SyncDirection value. It's used to determine if a specific
// synchronization direction is included in a set of directions.
//
// Parameters:
//   - slice: A slice of config.SyncDirection values to search through.
//   - v: The config.SyncDirection value to search for in the slice.
//
// Returns:
//   - bool: true if the SyncDirection v is found in the slice, false otherwise.
func containsDir(slice []config.SyncDirection, v config.SyncDirection) bool {
	for _, d := range slice {
		if d == v {
			return true
		}
	}
	return false
}

// addRecursive adds all directories under the specified root directory to the fsnotify watcher.
//
// This function recursively walks through the directory tree starting from the given root,
// and adds each directory to the watcher. It skips files and only adds directories.
//
// Parameters:
//   - w: A pointer to the fsnotify.Watcher to which directories will be added.
//   - root: A string representing the path to the root directory from which to start the recursive walk.
//
// Returns:
//   - error: An error if there was a problem walking the directory tree or adding a directory to the watcher,
//     or nil if all directories were successfully added.
func addRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return w.Add(p)
		}
		return nil
	})
}

// tickerTick safely selects on a ticker that may be nil.
func tickerTick(t *time.Ticker) <-chan time.Time {
	if t != nil {
		return t.C
	}
	return nil
}
