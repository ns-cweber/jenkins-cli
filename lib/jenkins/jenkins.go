package jenkins

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ns-cweber/jenkins-cli/lib/auth"
)

// Fetches the body from `url` using `auth`. Returns any errors encountered
// including non-200 responses.
func get(url string, auth auth.Credentials) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	path := strings.TrimRight(req.URL.Path, "/")
	if strings.HasSuffix(path, "/api/json") {
		path = path[:len(path)-len("/api/json")]
	}

	if strings.HasSuffix(path, "quill3_build_deploy") {
		return os.Open("/Users/cweber/Downloads/quill3_build_deploy.json")
	}

	return os.Open(filepath.Join("/tmp/foodir", path))

	//req.SetBasicAuth(auth.Username, auth.Password)
	//rsp, err := http.DefaultClient.Do(req)
	//if err != nil {
	//	return nil, err
	//}

	//if rsp.StatusCode != http.StatusOK {
	//	return nil, fmt.Errorf(
	//		"Bad status code for GET '%s': wanted 200, got %d",
	//		url,
	//		rsp.StatusCode,
	//	)
	//}

	//path := strings.TrimRight(req.URL.Path, "/")
	//if strings.HasSuffix(path, "/api/json") {
	//	path = path[:len(path)-len("/api/json")]
	//}
	//dir := filepath.Dir(path)
	//if err := os.MkdirAll(filepath.Join("/tmp/foodir", dir), 0777); err != nil {
	//	log.Println("ERR:", err)
	//	return rsp.Body, nil
	//}

	//file, err := os.Create(filepath.Join("/tmp/foodir", path))
	//if err != nil {
	//	log.Println("ERR:", err)
	//	return rsp.Body, nil
	//}
	//defer rsp.Body.Close()

	//if _, err := io.Copy(file, rsp.Body); err != nil {
	//	return nil, err
	//}

	//return file, nil
}

type CauseClass string

const CauseClassUpstream CauseClass = "hudson.model.Cause$UpstreamCause"
const CauseClassUserID CauseClass = "hudson.model.Cause$UserIdCause"

// `CauseNaginator` seems to be a retry plugin; `Causes` of this class seem to
// have a single `"shortDescription"` member that looks like this: `"Started by
// Naginator after build #1971 failure"`.
const CauseNaginator CauseClass = "com.chikli.hudson.plugin.naginator.NaginatorCause"

// `Cause` contains the cause data for a particular `hudson.model.CauseAction`.
type Cause struct {
	// `Class` identifies the Java class of the cause. This tells us which
	// fields are relevant for a given cause.
	Class CauseClass `json:"_class"`

	// `ShortDescription` is used by `hudson.model.Cause$UpstreamCause`.
	ShortDescription string `json:"shortDescription"`

	// `UpstreamBuild` is used by `hudson.model.Cause$UpstreamCause`.
	UpstreamBuild int `json:"upstreamBuild"`

	// `UpstreamProject` is used by `hudson.model.Cause$UpstreamCause`.
	UpstreamProject string `json:"upstreamProject"`

	// `UpstreamURL` contains the path part of the upstream URL (the scheme and
	// host information are excluded). For example, `/job/{project}`.  It is
	// used by `hudson.model.Cause$UpstreamCause`.
	UpstreamURL string `json:"upstreamUrl"`
}

type ActionClass string

const ActionClassParameters ActionClass = "hudson.model.ParametersAction"
const ActionClassCause ActionClass = "hudson.model.CauseAction"

// `Parameters` is a collection of key/value pairs.
type Parameters []struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// `Get` retrieves the value for the parameter named `name` or an empty string
// if `name` is not found.
func (ps Parameters) Get(name string) string {
	for _, p := range ps {
		if p.Name == name {
			return p.Value
		}
	}
	return ""
}

// `Action` contains information describing an action related to a `Build`.
type Action struct {
	// `Class` identifies the Java class of the action. This tells us which
	// fields are relevant for a given action.
	Class ActionClass `json:"_class"`

	// `Parameters` contains the parameters for an action of class
	// `hudson.model.ParametersAction`
	Parameters Parameters `json:"parameters"`

	// `Causes` contains the cause data for this action. It is used by class
	// `hudson.model.CauseAction`.
	Causes []Cause
}

type BuildResult string

const BuildResultAborted BuildResult = "ABORTED"
const BuildResultFailure BuildResult = "FAILURE"
const BuildResultPending BuildResult = ""
const BuildResultSuccess BuildResult = "SUCCESS"

// A `build` represents a Jenkins build
type Build struct {
	Number      string      `json:"id"`
	Description string      `json:"description"`
	Result      BuildResult `json:"result"`
	BuiltOn     string      `json:"builtOn"`
	Actions     []Action    `json:"actions"`

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
			for {
				var result Result
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
