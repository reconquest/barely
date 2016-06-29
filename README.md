# barely ![report](https://goreportcard.com/badge/github.com/reconquest/barely)

Dead simple but yet extensible status bar for displaying interactive progress
for the shell-based tools, written in Go-lang.

![barely-example](https://cloud.githubusercontent.com/assets/674812/16452342/c0ef1d74-3e29-11e6-83c2-a8960c3218ea.gif)

# Example

```go
package main

import (
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/reconquest/barely"
	"github.com/reconquest/loreley"
)

func main() {
	format, err := loreley.CompileWithReset(
		` {bg 2}{fg 15}{bold} {.Mode} `+
			`{bg 253}{fg 0} `+
			`{if .Updated}{fg 70}{end}{.Done}{fg 0}/{.Total} `,
		nil,
	)
	if err != nil {
		panic(err)
	}

	var (
		bar = barely.NewStatusBar(format.Template)
		wg  = sync.WaitGroup{}

		status = &struct {
			Mode  string
			Total int
			Done  int64

			Updated int64
		}{
			Mode:  "PROCESSING",
			Total: 100,
		}
	)

	bar.SetStatus(status)
	bar.Render(os.Stderr)

	for i := 1; i <= status.Total; i++ {
		wg.Add(1)

		go func(i int) {
			time.Sleep(
				time.Duration(rand.Intn(i)) * time.Millisecond * 300,
			)

			atomic.AddInt64(&status.Done, 1)
			atomic.AddInt64(&status.Updated, 1)
			bar.Render(os.Stderr)

			wg.Done()

			<-time.After(time.Millisecond * 500)
			atomic.AddInt64(&status.Updated, -1)
			bar.Render(os.Stderr)
		}(i)
	}

	wg.Wait()

	bar.Clear(os.Stderr)
}
```
