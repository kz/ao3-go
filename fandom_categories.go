package ao3

import (
	"regexp"
	"net/http"
	"github.com/PuerkitoBio/goquery"
)

type FandomCategory struct {
	Name string
	Slug string
}

// GetFandomCategories scrapes the fandoms list.
//
// Endpoint: https://archiveofourown.org/media
func (client *AO3Client) GetFandomCategories() ([]FandomCategory, *AO3Error) {
	const endpoint = "media"

	slugRegex := regexp.MustCompile("^/media/(.+)/fandoms$")

	// Fetch the HTML page and load the document
	res, err := client.HttpClient.Get(baseURL + endpoint)
	if err != nil {
		return nil, WrapError(http.StatusServiceUnavailable, err, "unable to fetch fandom categories page")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, NewError(res.StatusCode, "fetching fandom categories returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to parse fandom categories page")
	}

	var fandomCategories []FandomCategory

	// Match all the sections, then iterate through them
	categoryMatches := doc.Find(".medium.listbox.group > h3 > a")
	for i := range categoryMatches.Nodes {
		fandomCategory := FandomCategory{}

		categoryNode := categoryMatches.Eq(i)

		// Extract name (e.g., "Anime & Manga")
		name := categoryNode.Text()
		fandomCategory.Name = name

		// Extract and parse the slug (e.g., "Anime%20*a*%20Manga")
		link, ok := categoryNode.Attr("href")
		if !ok {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to find href attribute in category link")
		}

		slug := slugRegex.FindStringSubmatch(link)
		if len(slug) != 2 {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to process category link: "+link)
		}
		fandomCategory.Slug = slug[1]

		fandomCategories = append(fandomCategories, fandomCategory)
	}

	return fandomCategories, nil
}