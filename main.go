package main

import (
	"fmt"
	"github.com/cli/go-gh"
	flags "github.com/spf13/pflag"
	"log"
	"net/url"
	"os"
)

func main() {
	repo, err := gh.CurrentRepository()
	if err != nil {
		fmt.Println(err)
	}

	initArguments()

	// Queryを追加
	u := &url.URL{}
	u.Path = fmt.Sprintf("/repos/%s/%s/dependabot/alerts", repo.Owner(), repo.Name())
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

	u.RawQuery = q.Encode()

	args := []string{"api", "--hostname", repo.Host(), "--jq", jq, u.String()}
	stdOut, stdErr, err := gh.Exec(args...)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stdOut.String())
	fmt.Println(stdErr.String())
}

var ecosystem string
var scope string
var severity string
var state string
var jq string

func initArguments() {
	flags.StringVarP(&ecosystem, "ecosystem", "e", "", "specify comma-separated list. can be: composer, go, maven, npm, nuget, pip, pub, rubygems, rust")
	flags.StringVar(&scope, "scope", "", "specify comma-separated list. can be: development, runtime")
	flags.StringVar(&severity, "severity", "", "specify comma-separated list. can be: low, medium, high, critical")
	flags.StringVar(&state, "state", "", "specify comma-separated list. can be: dismissed, fixed, open")
	flags.StringVarP(&jq, "jq", "q", ".[] | [.created_at, .security_advisory.severity, .dependency.package.name, .html_url] | @csv", "Query to select values from the response using jq syntax")

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
