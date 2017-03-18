JENKINS
-------

`jenkins` is a CLI for querying recent builds from a Jenkins job. It supports
a simple querying syntax of `key=value` or `key!=value` pairs that can be
composed into more complex queries using `&` and `|` operators. For example,
`status!=SUCCESS&GIT_REF=master`. In this example, GIT_REF is a custom build
parameter; `jenkins` understands build parameters, and allows you to query on
them as well as first-class fields like `status`.

This tool will prompt you for your Jenkins credentials; after the first use,
these credentials will be cached on your system keyring.

![screenshot](screenshot.png)

## Installation

1. `go get github.com/ns-cweber/jenkins-cli/cmd/jenkins`

## Usage

`jenkins <job_name> [<query>]`
