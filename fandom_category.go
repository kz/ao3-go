package ao3

import (
	"regexp"
	"strconv"
	"net/http"
	"github.com/PuerkitoBio/goquery"
)

type Fandom struct {
	Name   string
	Letter string
	Slug   string
	Count  int
}

// GetFandomCategory returns a list of all the fandoms under a category.
//
// Endpoint: https://archiveofourown.org/media/[category]/fandoms
// Example: https://archiveofourown.org/media/Anime%20*a*%20Manga/fandoms
func (client *AO3Client) GetFandomCategory(category string) ([]Fandom, *AO3Error) {
	endpoint := "media/" + category + "/fandoms"

	slugRegex := regexp.MustCompile("^/tags/(.+)/works$")
	letterRegex := regexp.MustCompile("(\\S+)")
	countRegex := regexp.MustCompile("(?s)^.*\\((\\S+)\\)[\\n\\r\\s]*$")

	// Fetch the HTML page and load the document
	res, err := client.HttpClient.Get(baseURL + endpoint)
	if err != nil {
		return nil, WrapError(http.StatusServiceUnavailable, err, "unable to fetch fandom category page")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, NewError(res.StatusCode, "fetching fandom category returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to parse fandom category page")
	}

	var fandoms []Fandom

	// Although all the fandoms in a category are under a single page, they are
	// separated by sections corresponding to the first alphanumeric letter of
	// the fandom name. As a result, it is more efficient to iterate through
	// each section so the "letter" is only processed once.
	categorySectionsMatch := doc.Find("ol > .letter.listbox.group")
	for i := range categorySectionsMatch.Nodes {
		categorySectionNode := categorySectionsMatch.Eq(i)

		// Retrieve and parse the letter of the section (e.g., "A")
		letterNodeMatches := categorySectionNode.Find("h3")
		if len(letterNodeMatches.Nodes) != 1 {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to match fandom category letter node")
		}
		letterNode := letterNodeMatches.First()

		matchedLetter := letterRegex.FindStringSubmatch(letterNode.Text())
		if len(matchedLetter) != 2 {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to parse fandom category letter: "+letterNode.Text())
		}
		letter := matchedLetter[1]

		// Extract each fandom under the letter, where each the node contains
		// the name, count and slug of the fandom
		fandomsMatch := categorySectionNode.Find("ul > li")
		for i := range fandomsMatch.Nodes {
			fandomNode := fandomsMatch.Eq(i)

			fandom := Fandom{Letter: letter}

			// Extract and parse fandom works count (e.g., 468)
			matchedCount := countRegex.FindStringSubmatch(fandomNode.Text())
			if len(matchedCount) < 1 {
				return nil, NewError(http.StatusUnprocessableEntity, "unable to parse fandom category work count: "+fandomNode.Text())
			}

			count, err := strconv.Atoi(matchedCount[len(matchedCount)-1])
			if err != nil {
				return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to convert fandom category work count to integer: "+fandomNode.Text())
			}
			fandom.Count = count

			// Extract the node containing the name and slug of the fandom
			fandomLinkNodeMatches := fandomNode.Find("a")
			if len(fandomLinkNodeMatches.Nodes) != 1 {
				return nil, NewError(http.StatusUnprocessableEntity, "unable to match fandom category fandom's name/slug node")
			}
			fandomLinkNode := fandomLinkNodeMatches.First()

			// Extract the name directly (e.g., "Artemis Fowl - Eoin Colfer")
			fandom.Name = fandomLinkNode.Text()

			// Extract and parse the slug (e.g., "Artemis%20Fowl%20-%20Eoin%20Colfer")
			matchedFandomLink, ok := fandomLinkNode.Attr("href")
			if !ok {
				return nil, NewError(http.StatusUnprocessableEntity, "unable to extract href attribute from fandom category fandom's link")
			}

			matchedSlug := slugRegex.FindStringSubmatch(matchedFandomLink)
			if len(matchedSlug) != 2 {
				return nil, NewError(http.StatusUnprocessableEntity, "unable to parse href attribute of fandom category fandom link: "+matchedFandomLink)
			}
			fandom.Slug = matchedSlug[1]

			fandoms = append(fandoms, fandom)
		}
	}

	return fandoms, nil
}
