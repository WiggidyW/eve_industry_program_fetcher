package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	client = &http.Client{}
)

type EsiAuthRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func authenticate(
	clientId string,
	clientSecret string,
	refreshToken string,
) (
	accessToken string,
	expires time.Time,
	err error,
) {
	// build the request
	req, err := http.NewRequest(
		"GET",
		"https://login.eveonline.com/v2/oauth/token",
		bytes.NewBuffer([]byte(fmt.Sprintf(
			`grant_type=refresh_token&refresh_token=%s`,
			url.QueryEscape(refreshToken),
		))),
	)
	if err != nil {
		return "", time.Time{}, err
	}
	addHeaderUserAgent(req)
	addHeaderWwwContentType(req)
	addHeaderLoginHost(req)
	addHeadBasicAuth(req, clientId, clientSecret)

	// fetch the response
	httpRep, close, err := doRequest(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer close()

	var rep EsiAuthRefreshResponse
	err = json.NewDecoder(httpRep.Body).Decode(&rep)
	if err != nil {
		return "", time.Time{}, err
	}

	return rep.AccessToken, time.Now().Add(time.Duration(rep.ExpiresIn) * time.Second), nil
}

func getHead(
	url string,
	accessToken string,
) (
	pages int,
	expires time.Time,
	err error,
) {
	// build the request
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		return 0, time.Time{}, err
	}
	addHeaderUserAgent(req)
	addHeadBearerAuth(req, accessToken)

	// fetch the response
	httpRep, close, err := doRequest(req)
	if err != nil {
		return 0, time.Time{}, err
	}
	defer close()

	// parse the response headers
	expires, err = parseHeadExpires(httpRep)
	if err != nil {
		return 0, time.Time{}, err
	}
	pages, err = parseHeadPages(httpRep)
	if err != nil {
		return 0, time.Time{}, err
	}

	return pages, expires, nil
}

func getPage[M any](
	url string,
	accessToken string,
	model *M,
) (
	expires time.Time,
	err error,
) {
	// build the request
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		return time.Time{}, err
	}
	addHeadJsonContentType(req)
	addHeadBearerAuth(req, accessToken)

	// fetch the response
	httpRep, close, err := doRequest(req)
	if err != nil {
		return time.Time{}, err
	}
	defer close()

	// parse the response headers
	expires, err = parseHeadExpires(httpRep)
	if err != nil {
		return time.Time{}, err
	}

	// decode the body
	err = json.NewDecoder(httpRep.Body).Decode(model)
	if err != nil {
		return time.Time{}, err
	}

	return expires, nil
}

const (
	// only used for getPages
	numRetries          = 3
	sleepBetweenRetries = 5 * time.Second
)

func getPages[M any](
	url string,
	accessToken string,
	newModel func() *M,
) (
	chnRecv <-chan PageResult[M],
	pages int,
	expires time.Time,
	err error,
) {
	// get the head
	pages, expires, err = getHead(url, accessToken)
	if err != nil {
		return nil, 0, time.Time{}, err
	}

	// create the channel
	chn := make(chan PageResult[M], pages)

	// fetch the pages
	for i := 1; i <= pages; i++ {
		go func(i int) {
			model := newModel()
			var err error
			for j := 0; j <= numRetries; j++ {
				pageUrl := fmt.Sprintf("%s?page=%d", url, i)
				expires, err := getPage(pageUrl, accessToken, model)
				if err == nil {
					chn <- PageResult[M]{Model: *model, Expires: expires}
					return
				}
				time.Sleep(sleepBetweenRetries)
			}
			chn <- PageResult[M]{Err: err}
		}(i)
	}

	return chn, pages, expires, nil
}

type PageResult[M any] struct {
	Model   M
	Expires time.Time
	Err     error
}

func addHeaderUserAgent(
	req *http.Request,
) {
	req.Header.Add("User-Agent", "eve_industry_program_fetcher")
}

func addHeaderWwwContentType(req *http.Request) {
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
}

func addHeadJsonContentType(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
}

func addHeaderLoginHost(req *http.Request) {
	req.Header.Add("Host", "login.eveonline.com")
}

func addHeadBasicAuth(req *http.Request, clientId string, clientSecret string) {
	basic_auth := base64.StdEncoding.EncodeToString([]byte(
		fmt.Sprintf("%s:%s", clientId, clientSecret),
	))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", basic_auth))
}

func addHeadBearerAuth(req *http.Request, accessToken string) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
}

func doRequest(
	req *http.Request,
) (
	rep *http.Response,
	close func() error,
	err error,
) {
	rep, err = client.Do(req)
	if err != nil {
		close = voidClose
	} else {
		close = rep.Body.Close
		if rep.StatusCode != http.StatusOK {
			err = fmt.Errorf("http status code: %d", rep.StatusCode)
		}
	}
	return rep, close, err
}

func voidClose() error { return nil }

func parseHeadExpires(rep *http.Response) (
	expires time.Time,
	err error,
) {
	datestring := rep.Header.Get("Expires")
	if datestring == "" {
		return time.Time{}, fmt.Errorf("'Expires' missing from response headers")
	}

	expires, err = time.Parse(time.RFC1123, datestring)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"error parsing 'Expires' header: %w",
			err,
		)
	}

	return expires, nil
}

func parseHeadPages(rep *http.Response) (
	pages int,
	err error,
) {
	pagesstring := rep.Header.Get("X-Pages")
	if pagesstring == "" {
		return 0, fmt.Errorf("'X-Pages' missing from response headers")
	}

	pages, err = strconv.Atoi(pagesstring)
	if err != nil {
		return 0, fmt.Errorf(
			"error parsing 'X-Pages' header: %w",
			err,
		)
	}

	return pages, nil
}
