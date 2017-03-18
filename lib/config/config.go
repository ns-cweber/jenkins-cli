package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(-1)
}

const JobQuill3PullRequest = "quill3_pull_request"
const JobQuill3BuildDeploy = "quill3_build_deploy"

const EnvDefaultJob = "JENKINS_DEFAULT_JOB"
const EnvHostURL = "JENKINS_HOST_URL"

// `Host` returns the value in `$JENKINS_HOST_URL`. If the value is not empty,
// this function will make sure a single trailing slash is suffixed.
func Host() string {
	if host := os.Getenv(EnvHostURL); host != "" {
		return strings.TrimRight(host, "/") + "/"
	}
	return ""
}

// `MustHost` will return the value from `$JENKINS_HOST_URL` (with a guaranteed
// trailing slash) or die trying.
func MustHost() string {
	return strings.TrimRight(MustEnv(EnvHostURL), "/") + "/"
}

// `DefaultJob` returns the value in `$JENKINS_DEFAULT_JOB`.
func DefaultJob() string {
	return os.Getenv(EnvDefaultJob)
}

// `MustDefaultJob` prints the value in `$JENKINS_DEFAULT_JOB` or dies if the
// variable is empty.
func MustDefaultJob() string {
	return MustEnv(EnvDefaultJob)
}

// `MustEnv` will return the value for the environment variable `key` or die
// trying.
func MustEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	die("Key not set:", key)
	return "" // will not reach
}

const timeFmt = "2006-01-02 15:04"

// `FormatTimestamp` parses a jenkins timestamp (an integer number of
// milliseconds since the unix epoch) into the format "yyyy-mm-dd hh:mm". Time
// is represented on a 24 hour clock with zero-padded hours.
func FormatTimestamp(timestamp int64) string {
	return time.Unix(timestamp/1000, timestamp%1000*1000).Format(timeFmt)
}
