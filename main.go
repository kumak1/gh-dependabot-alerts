package main

import (
	"bytes"
	"fmt"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/fatih/color"
	flags "github.com/spf13/pflag"
	"log"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

var (
	repo         repository.Repository
	args         []string
	repositories []string
	hostname     string
	owner        string
	ecosystem    string
	scope        string
	severity     string
	state        string
	perPage      int
	jq           string
	quiet        bool
	sinceWeek    int
)

type response struct {
	name   string
	stdOut bytes.Buffer
	stdErr bytes.Buffer
}

func main() {
	initArguments()

	repos := targetRepos()

	for _, repoName := range repos {
		res := ghExec(repoName)
		res.print()
	}
}

func ghExec(repoName string) response {
	stdOut, stdErr, err := gh.Exec(append(args, []string{targetPath(repoName)}...)...)
	if err != nil {
		log.Fatal(err)
	}
	return response{name: repoName, stdOut: stdOut, stdErr: stdErr}
}

func (r response) print() {
	if !quiet {
		fmt.Println(r.name)
	}

	stdOutString := r.stdOut.String()
	if stdOutString == "" {
		return
	}

	if jq != "" {
		fmt.Print(stdOutString)
		return
	}

	filterTime, enableDateFilter := filterTime()

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, '\t', 0)
	for _, columns := range strings.Split(stdOutString, "\n") {
		var cols = strings.Split(columns, "\t")
		if cols[0] == "" {
			break
		}

		columnTime := parseResponseDate(cols[5])
		if enableDateFilter && columnTime.Before(filterTime) {
			break
		}

		cols[0] = formatIndex(cols[0])
		cols[1] = formatSeverity(cols[1])
		cols[5] = formatDate(columnTime)
		_, _ = fmt.Fprintln(w, strings.Join(cols, "\t"))
	}
	_ = w.Flush()

	stdErrString := r.stdErr.String()
	if stdErrString != "" {
		fmt.Print(stdErrString)
	}
}

func targetRepos() []string {
	if len(repositories) == 0 && repo != nil {
		return []string{repo.Name()}
	} else {
		return repositories
	}
}

func targetHostname() string {
	if hostname == "" && repo != nil {
		return repo.Host()
	} else {
		return hostname
	}
}

func targetOwner() string {
	if owner == "" && repo != nil {
		return repo.Owner()
	} else {
		return owner
	}
}

func filterTime() (time.Time, bool) {
	if sinceWeek == 0 {
		return time.Now(), false
	}

	return time.Now().AddDate(0, 0, -7*sinceWeek), true
}

func targetPath(repoName string) string {
	u := &url.URL{}
	u.Path = fmt.Sprintf("/repos/%s/%s/dependabot/alerts", targetOwner(), repoName)
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

func parseResponseDate(date string) time.Time {
	t, _ := time.Parse(time.RFC3339, date)
	return t
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

func initArguments() {
	flags.StringArrayVarP(&repositories, "repo", "r", []string{}, "specify github repository name")
	flags.StringVar(&hostname, "hostname", "", "specify github hostname")
	flags.StringVarP(&owner, "owner", "o", "", "specify github owner")
	flags.StringVarP(&ecosystem, "ecosystem", "e", "", "specify comma-separated list. can be: composer, go, maven, npm, nuget, pip, pub, rubygems, rust")
	flags.StringVar(&scope, "scope", "", "specify comma-separated list. can be: development, runtime")
	flags.StringVar(&severity, "severity", "", "specify comma-separated list. can be: low, medium, high, critical")
	flags.StringVar(&state, "state", "", "specify comma-separated list. can be: dismissed, fixed, open")
	flags.IntVar(&perPage, "per_page", 30, "The number of results per page (max 100).")
	flags.StringVarP(&jq, "jq", "q", "", "Query to select values from the response using jq syntax")
	flags.BoolVar(&quiet, "quiet", false, "show only github api response")
	flags.IntVar(&sinceWeek, "since_week", 0, "specified number of weeks. Valid only if --jq is not specified.")

	var help bool
	flags.BoolVarP(&help, "help", "h", false, "help")
	flags.Parse()

	if help {
		flags.PrintDefaults()
		os.Exit(1)
	}

	if hostname == "" || owner == "" {
		currentRepository, err := gh.CurrentRepository()
		if err != nil {
			fmt.Println(err)
		}
		repo = currentRepository
	}

	args = []string{"api", "--hostname", targetHostname(), "--jq", outputQuery()}
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
