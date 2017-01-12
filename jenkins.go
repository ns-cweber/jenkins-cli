package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

const timeFmt = "2006-01-02T15:04"

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

// Fetches the body from `url` using `auth`. Returns any errors encountered
// including non-200 responses.
func get(url string, auth credentials) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(auth.username, auth.password)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"Bad status code for GET '%s': wanted 200, got %d",
			url,
			rsp.StatusCode,
		)
	}

	return rsp.Body, nil
}

// Fetches a JSON payload from `url` using `auth` and decodes it into `v`.
// Returns any errors encountered including non-200 responses.
func httpDecode(url string, auth credentials, v interface{}) error {
	body, err := get(url, auth)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// A `build` epresents a Jenkins build
type build struct {
	Number      string `json:"id"`
	Description string `json:"description"`
	Result      string `json:"result"`
	BuiltOn     string `json:"builtOn"`

	// milliseconds from unix epoch
	Timestamp int64 `json:"timestamp"`
}

// A tuple of `index`, `build`, and `err` for channel convenience
type result struct {
	index int
	build build
	err   error
}

// Gets each url in `urls` using `auth` and sends the result into the returned
// channel. This function uses a number of worker goroutines to get the URLs
// in parallel.
func getAsync(auth credentials, urls []string) <-chan result {
	var lock sync.Mutex
	var cursor int
	var results = make(chan result, 100)

	// dispatch 8 worker goroutines to grab the build info
	for i := 0; i < 8; i++ {
		go func() {
			var result result
			for {
				// grab the next index
				lock.Lock()
				result.index = cursor
				cursor++
				lock.Unlock()

				// if there are no more URLs, this worker should quit
				if result.index >= len(urls) {
					return
				}

				// otherwise we should get the build and send it into the
				// channel
				result.err = httpDecode(urls[result.index], auth, &result.build)
				results <- result

				// if this index was the last, we should close the channel and
				// quit
				if result.index == len(urls)-1 {
					close(results)
					return
				}
			}
		}()
	}

	return results
}

// Gets the builds in `job` using `auth`. May return an error if there was an
// error getting the job information. Errors accessing each build will be
// stored in the results in the returned channel.
func jobBuilds(auth credentials, job string) (<-chan result, error) {
	url := jenkinsJobURL + job + "/api/json"
	var result struct {
		Builds []struct {
			URL string `json:"url"`
		} `json:"builds"`
	}
	if err := httpDecode(url, auth, &result); err != nil {
		return nil, err
	}

	// grab the URLs from the headers
	urls := make([]string, len(result.Builds))
	for i, header := range result.Builds {
		urls[i] = header.URL + "api/json"
	}

	return getAsync(auth, urls), nil
}
