package ao3_go

import (
	"regexp"
	"strconv"
	"net/http"
	"github.com/PuerkitoBio/goquery"
)

type FandomCategory struct {
	name string
	slug string
}

type Fandom struct {
	name   string
	letter string
	slug   string
	count  int
}

func getFandomCategories() ([]FandomCategory, *AO3Error) {
	const endpoint = "media"
	const baseSelector = ".medium.listbox.group > h3 > a"

	const slugRegex = "^\\/media\\/(.+)\\/fandoms$"
	slugMatcher := regexp.MustCompile(slugRegex)

	// Fetch the HTML page and load the document
	res, err := http.Get(baseURL + endpoint)
	if err != nil {
		return nil, WrapError(http.StatusServiceUnavailable, err, "fetching fandom categories returned an err")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, WrapError(res.StatusCode, err, "fetching fandom categories returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing fandom categories page with goquery failed")
	}

	// Find the categories
	var fandomCategories []FandomCategory
	categoryMatches := doc.Find(baseSelector)
	for i := range categoryMatches.Nodes {
		categoryNode := categoryMatches.Eq(i)

		// Extract name
		name := categoryNode.Text()

		// Extract slug
		matchedCategoryLink, ok := categoryNode.Attr("href")
		if !ok {
			return nil, NewError(http.StatusUnprocessableEntity, "extracting a fandom categories' category link href failed")
		}
		matchedSlug := slugMatcher.FindStringSubmatch(matchedCategoryLink)
		if len(matchedSlug) != 2 {
			return nil, NewError(http.StatusUnprocessableEntity, "regexing a fandom categories' category link href failed: "+matchedCategoryLink)
		}
		slug := matchedSlug[1]

		fandomCategories = append(fandomCategories, FandomCategory{
			name: name,
			slug: slug,
		})
	}

	return fandomCategories, nil
}

func getFandomCategory(category string) ([]Fandom, *AO3Error) {
	endpoint := "media/" + category + "/fandoms"
	const baseSelector = "ol > .letter.listbox.group"
	const headingSelector = "h3"
	const letterFandomSelector = "ul > li"

	const slugRegex = "^\\/tags\\/(.+)\\/works$"
	const letterRegex = "(\\S+)"
	const countRegex = "(?s)^.*\\((\\S+)\\)[\\n\\r\\s]*$"
	letterMatcher := regexp.MustCompile(letterRegex)
	slugMatcher := regexp.MustCompile(slugRegex)
	countMatcher := regexp.MustCompile(countRegex)

	// Fetch the HTML page and load the document
	res, err := http.Get(baseURL + endpoint)
	if err != nil {
		return nil, WrapError(http.StatusServiceUnavailable, err, "fetching fandom category returned an err")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, WrapError(res.StatusCode, err, "fetching fandom category returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing fandom category page with goquery failed")
	}

	// Find the category's fandoms
	var fandoms []Fandom
	categorySectionsMatch := doc.Find(baseSelector)
	for i := range categorySectionsMatch.Nodes {
		categorySectionNode := categorySectionsMatch.Eq(i)

		letterNodeMatches := categorySectionNode.Find(headingSelector)
		if len(letterNodeMatches.Nodes) != 1 {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing a fandom category letter with goquery did not return one node")
		}
		letterNode := letterNodeMatches.First()

		// Extract section letter
		matchedLetter := letterMatcher.FindStringSubmatch(letterNode.Text())
		if len(matchedLetter) != 2 {
			return nil, NewError(http.StatusUnprocessableEntity, "regexing a fandom category letter slug failed: "+letterNode.Text())
		}
		letter := matchedLetter[1]

		// Extract all fandoms under letter
		fandomsMatch := categorySectionNode.Find(letterFandomSelector)
		for i := range fandomsMatch.Nodes {
			fandomNode := fandomsMatch.Eq(i)

			fandomLinkNodeMatches := fandomNode.Find("a")
			if len(fandomLinkNodeMatches.Nodes) != 1 {
				return nil, NewError(http.StatusUnprocessableEntity, "parsing a fandom category's fandom link with goquery did not return one node")
			}
			fandomLinkNode := fandomLinkNodeMatches.First()

			// Extract count
			matchedCount := countMatcher.FindStringSubmatch(fandomNode.Text())
			if len(matchedCount) < 1 {
				return nil, NewError(http.StatusUnprocessableEntity, "regexing a fandom category's fandom count failed: "+fandomNode.Text())
			}
			count, err := strconv.Atoi(matchedCount[len(matchedCount)-1])
			if err != nil {
				return nil, NewError(http.StatusUnprocessableEntity, "strToInt on a fandom category's fandom count failed: "+fandomNode.Text())
			}

			// Extract name
			name := fandomLinkNode.Text()

			// Extract slug
			matchedFandomLink, ok := fandomLinkNode.Attr("href")
			if !ok {
				return nil, NewError(http.StatusUnprocessableEntity, "extracting a fandom category's fandom link href failed")
			}
			matchedSlug := slugMatcher.FindStringSubmatch(matchedFandomLink)
			if len(matchedSlug) != 2 {
				return nil, NewError(http.StatusUnprocessableEntity, "regexing a fandom category's fandom link href failed: "+matchedFandomLink)
			}
			slug := matchedSlug[1]

			fandoms = append(fandoms, Fandom{
				name:   name,
				letter: letter,
				count:  count,
				slug:   slug,
			})
		}
	}

	return fandoms, nil
}
