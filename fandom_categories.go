package ao3_go

import (
	"regexp"
	"net/http"
	"github.com/PuerkitoBio/goquery"
)

type FandomCategory struct {
	name string
	slug string
}

// GetFandomCategories scrapes the fandoms list.
//
// Endpoint: https://archiveofourown.org/media
func GetFandomCategories() ([]FandomCategory, *AO3Error) {
	const endpoint = "media"

	slugRegex := regexp.MustCompile("^/media/(.+)/fandoms$")

	// Fetch the HTML page and load the document
	res, err := http.Get(baseURL + endpoint)
	if err != nil {
		return nil, WrapError(http.StatusServiceUnavailable, err, "fetching fandom categories returned an err")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, NewError(res.StatusCode, "fetching fandom categories returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing fandom categories page with goquery failed")
	}

	var fandomCategories []FandomCategory

	// Match all the sections, then iterate through them
	categoryMatches := doc.Find(".medium.listbox.group > h3 > a")
	for i := range categoryMatches.Nodes {
		fandomCategory := FandomCategory{}

		categoryNode := categoryMatches.Eq(i)

		// Extract name (e.g., "Anime & Manga")
		name := categoryNode.Text()
		fandomCategory.name = name

		// Extract and parse the slug (e.g., "Anime%20*a*%20Manga")
		link, ok := categoryNode.Attr("href")
		if !ok {
			return nil, NewError(http.StatusUnprocessableEntity, "extracting a fandom categories' category link href failed")
		}

		slug := slugRegex.FindStringSubmatch(link)
		if len(slug) != 2 {
			return nil, NewError(http.StatusUnprocessableEntity, "regexing a fandom categories' category link href failed: "+link)
		}
		fandomCategory.slug = slug[1]

		fandomCategories = append(fandomCategories, fandomCategory)
	}

	return fandomCategories, nil
}