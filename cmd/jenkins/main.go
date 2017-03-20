package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/ns-cweber/cli"
	"github.com/ns-cweber/jenkins-cli/lib/auth"
	"github.com/ns-cweber/jenkins-cli/lib/boolsearch"
	"github.com/ns-cweber/jenkins-cli/lib/config"
	"github.com/ns-cweber/jenkins-cli/lib/jenkins"
	"github.com/olekukonko/tablewriter"
)

var table = tablewriter.NewWriter(os.Stdout)

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
	app := cli.Command{
		Name:        os.Args[0],
		Description: "A Jenkins querying cli",
		Arguments: []cli.Argument{{
			Name:        "JOB",
			Description: "The job to query",
			Required:    true,
		}, {
			Name:        "QUERY",
			Description: "The filter query. E.g., 'status=SUCCESS&FUNCTION=longrunning'",
			Required:    false,
			Default:     "",
		}},
		Action: func(args []string) error {
			auth, err := auth.GetCredentials("Password:")
			if err != nil {
				return fmt.Errorf("Error getting credentials: %v", err)
			}

			filter, err := boolsearch.ParseString(args[1])
			if err != nil {
				fmt.Errorf("Error parsing query: %v", err)
			}

			client := jenkins.Client{auth, config.MustHost()}
			results, err := client.JobBuilds(args[0])
			if err != nil {
				return fmt.Errorf("Error getting build info: %v", err)
			}
			table.SetHeader([]string{
				"STATUS",
				"DATE",
				"GIT REF",
				"BUILD NUM",
				"CAUSE",
			})

			for result := range results {
				var c compiler
				filter.Visit(&c)
				if !c.f(result.Build) {
					continue
				}
				var causeDesc string
			ACTIONS_LOOP:
				for _, action := range result.Build.Actions {
					if action.Class == jenkins.ActionClassCause {
						for _, cause := range action.Causes {
							// If we find a user cause, set the description and
							// stop looking; we prefer to display a user cause.
							if cause.Class == jenkins.CauseClassUserID {
								causeDesc = cause.ShortDescription
								break ACTIONS_LOOP
							}

							// Hang onto the most recent cause description in
							// case we never come across a user cause
							causeDesc = cause.ShortDescription
						}
					}
				}
				table.Append([]string{
					buildStatus(result.Build.Result),
					config.FormatTimestamp(result.Build.Timestamp),
					get(result.Build, "GIT_REF"),
					result.Build.Number,
					causeDesc,
				})
			}
			table.Render()
			return nil
		},
	}

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
