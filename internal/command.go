package internal

import (
	"fmt"
	"github.com/cli/go-gh"
	"github.com/spf13/pflag"
	"net/url"
	"os"
	"time"
)

var (
	RepoNames        []string
	Hostname         string
	OwnerName        string
	rawQuery         string
	outputQuery      string
	outputDefault    bool
	outputQuiet      bool
	enableTimeFilter bool
	filterTime       time.Time
)

func init() {
	currentRepository, _ := gh.CurrentRepository()

	help := pflag.BoolP("help", "h", false, "help")
	repos := pflag.StringArrayP("repo", "r", []string{}, "specify github repository repoName")
	host := pflag.String("hostname", "", "specify github hostname")
	owner := pflag.StringP("owner", "o", "", "specify github owner")
	jq := pflag.StringP("jq", "q", "", "Query to select values from the results using jq syntax")
	sinceWeek := pflag.Int("since_week", 0, "specified number of weeks. Valid only if --jq is not specified.")

	ecosystem := pflag.StringP("ecosystem", "e", "", "specify comma-separated list. can be: composer, go, maven, npm, nuget, pip, pub, rubygems, rust")
	scope := pflag.String("scope", "", "specify comma-separated list. can be: development, runtime")
	severity := pflag.String("severity", "", "specify comma-separated list. can be: low, medium, high, critical")
	state := pflag.String("state", "", "specify comma-separated list. can be: dismissed, fixed, open")
	perPage := pflag.Int("per_page", 30, "The number of results per page (max 100).")

	pflag.BoolVar(&outputQuiet, "quiet", false, "show only github api results")

	pflag.Parse()

	if *help {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	if len(*repos) == 0 && currentRepository != nil {
		RepoNames = []string{currentRepository.Name()}
	} else {
		RepoNames = *repos
	}
	if *host == "" && currentRepository != nil {
		Hostname = currentRepository.Host()
	} else {
		Hostname = *host
	}
	if *owner == "" && currentRepository != nil {
		OwnerName = currentRepository.Owner()
	} else {
		OwnerName = *owner
	}
	if *jq == "" {
		outputDefault = true
		outputQuery = ".[] | [.number, .security_advisory.severity, .dependency.package.ecosystem, .dependency.package.repoName, .html_url, .created_at] | @tsv"
	} else {
		outputQuery = *jq
	}
	if *sinceWeek > 0 {
		enableTimeFilter = true
		filterTime = time.Now().AddDate(0, 0, *sinceWeek*-7)
	}

	queryValues := url.Values{}
	if *ecosystem != "" {
		queryValues.Set("ecosystem", *ecosystem)
	}
	if *scope != "" {
		queryValues.Set("scope", *scope)
	}
	if *severity != "" {
		queryValues.Set("severity", *severity)
	}
	if *state != "" {
		queryValues.Set("state", *state)
	}
	queryValues.Set("per_page", fmt.Sprint(*perPage))
	rawQuery = queryValues.Encode()
}

func requestPath(ownerName string, repoName string) string {
	u := &url.URL{}
	u.Path = fmt.Sprintf("/repos/%s/%s/dependabot/alerts", ownerName, repoName)
	u.RawQuery = rawQuery
	return u.String()
}

func RequestArgs(hostname string, ownerName string, repoName string) []string {
	return []string{
		"api",
		"--hostname",
		hostname,
		"--jq",
		outputQuery,
		requestPath(ownerName, repoName),
	}
}
