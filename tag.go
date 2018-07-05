package ao3

import (
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strconv"
)

// TagWorks is a represented of a paginated /tags/.../works page
type TagWorks struct {
	Works []Work
	Count       int

	// Pagination-related values
	IsPaginated bool
	CurrentPage int
	LastPage    int
}

// GetTagWorks returns a paginated list of works from a tag. A tag can represent
// fandoms, characters, etc.
//
// Endpoint: https://archiveofourown.org/tags/[tag]/works?page=[page]
// Example: https://archiveofourown.org/tags/Action*s*Adventure/works
func (client *AO3Client) GetTagWorks(tag string, page int) (*TagWorks, *AO3Error) {
	endpoint := "/tags/" + tag + "/works"
	if page != 0 {
		endpoint += "?page=" + strconv.Itoa(page)
	}

	res, err := client.HttpClient.Get(baseURL + endpoint)
	if err != nil {
		return nil, WrapError(http.StatusServiceUnavailable, err, "fetching tagged works returned an err")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, NewError(res.StatusCode, "fetching tagged works returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing tagged works page with goquery failed")
	}

	var tagWorks TagWorks

	// Get the number of works returned by the result
	countMatches := doc.Find("#main > h2.heading")
	if len(countMatches.Nodes) != 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "unable to find works count node")
	}

	countRegex := regexp.MustCompile("(?m)(?:.+of )?(\\d+) Work.+")
	matchedCount := countRegex.FindStringSubmatch(countMatches.First().Text())
	tagWorks.Count, err = AtoiWithComma(matchedCount[1])
	if err != nil {
		return nil, NewError(http.StatusUnprocessableEntity, "unable to regex works count")
	}

	// Get pagination details. There are two pagination bars on each page.
	paginationMatches := doc.Find("ol.pagination")
	tagWorks.IsPaginated = len(paginationMatches.Nodes) == 2

	if tagWorks.IsPaginated {
		paginationNode := paginationMatches.First()

		// Get the current page number
		currentMatches := paginationNode.Find("span.current")
		if len(currentMatches.Nodes) != 1 {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to match current page")
		}

		tagWorks.CurrentPage, err = AtoiWithComma(currentMatches.First().Text())
		if err != nil {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to parse current page number")
		}

		// Get the last page number
		// The last page is always the penultimate <li> tag in the <ol> list.
		// Therefore, we assume there must be at least three <li> tags: the
		// previous page link, first page and next page link.
		paginationLinkNodes := paginationNode.Find("li")
		if len(paginationLinkNodes.Nodes) < 3 {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to parse current page number")
		}

		lastWorkNode := paginationLinkNodes.Eq(len(paginationLinkNodes.Nodes) - 2)
		tagWorks.LastPage, err = AtoiWithComma(lastWorkNode.Text())
		if err != nil {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to parse last page number")
		}
	}

	// Fetch the list of works for the page
	tagWorks.Works = []Work{}

	// Matches against the box displaying a single work
	workMatches := doc.Find(".work.blurb.group")
	for i := range workMatches.Nodes {
		node := workMatches.Eq(i)

		work, err := client.parseWorkNode(node)
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing tag work failed")
		}

		tagWorks.Works = append(tagWorks.Works, *work)
	}

	return &tagWorks, nil
}
