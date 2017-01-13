package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ns-cweber/jenkins-cli"
	"github.com/ns-cweber/jenkins-cli/auth"
)

const timeFmt = "2006-01-02T15:04"

var jenkinsHost string
var job string

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(-1)
}

func init() {
	jenkinsHost = os.Getenv("JENKINS_HOST_URL")
	if jenkinsHost == "" {
		die("JENKINS_HOST_URL is empty")
	}
	if !strings.HasSuffix(jenkinsHost, "/") {
		jenkinsHost += "/"
	}

	if len(os.Args) > 1 {
		job = os.Args[1]
	} else {
		job = os.Getenv("JENKINS_DEFAULT_JOB")
	}

	if job == "" {
		die("Job not specified. Pass a job name or set JENKINS_DEFAULT_JOB")
	}
}

// Parses a Jenkins timestamp (an integer number of milliseconds since the unix
// epoch) into the format YYYY-MM-DDThh:mm. Time is represented on a 24 hour
// clock with zero-padded hours.
func formatJenkinsTimestamp(timestamp int64) string {
	return time.Unix(timestamp/1000, timestamp%1000*1000).Format(timeFmt)
}

// Parses the relevant data out of `desc` (Jenkins encodes the descriptions
// with the format: `<a title="{desc}", href="{href}">{buildNumber}: </a>
// {desc}`
func parseDesc(desc string) string {
	const prefix string = "<a title=\""
	if !strings.HasPrefix(desc, prefix) {
		return desc
	}
	if end := strings.Index(desc, "\" href=\""); end > -1 {
		return desc[len(prefix):end]
	}
	return desc
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
	auth, err := auth.GetCredentials("Password:")
	if err != nil {
		die(err)
	}

	client := jenkins.Client{auth, jenkinsHost}
	results, err := client.JobBuilds(job)
	if err != nil {
		die(err)
	}

	// print each result
	exitCode := 0
	for result := range results {
		if result.Err != nil {
			fmt.Fprintln(os.Stderr, result.Err)
			exitCode = -1
			continue
		}

		fmt.Println(
			// The timestamp is really ugly, but it might be useful
			// formatJenkinsTimestamp(result.Build.Timestamp),
			buildStatus(result.Build.Result),
			parseDesc(result.Build.Description),
		)
	}

	os.Exit(exitCode)
}
