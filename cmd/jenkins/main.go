package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/ns-cweber/jenkins-cli"
	"github.com/ns-cweber/jenkins-cli/auth"
	"github.com/ns-cweber/jenkins-cli/config"
)

var jenkinsHost string
var job string

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(-1)
}

func init() {
	jenkinsHost = config.MustHost()

	if len(os.Args) > 1 {
		job = os.Args[1]
	} else {
		job = config.MustDefaultJob()
	}
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
func buildStatus(result jenkins.BuildResult) string {
	switch result {
	case jenkins.BuildResultPending:
		return color.YellowString("RUNNING")
	case jenkins.BuildResultSuccess:
		return color.GreenString("SUCCESS")
	case jenkins.BuildResultFailure:
		return color.RedString("FAILURE")
	case jenkins.BuildResultAborted:
		return color.MagentaString("ABORTED")
	default:
		return string(result)
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
			// config.FormatTimestamp(result.Build.Timestamp),
			buildStatus(result.Build.Result),
			parseDesc(result.Build.Description),
		)
	}

	os.Exit(exitCode)
}
