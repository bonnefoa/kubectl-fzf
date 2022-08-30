package util

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

func GetFromHttpServer(url string) (http.Header, []byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error on get of %s", url)
	}
	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("error retrieving resource from server: %s", resp.Status)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error reading response body")
	}
	return resp.Header, b, nil
}

func HeadFromHttpServer(url string) (http.Header, error) {
	resp, err := http.Head(url)
	if err != nil {
		return nil, errors.Wrapf(err, "error on get of %s", url)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error retrieving resource from server: %s", resp.Status)
	}
	return resp.Header, nil
}
