package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	refreshToken string,
	clientId string,
	clientSecret string,
) (
	accessToken string,
	err error,
) {
	req, err := http.NewRequest(
		"GET",
		"https://login.eveonline.com/v2/oauth/token",
		bytes.NewBuffer([]byte(fmt.Sprintf(
			`grant_type=refresh_token&refresh_token=%s`,
			url.QueryEscape(refreshToken),
		))),
	)
	if err != nil {
		log.Fatal(err)
	}
	addHeaderUserAgent(req)
	addHeaderWwwContentType(req)
	addHeaderLoginHost(req)
	addHeadBasicAuth(req, clientId, clientSecret)

	httpRep, close, err := doRequest(req)
	if err != nil {
		return "", err
	}
	defer close()

	var rep EsiAuthRefreshResponse
	err = json.NewDecoder(httpRep.Body).Decode(&rep)
	if err != nil {
		return "", err
	}

	return rep.AccessToken, nil
}

func addHeaderUserAgent(
	req *http.Request,
) {
	req.Header.Add("User-Agent", "eve_industry_program_fetcher")
}

func addHeaderWwwContentType(req *http.Request) {
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
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

// func getModel[M any](
// 	x cache.Context,
// 	url string,
// 	method string,
// 	auth *RefreshTokenAndApp,
// 	newModel func() *M,
// ) (
// 	model M,
// 	expires time.Time,
// 	err error,
// ) {
// 	// get accessToken if auth is not nil
// 	var accessToken string
// 	if auth != nil {
// 		accessToken, _, err = accessTokenGet(x, auth.RefreshToken, auth.App)
// 		if err != nil {
// 			return model, expires, err
// 		}
// 	}

// 	// build the request
// 	var req *http.Request
// 	req, err = newRequest(x.Ctx(), method, url, nil)
// 	if err != nil {
// 		return model, expires, err
// 	}
// 	addHeadJsonContentType(req)
// 	addHeadBearerAuth(req, accessToken)

// 	// fetch the response
// 	var httpRep *http.Response
// 	var close func() error
// 	httpRep, close, err = doRequest(req)
// 	defer close()
// 	if err != nil {
// 		return model, expires, err
// 	}

// 	// decode the body
// 	model, err = decode(httpRep.Body, newRepOrDefault(newModel))
// 	if err != nil {
// 		return model, expires, err
// 	}

// 	// parse the response headers
// 	expires, err = parseHeadExpires(httpRep)
// 	if err != nil {
// 		return model, expires, err
// 	}

// 	return model, expires, nil
// }

// func getHead(
// 	x cache.Context,
// 	url string,
// 	auth *RefreshTokenAndApp,
// ) (
// 	pages int,
// 	expires time.Time,
// 	err error,
// ) {
// 	// get accessToken if auth is not nil
// 	var accessToken string
// 	if auth != nil {
// 		accessToken, _, err = accessTokenGet(x, auth.RefreshToken, auth.App)
// 		if err != nil {
// 			return pages, expires, err
// 		}
// 	}

// 	// build the request
// 	var req *http.Request
// 	req, err = newRequest(x.Ctx(), http.MethodHead, url, nil)
// 	if err != nil {
// 		return pages, expires, err
// 	}
// 	addHeadBearerAuth(req, accessToken)

// 	// fetch the response
// 	var httpRep *http.Response
// 	var close func() error
// 	httpRep, close, err = doRequest(req)
// 	defer close()
// 	if err != nil {
// 		return pages, expires, err
// 	}

// 	// parse the response headers
// 	expires, err = parseHeadExpires(httpRep)
// 	if err != nil {
// 		return pages, expires, err
// 	}
// 	pages, err = parseHeadPages(httpRep)
// 	if err != nil {
// 		return pages, expires, err
// 	}

// 	return pages, expires, nil
// }
