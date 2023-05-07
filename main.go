package main

import (
	"git.pepabo.com/kumak1/gh-dependabot-alerts/internal"
)

func main() {
	results := internal.GhIndex(internal.Hostname, internal.OwnerName, internal.RepoNames)

	// 実行結果の出力順序を、オプションの指定順に固定する
	for _, repoName := range internal.RepoNames {
		results[repoName].Print()
	}
}
