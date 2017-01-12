JENKINS
-------

`jenkins` is a CLI for querying a Jenkins job for build status and description.

## Installation

`go get github.com/ns-cweber/jenkins`
`echo "export JENKINS_HOST_URL={host_url}" >> ~/.bash_profile`

## Usage

`jenkins [<job_name>]`

### Environment Variables

- `JENKINS_HOST_URL` (required): The Jenkins host address
- `JENKINS_DEFAULT_JOB` (optional): The default job to query for
