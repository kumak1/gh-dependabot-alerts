package main

import (
	"bytes"
	"fmt"
	"github.com/cli/go-gh"
	"github.com/fatih/color"
	flags "github.com/spf13/pflag"
	"log"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func main() {
	initArguments()

	repository, err := gh.CurrentRepository()
	if err != nil {
		fmt.Println(err)
	}

	args := []string{"api", "--hostname", repository.Host(), "--jq", outputQuery()}
	var owner = repository.Owner()
	var repos = targetRepos(repository.Name())
	var showRepositoryName = len(repos) > 1

	for _, repoName := range repos {
		if showRepositoryName {
			fmt.Println(repoName)
		}

		stdOut, stdErr, err := gh.Exec(append(args, []string{targetPath(owner, repoName)}...)...)
		if err != nil {
			log.Fatal(err)
		}

		printOut(stdOut)
		printError(stdErr)
	}
}

func printOut(stdOut bytes.Buffer) {
	stdOutString := stdOut.String()
	if stdOutString == "" {
		return
	}

	if jq != "" {
		fmt.Print(stdOutString)
		return
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, '\t', 0)
	for _, columns := range strings.Split(stdOutString, "\n") {
		var cols = strings.Split(columns, "\t")
		if cols[0] == "" {
			break
		}
		cols[0] = formatIndex(cols[0])
		cols[1] = formatSeverity(cols[1])
		cols[5] = formatDate(cols[5])
		_, _ = fmt.Fprintln(w, strings.Join(cols, "\t"))
	}
	_ = w.Flush()
}

func printError(stdErr bytes.Buffer) {
	stdErrString := stdErr.String()
	if stdErrString != "" {
		fmt.Print(stdErrString)
	}
}

func targetRepos(defaultRepoName string) []string {
	if len(repositories) == 0 {
		return []string{defaultRepoName}
	} else {
		return repositories
	}
}

func targetPath(owner string, repoName string) string {
	u := &url.URL{}
	u.Path = fmt.Sprintf("/repos/%s/%s/dependabot/alerts", owner, repoName)
	q := u.Query()

	if ecosystem != "" {
		q.Set("ecosystem", ecosystem)
	}
	if scope != "" {
		q.Set("scope", scope)
	}
	if severity != "" {
		q.Set("severity", severity)
	}
	if state != "" {
		q.Set("state", state)
	}
	q.Set("per_page", fmt.Sprint(perPage))

	u.RawQuery = q.Encode()

	return u.String()
}

func outputQuery() string {
	if jq != "" {
		return jq
	}

	return ".[] | [.number, .security_advisory.severity, .dependency.package.ecosystem, .dependency.package.name, .html_url, .created_at] | @tsv"
}

func formatIndex(index string) string {
	return color.GreenString("#" + index)
}

func formatDate(date string) string {
	t, _ := time.Parse(time.RFC3339, date)
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

var (
	repositories []string
	ecosystem    string
	scope        string
	severity     string
	state        string
	perPage      int
	jq           string
)

func initArguments() {
	flags.StringArrayVarP(&repositories, "repo", "r", []string{}, "")
	flags.StringVarP(&ecosystem, "ecosystem", "e", "", "specify comma-separated list. can be: composer, go, maven, npm, nuget, pip, pub, rubygems, rust")
	flags.StringVar(&scope, "scope", "", "specify comma-separated list. can be: development, runtime")
	flags.StringVar(&severity, "severity", "", "specify comma-separated list. can be: low, medium, high, critical")
	flags.StringVar(&state, "state", "", "specify comma-separated list. can be: dismissed, fixed, open")
	flags.IntVar(&perPage, "per_page", 30, "The number of results per page (max 100).")
	flags.StringVarP(&jq, "jq", "q", "", "Query to select values from the response using jq syntax")

	var help bool
	flags.BoolVarP(&help, "help", "h", false, "help")
	flags.Parse()

	if help {
		flags.PrintDefaults()
		os.Exit(1)
	}
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
