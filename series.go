package ao3

import (
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"regexp"
	"errors"
)

// Series is a representation of the series page
type Series struct {
	Title       string
	IsAnonymous bool
	Creators    []Link
	Begun       string
	Updated     string
	Description string
	Notes       string
	Words       int
	NumWorks    int
	IsComplete  bool
	Bookmarks   int

	Works []IndexedWork
}

// GetSeries returns the metadata and works for a series.
//
// Endpoint: https://archiveofourown.org/series/[series]
func (client *AO3Client) GetSeries(id string) (*Series, *AO3Error) {
	endpoint := "/series/" + id

	res, err := client.HttpClient.Get(baseURL + endpoint)
	if err != nil {
		return nil, WrapError(http.StatusServiceUnavailable, err, "unable to fetch series")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, NewError(res.StatusCode, "fetching series returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to parse series page")
	}

	var series Series

	// Extract title
	titleMatches := doc.Find("div#main > h2.heading")
	if len(titleMatches.Nodes) != 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "unable to match title node")
	}
	series.Title = strings.TrimSpace(titleMatches.First().Text())

	// Extract metadata nodes, ensuring that the number of nodes are even and
	// composed of <dt> tags each directly followed by a <dd> tag.
	metadataMatches := doc.Find("dl.series.meta.group")
	if len(metadataMatches.Nodes) != 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "unable to match metadata node")
	}

	metadataNodes := metadataMatches.Children()
	if len(metadataNodes.Nodes)%2 == 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "unable to match metadata nodes")
	}

	for i := 0; i < len(metadataNodes.Nodes); i += 2 {
		dtNode := metadataNodes.Eq(i)
		ddNode := metadataNodes.Eq(i + 1)

		err := client.parseDescriptionList(dtNode, ddNode, &series)
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to parse metadata node")
		}
	}

	// Fetch the list of works for the page
	series.Works = []IndexedWork{}

	// Matches against the box displaying a single work
	workMatches := doc.Find(".work.blurb.group")
	for i := range workMatches.Nodes {
		node := workMatches.Eq(i)

		work, err := client.parseIndexedWorkNode(node)
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing series work failed")
		}

		series.Works = append(series.Works, *work)
	}

	return &series, nil
}

func (client *AO3Client) parseDescriptionList(definitionNode *goquery.Selection, descriptionNode *goquery.Selection, series *Series) error {
	authorSlugRegex := regexp.MustCompile("/users/(.+)/pseuds/.+")

	if !definitionNode.Is("dt") || !descriptionNode.Is("dd") {
		return errors.New("unable to extract individual metadata pairs")
	}

	// Parse individual metadata
	var err error
	if strings.Contains(definitionNode.Text(), "Creator") {
		// Special cases include having an anonymous author and multiple authors
		if descriptionNode.Text() == "Anonymous" {
			series.IsAnonymous = true
		} else {
			series.IsAnonymous = false

			series.Creators = []Link{}

			authorMatches := descriptionNode.Find("a")
			for i := range authorMatches.Nodes {
				authorNode := authorMatches.Eq(i)

				authorUrl, ok := authorNode.Attr("href")
				if !ok {
					return errors.New("extracting work author link failed")
				}

				authorSlugMatches := authorSlugRegex.FindStringSubmatch(authorUrl)
				if len(authorSlugMatches) != 2 {
					return errors.New("parsing work author link failed")
				}

				series.Creators = append(series.Creators, Link{Text: authorNode.Text(), Slug: authorSlugMatches[1]})
			}
		}
	} else if strings.Contains(definitionNode.Text(), "Series Begun") {
		series.Begun = descriptionNode.Text()
	} else if strings.Contains(definitionNode.Text(), "Series Updated") {
		series.Updated = descriptionNode.Text()
	} else if strings.Contains(definitionNode.Text(), "Description") {
		series.Description = client.HtmlSanitizer.Sanitize(strings.TrimSpace(descriptionNode.Text()))
	} else if strings.Contains(definitionNode.Text(), "Notes") {
		series.Notes = client.HtmlSanitizer.Sanitize(strings.TrimSpace(descriptionNode.Text()))
	} else if strings.Contains(definitionNode.Text(), "Words") {
		series.Words, err = AtoiWithComma(descriptionNode.Text())
		if err != nil {
			return errors.New("converting word count to integer failed")
		}
	} else if strings.Contains(definitionNode.Text(), "Works") {
		series.NumWorks, err = AtoiWithComma(descriptionNode.Text())
		if err != nil {
			return errors.New("converting works count to integer failed")
		}
	} else if strings.Contains(definitionNode.Text(), "Complete") {
		series.IsComplete = descriptionNode.Text() == "Yes"
	} else if strings.Contains(definitionNode.Text(), "Bookmarks") {
		series.Bookmarks, err = AtoiWithComma(strings.TrimSpace(descriptionNode.Text()))
		if err != nil {
			return errors.New("converting bookmark count to integer failed")
		}
	} else if strings.Contains(definitionNode.Text(), "Stats") {
		sublistMatches := descriptionNode.Find("dl.stats")
		if len(sublistMatches.Nodes) != 1 {
			return errors.New("unable to match nested metadata node")
		}

		metadataNodes := sublistMatches.First().Children()
		if len(metadataNodes.Nodes)%2 == 1 {
			return errors.New("unable to match metadata nodes")
		}

		for i := 0; i < len(metadataNodes.Nodes); i += 2 {
			dtNode := metadataNodes.Eq(i)
			ddNode := metadataNodes.Eq(i + 1)

			err := client.parseDescriptionList(dtNode, ddNode, series)
			if err != nil {
				return errors.New("unable to parse metadata node")
			}
		}
	}

	return nil
}
