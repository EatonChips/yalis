package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func login(c *http.Client, username, password string) error {
	loginPageURL := "https://www.linkedin.com/uas/login"
	postURL := "https://www.linkedin.com/uas/login-submit"

	// Make initial request to get cookies and such
	req, err := http.NewRequest("GET", loginPageURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Forward recieved cookies
	forwardCookies(c, resp.Cookies())

	// Build login form from login inputs
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	form := url.Values{}

	doc.Find("input").Each(func(index int, item *goquery.Selection) {
		name, _ := item.Attr("name")
		value, _ := item.Attr("value")
		form.Add(name, value)
	})

	form.Set("session_key", username)
	form.Set("session_password", password)

	// Send login request
	resp, err = c.PostForm(postURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Forward cookies
	forwardCookies(c, resp.Cookies())

	// Read login response
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	// Check to see if login was successful
	title := doc.Find("title").Text()
	if strings.Contains(title, "Security Verification") {
		return errors.New("linkedin thinks you're a bot. fix captcha manually and try again")
	} else if title != "LinkedIn" {
		return errors.New("invalid creds - " + username)
	}

	return nil
}

func getCompanyID(c *http.Client, name string) (string, error) {
	searchURL := "https://www.linkedin.com/voyager/api/search/blended"
	start := 0
	idList := []string{}

	for {
		// Build query
		query := fmt.Sprintf("?count=10&filters=List(resultType->COMPANIES)&keywords=%s&origin=GLOBAL_SEARCH_HEADER&q=all&queryContext=List(spellCorrectionEnabled->true)&start=%d", name, start)

		// Urlencode some chars
		query = strings.ReplaceAll(query, ">", "%3E")
		query = strings.ReplaceAll(query, " ", "%20")

		// Build search request
		req, err := http.NewRequest("GET", searchURL+query, nil)
		if err != nil {
			return "", err
		}
		buildAPIRequest(c.Jar, req)

		// Send search request
		resp, err := c.Do(req)
		if err != nil {
			return "", err
		}

		// Forward recieved cookies
		forwardCookies(c, resp.Cookies())

		// Read server response
		searchResp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		// Parse server response
		searchData := CompanyLookupResponse{}
		err = json.Unmarshal(searchResp, &searchData)
		if err != nil {
			return "", err
		}

		// Display results
		fmt.Println("Choosing company:\n")
		if len(searchData.Data.Elements) == 0 {
			return "", errors.New("No companies found")
		}

		for i, e := range searchData.Data.Elements[0].Elements {
			splitURN := strings.Split(e.TargetURN, ":")
			id := splitURN[len(splitURN)-1]
			idList = append(idList, id)
			fmt.Printf("\t%d) %s - %s\n", start+i, id, e.Title.Text)
		}

		// Get user choice
		input := 0
		fmt.Print("\nEnter index (-1 for more): ")
		_, err = fmt.Scan(&input)
		if err != nil {
			return "", err
		}

		if input >= 0 && input <= len(idList)-1 {
			return idList[input], nil
		}

		start += 10
	}

	return "", nil
}

func getPeople(c *http.Client, companyID string, start int) ([]Person, error) {
	searchEndpoint := "https://www.linkedin.com/voyager/api/search/hits"
	query := fmt.Sprintf("?count=%d&educationEndYear=List()&educationStartYear=List()&facetCurrentCompany=List(%s)&facetCurrentFunction=List()&facetFieldOfStudy=List()&facetGeoRegion=List()&facetNetwork=List()&facetSchool=List()&facetSkillExplicit=List()&keywords=List()&maxFacetValues=15&origin=organization&q=people&start=%d&supportedFacets=List(GEO_REGION,SCHOOL,CURRENT_COMPANY,CURRENT_FUNCTION,FIELD_OF_STUDY,SKILL_EXPLICIT,NETWORK)", count, companyID, start)

	// Build search request
	req, err := http.NewRequest("GET", searchEndpoint+query, nil)
	if err != nil {
		return nil, err
	}
	buildAPIRequest(c.Jar, req)

	// Send search request
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	// Forward cookies
	forwardCookies(c, resp.Cookies())

	// Read server response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse server response
	results := SearchResponse{}
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, err
	}

	people := []Person{}

	for _, p := range results.Included {
		if p.FirstName != "" {
			p.CompanyID = companyID
			people = appendIfMissing(people, p)
		}
	}

	return people, nil
}

// Adds required api headers
func buildAPIRequest(j http.CookieJar, req *http.Request) error {
	csrfToken := getCSRFToken(j)
	req.Header.Set("Csrf-Token", csrfToken)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.linkedin.normalized+json+2.1")
	req.Header.Set("x-restli-protocol-version", "2.0.0")
	req.Header.Add("x-li-lang", "en_US")
	req.Header.Add("x-li-track", "{\"clientVersion\":\"1.5.*\",\"osName\":\"web\",\"timezoneOffset\":-6,\"deviceFormFactor\":\"DESKTOP\",\"mpName\":\"voyager-web\"}")

	return nil
}

// Returns CSRF Token from cookie
func getCSRFToken(jar http.CookieJar) string {
	for _, cookie := range jar.Cookies(cookieURL) {
		if cookie.Name == "JSESSIONID" {
			return cookie.Value
		}
	}

	return ""
}

// Forwards response cookies to client
func forwardCookies(c *http.Client, cookies []*http.Cookie) {
	c.Jar.SetCookies(cookieURL, cookies)
}

// Append if element is not in slice. To keep slice unique
func appendIfMissing(slice []Person, newPerson Person) []Person {
	for _, p := range slice {
		if p.FirstName == newPerson.FirstName && p.LastName == newPerson.LastName {
			return slice
		}
	}

	return append(slice, newPerson)
}
