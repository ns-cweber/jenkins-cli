package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
)

var jenkinsJobURL string
var job string

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(-1)
}

func init() {
	jenkinsHost := os.Getenv("JENKINS_HOST_URL")
	if jenkinsHost == "" {
		die("JENKINS_HOST_URL is empty")
	}
	if !strings.HasSuffix(jenkinsHost, "/") {
		jenkinsHost += "/"
	}
	jenkinsJobURL = jenkinsHost + "job/"

	if len(os.Args) > 1 {
		job = os.Args[1]
	} else {
		job = os.Getenv("JENKINS_DEFAULT_JOB")
	}

	if job == "" {
		die("Job not specified. Pass a job name or set JENKINS_DEFAULT_JOB")
	}
}

// Jenkins represents RUNNING as an empty string; let's expand that in our
// output.
func buildStatus(result string) string {
	switch result {
	case "":
		return color.YellowString("RUNNING")
	case "SUCCESS":
		return color.GreenString("SUCCESS")
	case "FAILURE":
		return color.RedString("FAILURE")
	case "ABORTED":
		return color.MagentaString("ABORTED")
	default:
		return result
	}
}

func main() {
	auth, err := getCredentials("Password:")
	if err != nil {
		die(err)
	}

	results, err := jobBuilds(auth, job)
	if err != nil {
		die(err)
	}

	// collect the results and sort them by index
	sortable := make(byIndex, 0, 10)
	for result := range results {
		sortable = append(sortable, result)
	}
	sort.Sort(sortable)

	// print each result
	exitCode := 0
	for _, result := range sortable {
		if result.err != nil {
			fmt.Fprintln(os.Stderr, result.err)
			exitCode = -1
			continue
		}

		fmt.Println(
			// formatJenkinsTimestamp(result.build.Timestamp),
			buildStatus(result.build.Result),
			parseDesc(result.build.Description),
		)
	}

	os.Exit(exitCode)
}
