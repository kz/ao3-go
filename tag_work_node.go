package ao3

import (
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strings"
	"strconv"
	"errors"
)

type Link struct {
	Text string
	Slug string
}

// TagWork represents a work listed in a list of tags
type TagWork struct {
	Title       string
	Slug        string
	LastUpdated string

	IsAnonymous bool
	Authors     []Link
	Recipients  []Link

	Rating   string
	Warnings string
	Category string
	Status   string

	FandomTags       []Link
	WarningTags      []Link
	RelationshipTags []Link
	CharacterTags    []Link
	FreeformTags     []Link

	IsSeries   bool
	Series     Link
	SeriesPart int

	Summary   string
	Language  string
	Words     int
	Chapters  string
	Kudos     int
	Bookmarks int
	Hits      int
}

// parseTagWorkNode parses the standardised listing of a work displayed in the
// search results. The summary is sanitized according to the sanitization policy.
func (client *AO3Client) parseTagWorkNode(node *goquery.Selection) (*TagWork, error) {
	workSlugRegex := regexp.MustCompile("^work_(\\S+)$")
	archivistRegex := regexp.MustCompile("(?m).*by\\s*(.+ \\[archived by .+])")
	userSlugRegex := regexp.MustCompile("^/(?:users/(.+)/(?:pseuds|gifts).*)|(?:gifts\\?recipient=(.+)\\s*)$")
	fandomSlugRegex := regexp.MustCompile("/tags/(.+)/works")
	seriesRegex := regexp.MustCompile("(?m)Part <strong>(.+)</strong> of <a href=\".*/series/(.+)\">(.+)</a>")
	tagRegex := regexp.MustCompile("(?m)<a class=\"tag\" href=\".*/tags/(.+)/works\">(.+)</a>")

	work := TagWork{}

	// Extract the work Slug by matching against the ID
	workLink, ok := node.Attr("id")
	if !ok {
		return nil, errors.New("parsing work Slug failed")
	}

	workSlugMatches := workSlugRegex.FindStringSubmatch(workLink)
	if len(workSlugMatches) != 2 {
		return nil, errors.New("regexing work Slug failed: " + workLink)
	}
	work.Slug = workSlugMatches[1]

	// Extract the last updated string by matching against the datetime class
	lastUpdatedMatches := node.Find(".datetime")
	if len(lastUpdatedMatches.Nodes) != 1 {
		return nil, errors.New("parsing last updated with goquery failed")
	}
	work.LastUpdated = lastUpdatedMatches.First().Text()

	// Extract the fandoms
	work.FandomTags = []Link{}
	fandomMatches := node.Find(".fandoms.heading > a")
	if len(fandomMatches.Nodes) < 1 {
		return nil, errors.New("parsing work fandoms with goquery failed")
	}

	for i := 0; i < len(fandomMatches.Nodes); i++ {
		fandomNode := fandomMatches.Eq(i)

		fandomLink, ok := fandomNode.Attr("href")
		if !ok {
			return nil, errors.New("parsing work fandom link failed")
		}

		fandomSlugMatches := fandomSlugRegex.FindStringSubmatch(fandomLink)
		if len(fandomSlugMatches) != 2 {
			return nil, errors.New("parsing work fandom Slug failed")
		}

		work.FandomTags = append(work.FandomTags, Link{
			Text: fandomNode.Text(),
			Slug: fandomSlugMatches[1],
		})
	}

	// Extract the language string by matching against <dd class="language">
	languageMatches := node.Find("dd.language")
	if len(languageMatches.Nodes) != 1 {
		return nil, errors.New("parsing work language with goquery failed")
	}
	work.Language = languageMatches.First().Text()

	// Extract the words string by matching against <dd class="words">
	// Note that the word count may contain commas (e.g., 3,884) or may not
	// contain any count at all.
	wordsMatches := node.Find("dd.words")
	if len(wordsMatches.Nodes) != 1 {
		return nil, errors.New("parsing work word count with goquery failed")
	}

	var err error
	wordMatch := wordsMatches.First().Text()
	wordCount, err := AtoiWithComma(wordMatch)
	if err == nil {
		work.Words = wordCount
	}

	// Extract the chapters string by matching against <dd class="chapters">
	// Examples: "1/1", "3/?"
	chaptersMatches := node.Find("dd.chapters")
	if len(chaptersMatches.Nodes) != 1 {
		return nil, errors.New("parsing work chapters count failed")
	}
	work.Chapters = chaptersMatches.First().Text()

	// Extract the optional kudos count by matching against <dd class="kudos">
	// It is assumed that the kudos count will not contain numbers
	kudosMatches := node.Find("dd.kudos > a")
	if len(kudosMatches.Nodes) == 1 {
		work.Kudos, err = AtoiWithComma(kudosMatches.First().Text())
		if err != nil {
			return nil, errors.New("parsing work kudos count failed")
		}
	}

	// Extract the optional bookmarks count by matching against <dd class="bookmarks">
	bookmarksMatches := node.Find("dd.bookmarks  > a")
	if len(bookmarksMatches.Nodes) == 1 {
		work.Bookmarks, err = AtoiWithComma(bookmarksMatches.First().Text())
		if err != nil {
			return nil, errors.New("parsing work bookmarks count failed")
		}
	}

	// Extract the optional hits count by matching against <dd class="hits">
	hitsMatches := node.Find("dd.hits")
	if len(hitsMatches.Nodes) == 1 {
		work.Hits, err = AtoiWithComma(hitsMatches.First().Text())
		if err != nil {
			return nil, errors.New("parsing work hits count failed")
		}
	}

	// Extract the optional series
	work.IsSeries = false
	seriesMatches := node.Find(".series > li")
	if len(seriesMatches.Nodes) == 1 {
		// Extract the HTML with format "Part <strong>PART</strong> of <a href="/series/SLUG">TITLE</a>"
		// and use a regex to extract the three relevant parts
		seriesHTML, err := seriesMatches.Html()
		if err != nil {
			return nil, errors.New("parsing work series with goquery failed")
		}

		seriesValueMatches := seriesRegex.FindStringSubmatch(seriesHTML)
		if len(seriesValueMatches) != 4 {
			return nil, errors.New("parsing work series failed")
		}

		work.IsSeries = true
		work.Series.Slug = seriesValueMatches[2]
		work.Series.Text = seriesValueMatches[3]
		work.SeriesPart, err = strconv.Atoi(seriesValueMatches[1])
		if err != nil {
			return nil, errors.New("parsing work series part failed")
		}
	}

	// Extract all optional tags
	optionalTagMatches := node.Find("ul.tags.commas > li")
	for i := range optionalTagMatches.Nodes {
		tagNode := optionalTagMatches.Eq(i)

		// Retrieve the Slug and Text from the nested link
		tagNodeHtml, err := tagNode.Html()
		if err != nil {
			return nil, errors.New("unable to parse tag node HTML")
		}

		tagMatches := tagRegex.FindStringSubmatch(tagNodeHtml)
		if len(tagMatches) != 3 {
			return nil, errors.New("unable to match tag node HTML")
		}
		link := Link{Slug: tagMatches[1], Text: tagMatches[2]}

		// Retrieve the type of tag
		tagType, ok := tagNode.Attr("class")
		if !ok {
			return nil, errors.New("unable to parse tag type")
		}

		if strings.Contains(tagType, "warnings") {
			work.WarningTags = append(work.WarningTags, link)
		} else if strings.Contains(tagType, "relationships") {
			work.RelationshipTags = append(work.RelationshipTags, link)
		} else if strings.Contains(tagType, "characters") {
			work.CharacterTags = append(work.CharacterTags, link)
		} else if strings.Contains(tagType, "freeforms") {
			work.FreeformTags = append(work.FreeformTags, link)
		} else {
			return nil, errors.New("unable to infer tag type")
		}
	}

	// Extract the symbols. There are four symbols: rating, category, warnings
	// and whether a work is in progress (iswip)
	symbolMatches := node.Find(".required-tags")
	if len(symbolMatches.Nodes) != 1 {
		return nil, errors.New("unable to find symbols node")
	}
	symbolNode := symbolMatches.First()

	symbolRatingMatches := symbolNode.Find(".rating")
	if len(symbolRatingMatches.Nodes) != 1 {
		return nil, errors.New("unable to find symbol rating node")
	}

	work.Rating, ok = symbolRatingMatches.First().Attr("title")
	if !ok {
		return nil, errors.New("unable to get symbol rating")
	}

	symbolWarningsMatches := symbolNode.Find(".warnings")
	if len(symbolWarningsMatches.Nodes) != 1 {
		return nil, errors.New("unable to find symbol warnings node")
	}

	work.Warnings, ok = symbolWarningsMatches.First().Attr("title")
	if !ok {
		return nil, errors.New("unable to get symbol warnings")
	}

	symbolCategoryMatches := symbolNode.Find(".category")
	if len(symbolCategoryMatches.Nodes) != 1 {
		return nil, errors.New("unable to find symbol category node")
	}

	work.Category, ok = symbolCategoryMatches.First().Attr("title")
	if !ok {
		return nil, errors.New("unable to get symbol category")
	}

	symbolCompleteMatches := symbolNode.Find(".iswip")
	if len(symbolCompleteMatches.Nodes) != 1 {
		return nil, errors.New("unable to find symbol complete node")
	}

	work.Status, ok = symbolCompleteMatches.First().Attr("title")
	if !ok {
		return nil, errors.New("unable to get symbol complete")
	}

	// Retrieve the summary and sanitize the HTML tags
	summaryMatches := node.Find("blockquote.summary")
	if len(summaryMatches.Nodes) == 1 {
		summaryHTML, err := summaryMatches.First().Html()
		if err != nil {
			return nil, errors.New("unable to fetch HTML from summary node")
		}
		work.Summary = client.HtmlSanitizer.Sanitize(summaryHTML)
	}

	// Retrieve the the header which contains the title, authors and recipients.
	// There is always one author. However, there can optionally be multiple
	// authors and multiple recipients.
	//
	// Additionally, there is an edge case for archived texts, which is displayed as:
	// ```
	// <a href="/works/WORK_ID">WORK_NAME</a>
	//    by
	//    AUTHOR_NAME [archived by <a rel="author" href="ARCHIVIST_ID">ARCHIVIST_NAME</a>]
	// ```
	// In this case, the name should be "AUTHOR_NAME [archived by ARCHIVIST_NAME]"
	// and the link should be the link of the archivist.
	// In all other cases, it suffices to iterate through `a[rel="author"]` nodes.
	headingMatches := node.Find(".header.module > h4.heading")
	if len(headingMatches.Nodes) < 1 {
		return nil, errors.New("parsing work header with goquery failed")
	}
	headingNode := headingMatches.First()

	titleLinkMatches := headingNode.Find("a")
	if len(titleLinkMatches.Nodes) < 1 {
		return nil, errors.New("parsing work title header with goquery failed")
	}

	// Extract the title from the header of the box
	work.Title = titleLinkMatches.First().Text()

	work.IsAnonymous = len(titleLinkMatches.Nodes) == 1
	if !work.IsAnonymous {
		if strings.Contains(headingNode.Text(), "[archived by") {
			// Extract archivist as detailed above
			archivistNode := titleLinkMatches.Eq(1)
			if !archivistRegex.MatchString(archivistNode.Text()) {
				return nil, errors.New("parsing archivist name failed")
			}

			archivistLink, ok := archivistNode.Attr("href")
			if !ok {
				return nil, errors.New("parsing work archivist link failed")
			}
			work.Authors = []Link{{
				Text: archivistRegex.FindStringSubmatch(headingNode.Text())[1],
				Slug: userSlugRegex.FindStringSubmatch(archivistLink)[1],
			}}
		} else {
			// Extract authors and recipients

			// There can be multiple authors and recipients, hence we need to
			// loop through all nodes. Authors will have a "rel" attr with value
			// "author", so we can use that to figure out whether each user is
			// an author or recipient
			for i := 1; i < len(titleLinkMatches.Nodes); i++ {
				userNode := titleLinkMatches.Eq(i)

				user := Link{}

				// Extract user name
				user.Text = userNode.Text()

				// Extract user Slug
				userLink, ok := userNode.Attr("href")
				if !ok {
					return nil, errors.New("parsing work user link failed")
				}

				userSlugMatches := userSlugRegex.FindStringSubmatch(userLink)
				if len(userSlugMatches) != 3 {
					return nil, errors.New("regexing work user Slug failed: " + userLink)
				}

				// Append to the correct type of user
				userRel, ok := userNode.Attr("rel")
				if ok && userRel == "author" {
					user.Slug = userSlugMatches[1]
					work.Authors = append(work.Authors, user)
				} else {
					user.Slug = userSlugMatches[1]
					work.Recipients = append(work.Recipients, user)
				}
			}
		}
	}

	return &work, nil
}
