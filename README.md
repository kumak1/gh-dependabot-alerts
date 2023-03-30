# gh-dependabot-alerts
show dependabot alerts list.

## Install

```shell
gh extension install kumak1/gh-dependabot-alerts
```

## Usage

```shell
% gh dependabot-alerts -h
  -e, --ecosystem string   specify comma-separated list. can be: composer, go, maven, npm, nuget, pip, pub, rubygems, rust
  -h, --help               help
      --hostname string    specify github hostname
  -q, --jq string          Query to select values from the response using jq syntax
  -o, --owner string       specify github owner
      --per_page int       The number of results per page (max 100). (default 30)
      --quiet              show only github api response
  -r, --repo stringArray   specify github repository name
      --scope string       specify comma-separated list. can be: development, runtime
      --severity string    specify comma-separated list. can be: low, medium, high, critical
      --state string       specify comma-separated list. can be: dismissed, fixed, open
```