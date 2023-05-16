package internal

import (
	"bytes"
	"fmt"
	"github.com/cli/go-gh"
	"log"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

type ghResult struct {
	repoName string
	stdOut   bytes.Buffer
	stdErr   bytes.Buffer
}

func (r ghResult) Print() {
	if !outputQuiet {
		fmt.Println(r.repoName)
	}

	if stdOutString := r.stdOut.String(); stdOutString != "" {
		if outputDefault {
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 4, 8, 1, '\t', 0)
			for _, columns := range strings.Split(stdOutString, "\n") {
				var cols = strings.Split(columns, "\t")
				if cols[0] == "" {
					break
				}

				columnTime, _ := time.Parse(time.RFC3339, cols[6])
				if enableTimeFilter && columnTime.Before(filterTime) {
					break
				}

				cols[0] = formatIndex(cols[0])
				cols[1] = formatSeverity(cols[1])
				cols[6] = formatDate(columnTime)
				_, _ = fmt.Fprintln(w, strings.Join(cols, "\t"))
			}
			_ = w.Flush()
		} else {
			fmt.Print(stdOutString)
		}
	}

	if stdErrString := r.stdErr.String(); stdErrString != "" {
		fmt.Print(stdErrString)
	}
}

func GhIndex(hostname string, ownerName string, repoNames []string) map[string]ghResult {
	results := make(map[string]ghResult, len(repoNames))

	var wg sync.WaitGroup
	for _, repoName := range repoNames {
		wg.Add(1)
		repoName := repoName
		go func() {
			defer wg.Done()
			stdOut, stdErr, err := gh.Exec(RequestArgs(hostname, ownerName, repoName)...)
			if err != nil {
				log.Fatal(err)
			}
			results[repoName] = ghResult{repoName: repoName, stdOut: stdOut, stdErr: stdErr}
		}()
	}
	wg.Wait()

	return results
}
