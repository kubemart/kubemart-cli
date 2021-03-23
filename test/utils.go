package test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kubemart/kubemart/pkg/utils"
)

// HasNamespaceGone will wait until the namespace gets completely deleted
// and return 'true' when it completes. Otherwise it returns 'false' if
// the namespace still exists after maximum waiting period (maxRetries).
func HasNamespaceGone(namespace string) bool {
	canProceed := false
	current := 0
	maxRetries := 40

	for {
		current++
		if current > maxRetries {
			break
		}

		exists, _ := utils.IsNamespaceExist(namespace)
		if !exists {
			canProceed = true
			break
		}

		fmt.Printf("Waiting for %s namespace to get terminated...\n", namespace)
		time.Sleep(3 * time.Second)
	}

	return canProceed
}

// RecordStdOutStdErr will execute callerFunc (that contains commands to execute),
// wait until the execution ends and return its stdout and stderr
func RecordStdOutStdErr(callerFunc func()) (string, string) {
	stdOut := os.Stdout
	stdErr := os.Stderr

	outPipeReader, outPipeWriter, _ := os.Pipe()
	errPipeReader, errPipeWriter, _ := os.Pipe()
	defer outPipeReader.Close()
	defer errPipeReader.Close()

	os.Stdout = outPipeWriter
	os.Stderr = errPipeWriter

	// execute the caller function
	callerFunc()

	outPipeWriter.Close()
	errPipeWriter.Close()

	os.Stdout = stdOut
	os.Stderr = stdErr

	out := new(strings.Builder)
	io.Copy(out, outPipeReader)

	err := new(strings.Builder)
	io.Copy(err, errPipeReader)

	outStr := out.String()
	errStr := err.String()

	return outStr, errStr
}
