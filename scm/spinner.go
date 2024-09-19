package scm

import (
	"time"

	"github.com/briandowns/spinner"
)

var (
	spinningSpinner *spinner.Spinner
)

func init() {
	spinningSpinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
}
