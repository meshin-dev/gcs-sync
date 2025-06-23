package gsutil

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// RSync performs a recursive synchronization between a source and destination using gsutil.
// It wraps the `gsutil rsync -r` command with additional options for parallel execution
// and the ability to ignore specific patterns.
//
// Parameters:
//   - src: The source path or URL to synchronize from.
//   - dst: The destination path or URL to synchronize to.
//   - deleteRemote: If true, deletes files in the destination that are not present in the source.
//   - ignoreRegex: A slice of regular expressions used to exclude files from synchronization.
//   - log: A logrus.Entry for logging the operation's progress and any errors.
//
// The function does not return any value, but logs the operation's progress and any errors encountered.
func RSync(src, dst string, deleteRemote bool, ignoreRegex []*regexp.Regexp, log *logrus.Entry) {
	args := []string{
		"-m", // parallel
		"-o", "GSUtil:parallel_process_count=1",
		"-o", "GSUtil:sliced_object_download_threshold=0",
		"rsync", "-r",
		"-e", // ⇐  skip symlinks that point outside the tree / are broken
	}
	if deleteRemote {
		args = append(args, "-d")
	}
	for _, re := range ignoreRegex {
		args = append(args, "-x", re.String())
	}
	args = append(args, src, dst)

	log.Infof("gsutil %s", strings.Join(args, " "))

	cmd := exec.Command("gsutil", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	start := time.Now()
	if err := cmd.Run(); err != nil {
		log.WithError(err).Error("gsutil exited with error")
	}
	log.Infof("gsutil finished in %s", time.Since(start).Round(time.Millisecond))
}
