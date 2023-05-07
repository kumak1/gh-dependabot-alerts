package main

import (
	"bytes"
	"fmt"
	"git.pepabo.com/kumak1/gh-dependabot-alerts/internal"
	"github.com/cli/go-gh"
	"github.com/fatih/color"
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

func main() {
	results := make(map[string]ghResult, len(internal.RepoNames))

	var wg sync.WaitGroup
	for _, repoName := range internal.RepoNames {
		wg.Add(1)
		repoName := repoName
		go func() {
			results[repoName] = ghExec(internal.Hostname, internal.OwnerName, repoName)
			wg.Done()
		}()
	}
	wg.Wait()

	// 実行結果の出力順序を、オプションの指定順に固定する
	for _, repoName := range internal.RepoNames {
		results[repoName].print()
	}
}

func ghExec(hostname string, ownerName string, repoName string) ghResult {
	stdOut, stdErr, err := gh.Exec(internal.RequestArgs(hostname, ownerName, repoName)...)
	if err != nil {
		log.Fatal(err)
	}
	return ghResult{repoName: repoName, stdOut: stdOut, stdErr: stdErr}
}

func (r ghResult) print() {
	if !internal.OutputQuiet {
		fmt.Println(r.repoName)
	}

	if stdOutString := r.stdOut.String(); stdOutString != "" {
		if internal.OutputDefault {
			filterTime, enableDateFilter := filterTime()
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 4, 8, 1, '\t', 0)
			for _, columns := range strings.Split(stdOutString, "\n") {
				var cols = strings.Split(columns, "\t")
				if cols[0] == "" {
					break
				}

				columnTime, _ := time.Parse(time.RFC3339, cols[5])
				if enableDateFilter && columnTime.Before(filterTime) {
					break
				}

				cols[0] = formatIndex(cols[0])
				cols[1] = formatSeverity(cols[1])
				cols[5] = formatDate(columnTime)
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

func filterTime() (time.Time, bool) {
	if internal.OutputSinceWeek == 0 {
		return time.Now(), false
	}

	return time.Now().AddDate(0, 0, -7*internal.OutputSinceWeek), true
}

func formatIndex(index string) string {
	return color.GreenString("#" + index)
}

func formatDate(t time.Time) string {
	return color.WhiteString(t.Format("2006-01-02"))
}

func formatSeverity(severity string) string {
	switch severity {
	case "low":
		return color.WhiteString(severity)
	case "medium":
		return color.YellowString(severity)
	case "high":
		return color.RedString(severity)
	case "critical":
		return color.HiRedString(severity)
	default:
		return severity
	}
}
