package httputil

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func UnmarshalResponseBody(resp *http.Response, result interface{}) error {
	r1, r2, err := DrainBody(resp.Body)
	resp.Body = r2

	if err != nil {
		return err
	}

	responseBody, err := ioutil.ReadAll(r1)
	if err != nil {
		return err
	}

	// this cheat is to bypass an error on a time.Time field with empty string ""
	body := strings.Replace(string(responseBody), `:""`, `:null`, 1)
	body = strings.Replace(body, `: ""`, `:null`, 1)
	if err := json.Unmarshal([]byte(body), result); err != nil {
		return err
	}

	return nil
}
