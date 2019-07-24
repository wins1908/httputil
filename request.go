package httputil

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

func UnmarshalRequestBody(req *http.Request, result interface{}) error {
	r1, r2, err := DrainBody(req.Body)
	req.Body = r2

	if err != nil {
		return err
	}

	responseBody, err := ioutil.ReadAll(r1)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(responseBody, result); err != nil {
		return err
	}

	return nil
}

func CopyRequest(r *http.Request) (*http.Request, error) {
	r2 := new(http.Request)
	*r2 = *r

	if r.URL != nil {
		r2URL := new(url.URL)
		*r2URL = *r.URL
		r2.URL = r2URL
	}

	if r.Body != nil {
		var err error
		r.Body, r2.Body, err = DrainBody(r.Body)
		if err != nil {
			return nil, err
		}
	}
	return r2, nil
}
