package mediasearch

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

var imdbIDRe = regexp.MustCompile(`\/(tt\d+)\/`)

// searchDuckDuckGo searches DuckDuckGo's HTML interface for an IMDB page
// matching the query, then resolves the IMDB ID via MovieDB.
// This replaces the previous Google "I'm Feeling Lucky" approach, which
// started redirecting EU/UK users to a GDPR consent page.
func searchDuckDuckGo(query, year string, mediatype MediaType) ([]Result, error) {
	q := query
	if year != "" {
		q += " " + year
	}
	if string(mediatype) != "" {
		q += " " + string(mediatype)
	}
	q += " site:imdb.com"
	if debugMode {
		log.Printf("Searching DuckDuckGo for '%s'", q)
	}
	v := url.Values{}
	v.Set("q", q)
	req, err := http.NewRequest("GET", "https://html.duckduckgo.com/html/?"+v.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// extract first IMDB title ID from response HTML
	m := imdbIDRe.FindSubmatch(body)
	if len(m) == 0 {
		return nil, fmt.Errorf("No IMDB match in DuckDuckGo results")
	}
	r, err := imdbGet(imdbID(m[1]), mediatype)
	if err != nil {
		return nil, err
	}
	return []Result{r}, nil
}
