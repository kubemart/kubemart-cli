package test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kubemart/kubemart/pkg/utils"
)

// CanProceedWithInit ...
func CanProceedWithInit() bool {
	canProceed := false
	current := 0
	maxRetries := 20

	for {
		current++
		if current > maxRetries {
			break
		}

		exists, _ := utils.IsNamespaceExist("kubemart-system")
		if !exists {
			canProceed = true
			break
		}

		fmt.Println("Waiting for kubemart-system namespace to get terminated...")
		time.Sleep(3 * time.Second)
	}

	return canProceed
}

// RecordStdOutStdErr ...
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
