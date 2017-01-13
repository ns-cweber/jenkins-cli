package jenkins

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
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
					wg.Done()
					return
				}

				// otherwise we should get the build and send it into the
				// channel
				result.Err = c.httpDecode(urls[result.Index], &result.Build)
				results <- result
			}
		}()
	}

	// Another goroutine will wait until all others have completed before
	// closing the channel
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func sort(in <-chan Result) <-chan Result {
	out := make(chan Result)
	go func() {
		cache := map[int]Result{}
		wanted := 0
		for {
			// first check to see if the desired value is in the cache
			if result, found := cache[wanted]; found {
				out <- result
				delete(cache, wanted)
				wanted++
				continue
			}

			// then take the next result from the inbound list
			if result, more := <-in; more {
				if result.Index == wanted {
					out <- result
					wanted++
					continue
				}
				cache[result.Index] = result
				continue
			}

			// Because we assign the indexes, we know that there are no gaps in
			// the sequence, and as such we should have no extra items in the
			// cache by the time we get here. If we still have extra items,
			// it's a programming error, and we should log it and panic.
			if len(cache) > 0 {
				for index := range cache {
					fmt.Fprintln(os.Stderr, "REMAINING:", index)
				}
				panic("Extra items in cache!")
			}

			// Then close the channel and exit the loop.
			close(out)
			break
		}
	}()
	return out
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

	return sort(c.getAsync(urls)), nil
}
