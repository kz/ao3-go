package ao3

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"regexp"
	"errors"
	"strings"
)

type Work struct {
	Title       string
	IsAnonymous bool
	Authors     []Link

	RatingTags    []Link
	FandomTags    []Link
	WarningTags   []Link
	CategoryTags  []Link
	CharacterTags []Link
	FreeformTags  []Link

	IsSeries   bool
	Series     Link
	SeriesPart int

	Language  string
	Published string
	Updated   string
	Words     int
	Chapters  string
	Comments  int
	Kudos     int
	Bookmarks int
	Hits      int

	Summary string

	HTMLDownloadSlug string
}

// GetWork retrieves a work from its page
//
// Endpoint: https://archiveofourown.org/works/[work]?view_adult=true
func (client *AO3Client) GetWork(id string) (*Work, *AO3Error) {
	authorSlugRegex := regexp.MustCompile("/users/(.+)/pseuds/.+")
	seriesRegex := regexp.MustCompile("(?m)Part (.+) of the <a href=\".*/series/(.+)\">(.+)</a> series")
	endpoint := "/works/" + id + "?view_adult=true"

	res, err := client.HttpClient.Get(baseURL + endpoint)
	if err != nil {
		return nil, WrapError(http.StatusServiceUnavailable, err, "fetching work returned an err")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, NewError(res.StatusCode, "fetching work returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing work page with goquery failed")
	}

	var work Work

	// Find the metadata box which contains the majority of information
	metaNodeMatches := doc.Find(".work.meta.group")
	if len(metaNodeMatches.Nodes) != 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "unable to find metadata box on work page")
	}
	metaNode := metaNodeMatches.First()

	// Extract ratings
	work.RatingTags = []Link{}
	ratingNodeMatches := metaNode.Find("dd.rating > ul > li > a")
	if len(ratingNodeMatches.Nodes) > 0 {
		work.RatingTags, err = extractMetadataLinks(ratingNodeMatches)
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to extract rating tags")
		}
	}

	// Extract warnings
	work.WarningTags = []Link{}
	warningNodeMatches := metaNode.Find("dd.warning > ul > li > a")
	if len(warningNodeMatches.Nodes) > 0 {
		work.WarningTags, err = extractMetadataLinks(warningNodeMatches)
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to extract warning tags")
		}
	}

	// Extract categories
	work.CategoryTags = []Link{}
	categoryNodeMatches := metaNode.Find("dd.category > ul > li > a")
	if len(categoryNodeMatches.Nodes) > 0 {
		work.CategoryTags, err = extractMetadataLinks(categoryNodeMatches)
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to extract category tags")
		}
	}

	// Extract character
	work.CharacterTags = []Link{}
	characterNodeMatches := metaNode.Find("dd.character > ul > li > a")
	if len(characterNodeMatches.Nodes) > 0 {
		work.CharacterTags, err = extractMetadataLinks(characterNodeMatches)
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to extract character tags")
		}
	}

	// Extract fandoms
	work.FandomTags = []Link{}
	fandomNodeMatches := metaNode.Find("dd.fandom > ul > li > a")
	if len(fandomNodeMatches.Nodes) > 0 {
		work.FandomTags, err = extractMetadataLinks(fandomNodeMatches)
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to extract fandom tags")
		}
	}

	// Extract freeforms
	work.FreeformTags = []Link{}
	freeformNodeMatches := metaNode.Find("dd.freeform > ul > li > a")
	if len(freeformNodeMatches.Nodes) > 0 {
		work.FreeformTags, err = extractMetadataLinks(freeformNodeMatches)
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to extract freeform tags")
		}
	}

	// Extract language
	languageNodeMatches := metaNode.Find("dd.language")
	if len(languageNodeMatches.Nodes) > 0 {
		work.Language = strings.TrimSpace(languageNodeMatches.First().Text())
	}

	// Extract series
	seriesMatches := metaNode.Find("span.series > span.position")
	work.IsSeries = false
	if len(seriesMatches.Nodes) > 0 {
		seriesHTML, err := seriesMatches.Html()
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing work series with goquery failed")
		}

		seriesValueMatches := seriesRegex.FindStringSubmatch(seriesHTML)
		if len(seriesValueMatches) != 4 {
			return nil, NewError(http.StatusUnprocessableEntity,"parsing work series failed")
		}

		work.IsSeries = true
		work.Series.Slug = seriesValueMatches[2]
		work.Series.Text = seriesValueMatches[3]
		work.SeriesPart, err = AtoiWithComma(seriesValueMatches[1])
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing work series part failed")
		}
	}

	// Extract published
	publishedNodeMatches := metaNode.Find("dd.published")
	if len(publishedNodeMatches.Nodes) > 0 {
		work.Published = publishedNodeMatches.First().Text()
	}

	// Extract updated
	updatedNodeMatches := metaNode.Find("dd.status")
	if len(updatedNodeMatches.Nodes) > 0 {
		work.Updated = updatedNodeMatches.First().Text()
	}

	// Extract words
	wordsNodeMatches := metaNode.Find("dd.words")
	if len(wordsNodeMatches.Nodes) > 0 {
		work.Words, err = AtoiWithComma(wordsNodeMatches.First().Text())
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing work words count failed")
		}
	}

	// Extract chapters
	chaptersNodeMatches := metaNode.Find("dd.chapters")
	if len(chaptersNodeMatches.Nodes) > 0 {
		work.Chapters = chaptersNodeMatches.First().Text()
	}

	// Extract comments
	commentsNodeMatches := metaNode.Find("dd.comments")
	if len(commentsNodeMatches.Nodes) > 0 {
		work.Comments, err = AtoiWithComma(commentsNodeMatches.First().Text())
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing work comments count failed")
		}
	}

	// Extract kudos
	kudosNodeMatches := metaNode.Find("dd.kudos")
	if len(kudosNodeMatches.Nodes) > 0 {
		work.Kudos, err = AtoiWithComma(kudosNodeMatches.First().Text())
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing work kudos count failed")
		}
	}

	// Extract bookmarks
	bookmarksNodeMatches := metaNode.Find("dd.bookmarks")
	if len(bookmarksNodeMatches.Nodes) > 0 {
		work.Bookmarks, err = AtoiWithComma(bookmarksNodeMatches.First().Text())
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing work bookmarks count failed")
		}
	}

	// Extract hits
	hitsNodeMatches := metaNode.Find("dd.hits")
	if len(hitsNodeMatches.Nodes) > 0 {
		work.Hits, err = AtoiWithComma(hitsNodeMatches.First().Text())
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing work hits count failed")
		}
	}

	// Extract title
	titleMatches := doc.Find(".preface > h2.title")
	if len(titleMatches.Nodes) != 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "unable to extract title node")
	}

	work.Title = strings.TrimSpace(titleMatches.First().Text())

	// Extract summary
	summaryMatches := doc.Find(".summary > blockquote.userstuff")
	if len(summaryMatches.Nodes) > 0 {
		summaryHtml, err := summaryMatches.First().Html()
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to extract summary HTML")
		}

		work.Summary = client.HtmlSanitizer.Sanitize(strings.TrimSpace(summaryHtml))
	}

	// Extract author(s), handling the case where the author is anonymous
	authorNodeMatches := doc.Find("#workskin > div.preface > h3.byline.heading")
	if len(authorNodeMatches.Nodes) != 1 {
		return nil, WrapError(http.StatusUnprocessableEntity, err, "unable to extract author node")
	}

	if strings.TrimSpace(authorNodeMatches.First().Text()) == "Anonymous" {
		work.IsAnonymous = true
	} else {
		work.IsAnonymous = false

		work.Authors = []Link{}
		authorMatches := authorNodeMatches.First().Find("a")
		for i := range authorMatches.Nodes {
			authorNode := authorMatches.Eq(i)

			authorUrl, ok := authorNode.Attr("href")
			if !ok {
				return nil, NewError(http.StatusUnprocessableEntity, "extracting work author link failed")
			}

			authorSlugMatches := authorSlugRegex.FindStringSubmatch(authorUrl)
			if len(authorSlugMatches) != 2 {
				return nil, NewError(http.StatusUnprocessableEntity, "parsing work author link failed")
			}

			work.Authors = append(work.Authors, Link{Text: authorNode.Text(), Slug: authorSlugMatches[1]})
		}
	}

	// Extract the HTML download slug
	downloadMatches := doc.Find("li.download > ul > li > a")
	for i := range downloadMatches.Nodes {
		downloadNode := downloadMatches.Eq(i)
		if downloadNode.Text() != "HTML" {
			continue
		}

		// A regex would be unnecessary when we can just trim "/downloads" from
		// the start of the URL
		downloadLink, ok := downloadNode.Attr("href")
		if !ok {
			return  nil, NewError(http.StatusUnprocessableEntity, "retrieving work download link failed")
		}

		work.HTMLDownloadSlug = strings.TrimLeft(downloadLink, "/downloads/")
	}

	if work.HTMLDownloadSlug == "" {
		return nil, NewError(http.StatusUnprocessableEntity, "unable to find work download link")
	}

	return &work, nil
}

func extractMetadataLinks(node *goquery.Selection) ([]Link, error) {
	var tags []Link

	tagRegex := regexp.MustCompile("/tags/(.+)/works")

	for i := range node.Nodes {
		tagNode := node.Eq(i)

		url, ok := tagNode.Attr("href")
		if !ok {
			return nil, errors.New("unable to extract slug from URL")
		}

		matches := tagRegex.FindStringSubmatch(url)
		if len(matches) != 2 {
			return nil, errors.New("unable to extract slug with regex")
		}

		tags = append(tags, Link{Text: tagNode.Text(), Slug: matches[1]})
	}

	return tags, nil
}
