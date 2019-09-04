/*
Utils package contains utility methods for client code
*/
package incidents

import (
	"crypto/tls"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

const (
	TIMEOUT        = 2   // no of secs for url timeout
	RETRY          = 5   // no of retries for url failures
	CONTENT_LENGTH = 900 // expected response length +/- 100
)

// Initialize httpClient and request the given url
// Retry 5 times, while connecting to the server in-case of error
func GetResponse(url string) (res *http.Response, err error) {
	// InsecureSkipVerify to false for production
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   time.Second * TIMEOUT, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	retry := RETRY
	for i := 1; i <= retry; i++ {
		var getErr error
		// make the request
		res, getErr = client.Do(req)
		if getErr != nil {
			log.Warn("Attempt ", i, getErr)
			if i >= retry {
				return nil, getErr
			}
		} else {
			retry = 0
		}
	}
	return res, nil
}

// Validate the response based on response header
func ValidateResponse(res *http.Response) (err error) {
	var resLength int
	// non 200 errors
	if res.StatusCode != 200 {
		err = fmt.Errorf("Received %d status code\n", res.StatusCode)
	} else if res.Header["Content-Type"][0] != "application/json" {
		err = fmt.Errorf("Content type not spplication/json. Received => %s\n", res.Header["Content-Type"][0])
	} else {
		if len(res.Header["Content-Length"]) > 0 {
			resLength, err = strconv.Atoi(res.Header["Content-Length"][0])
			if err == nil && resLength < (CONTENT_LENGTH-100) || resLength > (CONTENT_LENGTH+100) {
				err = fmt.Errorf("content-Length mismatch 905 vs %d\n", resLength)
			}
		}
	}
	return err
}
