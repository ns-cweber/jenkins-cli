package jenkins

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/ns-cweber/jenkins-cli/auth"
)

// Fetches the body from `url` using `auth`. Returns any errors encountered
// including non-200 responses.
func get(url string, auth auth.Credentials) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(auth.Username, auth.Password)
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

// A `build` epresents a Jenkins build
type Build struct {
	Number      string `json:"id"`
	Description string `json:"description"`
	Result      string `json:"result"`
	BuiltOn     string `json:"builtOn"`

	// milliseconds from unix epoch
	Timestamp int64 `json:"timestamp"`
}

// A tuple of `Index`, `Build`, and `Err` for channel convenience
type Result struct {
	Index int
	Build Build
	Err   error
}

type Client struct {
	Auth auth.Credentials

	// The URL of the Jenkins host. This should have a trailing slash.
	HostURL string
}

// Fetches a JSON payload from `url` and decodes it into `v`. Returns any
// errors encountered including non-200 responses.
func (c Client) httpDecode(url string, v interface{}) error {
	body, err := get(url, c.Auth)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Gets each url in `urls` and sends the result into the returned channel. This
// function uses a number of worker goroutines to get the URLs in parallel.
func (c Client) getAsync(urls []string) <-chan Result {
	var lock sync.Mutex
	var cursor int
	var results = make(chan Result, 100)

	// dispatch 8 worker goroutines to grab the build info
	for i := 0; i < 8; i++ {
		go func() {
			var result Result
			for {
				// grab the next index
				lock.Lock()
				result.Index = cursor
				cursor++
				lock.Unlock()

				// if there are no more URLs, this worker should quit
				if result.Index >= len(urls) {
					return
				}

				// otherwise we should get the build and send it into the
				// channel
				result.Err = c.httpDecode(urls[result.Index], &result.Build)
				results <- result

				// if this index was the last, we should close the channel and
				// quit
				if result.Index == len(urls)-1 {
					close(results)
					return
				}
			}
		}()
	}

	return results
}

// Gets the builds in the job called `name`. May return an error if any were
// encountered while getting the job information. Errors accessing each build
// will be stored in the corresponding result in the returned channel.
func (c Client) JobBuilds(name string) (<-chan Result, error) {
	url := c.HostURL + "job/" + name + "/api/json"
	var result struct {
		Builds []struct {
			URL string `json:"url"`
		} `json:"builds"`
	}
	if err := c.httpDecode(url, &result); err != nil {
		return nil, err
	}

	// grab the URLs from the headers
	urls := make([]string, len(result.Builds))
	for i, header := range result.Builds {
		urls[i] = header.URL + "api/json"
	}

	return c.getAsync(urls), nil
}
