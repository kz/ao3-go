package ao3

import (
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strings"
	"strconv"
)

type Link struct {
	text string
	slug string
}

// Work represents a work listed in a list of tags
// Optional: kudos, bookmarks, hits, recipients, tags, series
type Work struct {
	Title       string
	Slug        string
	LastUpdated string
	Authors     []Link
	Recipients  []Link
	Symbols struct {
		Rating        string
		Relationships string
		Warnings      string
		Status        string
	}
	Tags struct {
		Fandoms       []Link
		Warnings      []Link
		Relationships []Link
		Characters    []Link
		Freeforms     []Link
	}
	Series struct {
		Name string
		Slug string
		Part int
	}
	Blurb     string
	Language  string
	Words     int
	Chapters  string
	Kudos     int
	Bookmarks int
	Hits      int
}

// ParseWorkNode parses the standardised listing of a work displayed in the
// search results
func ParseWorkNode(node *goquery.Selection) (*Work, *AO3Error) {
	workSlugRegex := regexp.MustCompile("^work_(\\S+)$")
	archivistRegex := regexp.MustCompile("(?m).*by\\s*(.+ \\[archived by .+])")
	userSlugRegex := regexp.MustCompile("^/users/(.+)/(?:pseuds|gifts).*$")
	fandomSlugRegex := regexp.MustCompile("^/tags/(.+)/works$")
	seriesRegex := regexp.MustCompile("(?m)Part <strong>(.+)</strong> of <a href=\"/series/(.+)\">(.+)</a>")
	tagRegex := regexp.MustCompile("(?m)<a class=\"tag\" href=\"/tags/(.+)/works\">(.+)</a>")

	work := Work{}

	// Extract the work slug by matching against the ID
	workLink, ok := node.Attr("id")
	if !ok {
		return nil, NewError(http.StatusUnprocessableEntity, "parsing work slug failed")
	}

	workSlugMatches := workSlugRegex.FindStringSubmatch(workLink)
	if len(workSlugMatches) != 2 {
		return nil, NewError(http.StatusUnprocessableEntity, "regexing work slug failed: "+workLink)
	}
	work.Slug = workSlugMatches[1]

	// Extract the last updated string by matching against the datetime class
	lastUpdatedMatches := node.Find(".datetime")
	if len(lastUpdatedMatches.Nodes) != 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "parsing last updated with goquery failed")
	}
	work.LastUpdated = lastUpdatedMatches.First().Text()

	// Extract the fandoms
	work.Tags.Fandoms = []Link{}
	fandomMatches := node.Find(".fandoms.heading > a")
	if len(fandomMatches.Nodes) < 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "parsing work fandoms with goquery failed")
	}

	for i := 0; i < len(fandomMatches.Nodes); i++ {
		fandomNode := fandomMatches.Eq(i)
		fandomLink, ok := fandomNode.Attr("href")
		if !ok {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work fandom link failed")
		}

		work.Tags.Fandoms = append(work.Tags.Fandoms, Link{
			text: fandomNode.Text(),
			slug: fandomSlugRegex.FindStringSubmatch(fandomLink)[1],
		})
	}

	// Extract the language string by matching against <dd class="language">
	languageMatches := node.Find("dd.language")
	if len(languageMatches.Nodes) != 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "parsing work language with goquery failed")
	}
	work.Language = languageMatches.First().Text()

	// Extract the words string by matching against <dd class="words">
	// Note that the word count may contain commas (e.g., 3,884)
	wordsMatches := node.Find("dd.words")
	if len(wordsMatches.Nodes) != 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "parsing work word count with goquery failed")
	}

	var err error
	wordMatch := wordsMatches.First().Text()
	work.Words, err = AtoiWithComma(wordMatch)
	if err != nil {
		return nil, NewError(http.StatusUnprocessableEntity, "parsing work word count failed")
	}

	// Extract the chapters string by matching against <dd class="chapters">
	// Examples: "1/1", "3/?"
	chaptersMatches := node.Find("dd.chapters")
	if len(chaptersMatches.Nodes) != 1 {
		return nil, NewError(http.StatusUnprocessableEntity, "parsing work chapters count failed")
	}
	work.Chapters = chaptersMatches.First().Text()

	// Extract the optional kudos count by matching against <dd class="kudos">
	// It is assumed that the kudos count will not contain numbers
	kudosMatches := node.Find("dd.kudos > a")
	if len(kudosMatches.Nodes) == 1 {
		work.Kudos, err = AtoiWithComma(kudosMatches.First().Text())
		if err != nil {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work kudos count failed")
		}
	}

	// Extract the optional bookmarks count by matching against <dd class="bookmarks">
	bookmarksMatches := node.Find("dd.bookmarks  > a")
	if len(bookmarksMatches.Nodes) == 1 {
		work.Bookmarks, err = AtoiWithComma(bookmarksMatches.First().Text())
		if err != nil {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work bookmarks count failed")
		}
	}

	// Extract the optional hits count by matching against <dd class="hits">
	hitsMatches := node.Find("dd.hits")
	if len(hitsMatches.Nodes) == 1 {
		work.Hits, err = AtoiWithComma(hitsMatches.First().Text())
		if err != nil {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work hits count failed")
		}
	}

	// Extract the optional series
	seriesMatches := node.Find(".series > li")
	if len(seriesMatches.Nodes) == 1 {
		// Extract the HTML with format "Part <strong>PART</strong> of <a href="/series/SLUG">TITLE</a>"
		// and use a regex to extract the three relevant parts
		seriesHTML, err := seriesMatches.Html()
		if err != nil {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work series with goquery failed")
		}

		seriesValueMatches := seriesRegex.FindStringSubmatch(seriesHTML)
		if len(seriesValueMatches) != 4 {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work series failed")
		}

		work.Series.Slug = seriesValueMatches[2]
		work.Series.Name = seriesValueMatches[3]
		work.Series.Part, err = strconv.Atoi(seriesValueMatches[1])
		if err != nil {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work series part failed")
		}
	}

	// Extract all optional tags
	optionalTagMatches := node.Find("ul.tags.commas > li")
	for i := range optionalTagMatches.Nodes {
		tagNode := optionalTagMatches.Eq(i)

		// Retrieve the slug and text from the nested link
		tagNodeHtml, err := tagNode.Html()
		if err != nil {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to parse tag node HTML")
		}

		tagMatches := tagRegex.FindStringSubmatch(tagNodeHtml)
		if len(tagMatches) != 3 {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to match tag node HTML")
		}
		link := Link{slug: tagMatches[1], text: tagMatches[2]}

		// Retrieve the type of tag
		tagType, ok := tagNode.Attr("class")
		if !ok {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to parse tag type")
		}

		if strings.Contains(tagType, "warnings") {
			work.Tags.Warnings = append(work.Tags.Warnings, link)
		} else if strings.Contains(tagType, "relationships") {
			work.Tags.Relationships = append(work.Tags.Relationships, link)
		} else if strings.Contains(tagType, "characters") {
			work.Tags.Characters = append(work.Tags.Characters, link)
		} else if strings.Contains(tagType, "freeforms") {
			work.Tags.Freeforms = append(work.Tags.Freeforms, link)
		} else {
			return nil, NewError(http.StatusUnprocessableEntity, "unable to infer tag type")
		}
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
		return nil, NewError(http.StatusUnprocessableEntity, "parsing work header with goquery failed")
	}
	headingNode := headingMatches.First()

	titleLinkMatches := headingNode.Find("a")
	if len(titleLinkMatches.Nodes) < 2 {
		return nil, NewError(http.StatusUnprocessableEntity, "parsing work title header with goquery failed")
	}

	// Extract the title from the header of the box
	work.Title = titleLinkMatches.First().Text()

	if strings.Contains(headingNode.Text(), "[archived by") {
		// Extract archivist as detailed above
		archivistNode := titleLinkMatches.Eq(1)
		if !archivistRegex.MatchString(archivistNode.Text()) {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing archivist name failed")
		}

		archivistLink, ok := archivistNode.Attr("href")
		if !ok {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work archivist link failed")
		}
		work.Authors = []Link{{
			text: archivistRegex.FindStringSubmatch(headingNode.Text())[1],
			slug: userSlugRegex.FindStringSubmatch(archivistLink)[1],
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
			user.text = userNode.Text()

			// Extract user slug
			userLink, ok := userNode.Attr("href")
			if !ok {
				return nil, NewError(http.StatusUnprocessableEntity, "parsing work author link failed")
			}

			userSlugMatches := userSlugRegex.FindStringSubmatch(userLink)
			if len(userSlugMatches) != 2 {
				return nil, NewError(http.StatusUnprocessableEntity, "regexing work author slug failed: "+userLink)
			}
			user.slug = userSlugMatches[1]

			// Append to the correct type of user
			userRel, ok := userNode.Attr("rel")
			if ok && userRel == "author" {
				work.Authors = append(work.Authors, user)
			} else {
				work.Recipients = append(work.Recipients, user)
			}
		}
	}

	return &work, nil
}

// GetTaggedWorks returns a paginated list of works from a tag. A tag can represent
// fandoms, characters, etc.
//
// Endpoint: https://archiveofourown.org/tags/[tag]/works
// Example: https://archiveofourown.org/tags/Action*s*Adventure/works
func (client *AO3Client) GetTaggedWorks(tag string) ([]Work, *AO3Error) {
	endpoint := "/tags/" + tag + "/works"

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

	var works []Work

	// Matches against the box displaying a single work
	workMatches := doc.Find(".work.blurb.group")
	for i := range workMatches.Nodes {
		node := workMatches.Eq(i)

		work, err := ParseWorkNode(node)
		if err != nil {
			return nil, WrapError(http.StatusUnprocessableEntity, err, "parsing tag work failed")
		}

		works = append(works, *work)
	}

	return works, nil
}
