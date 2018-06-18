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
	title       string
	slug        string
	lastUpdated string
	authors     []Link
	recipients  []Link
	symbols struct {
		rating        string
		relationships string
		warnings      string
		status        string
	}
	tags struct {
		fandoms       []Link
		warnings      []Link
		relationships []Link
		characters    []Link
	}
	series struct {
		name string
		slug string
		part int
	}
	blurb     string
	language  string
	words     int
	chapters  string
	kudos     int
	bookmarks int
	hits      int
}

// GetTaggedWorks returns a paginated list of works from a tag. A tag can represent
// fandoms, characters, etc.
//
// Endpoint: https://archiveofourown.org/tags/[tag]/works
// Example: https://archiveofourown.org/tags/Action*s*Adventure/works
func (client *AO3Client) GetTaggedWorks(tag string) ([]Work, *AO3Error) {
	endpoint := "/tags/" + tag + "/works"

	workSlugRegex := regexp.MustCompile("^work_(\\S+)$")
	archivistRegex := regexp.MustCompile("(?m).*by\\s*(.+ \\[archived by .+])")
	userSlugRegex := regexp.MustCompile("^/users/(.+)/(?:pseuds|gifts).*$")
	fandomSlugRegex := regexp.MustCompile("^/tags/(.+)/works$")
	seriesRegex := regexp.MustCompile("(?m)Part <strong>(.+)</strong> of <a href=\"/series/(.+)\">(.+)</a>")

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
		workNode := workMatches.Eq(i)

		work := Work{}

		// Extract the work slug by matching against the ID
		workLink, ok := workNode.Attr("id")
		if !ok {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work slug failed")
		}

		workSlugMatches := workSlugRegex.FindStringSubmatch(workLink)
		if len(workSlugMatches) != 2 {
			return nil, NewError(http.StatusUnprocessableEntity, "regexing work slug failed: "+workLink)
		}
		work.slug = workSlugMatches[1]

		// Extract the last updated string by matching against the datetime class
		lastUpdatedMatches := workNode.Find(".datetime")
		if len(lastUpdatedMatches.Nodes) != 1 {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing last updated with goquery failed")
		}
		work.lastUpdated = lastUpdatedMatches.First().Text()

		// Extract the fandoms
		work.tags.fandoms = []Link{}
		fandomMatches := workNode.Find(".fandoms.heading > a")
		if len(fandomMatches.Nodes) < 1 {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work fandoms with goquery failed")
		}

		for i := 0; i < len(fandomMatches.Nodes); i++ {
			fandomNode := fandomMatches.Eq(1)
			fandomLink, ok := fandomNode.Attr("href")
			if !ok {
				return nil, NewError(http.StatusUnprocessableEntity, "parsing work fandom link failed")
			}

			work.tags.fandoms = append(work.tags.fandoms, Link{
				text: fandomNode.Text(),
				slug: fandomSlugRegex.FindStringSubmatch(fandomLink)[1],
			})
		}

		// Extract the language string by matching against <dd class="language">
		languageMatches := workNode.Find("dd.language")
		if len(languageMatches.Nodes) != 1 {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work language with goquery failed")
		}
		work.language = languageMatches.First().Text()

		// Extract the words string by matching against <dd class="words">
		// Note that the word count may contain commas (e.g., 3,884)
		wordsMatches := workNode.Find("dd.words")
		if len(wordsMatches.Nodes) != 1 {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work word count with goquery failed")
		}

		work.words, err = strconv.Atoi(wordsMatches.First().Text())
		if err != nil {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work word count failed")
		}

		// Extract the chapters string by matching against <dd class="chapters">
		// Examples: "1/1", "3/?"
		chaptersMatches := workNode.Find("dd.chapters")
		if len(chaptersMatches.Nodes) != 1 {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work chapters count failed")
		}
		work.chapters = chaptersMatches.First().Text()

		// Extract the optional kudos count by matching against <dd class="kudos">
		// It is assumed that the kudos count will not contain numbers
		kudosMatches := workNode.Find("dd.kudos > a")
		if len(kudosMatches.Nodes) == 1 {
			work.kudos, err = strconv.Atoi(kudosMatches.First().Text())
			if err != nil {
				return nil, NewError(http.StatusUnprocessableEntity, "parsing work kudos count failed")
			}
		}

		// Extract the optional bookmarks count by matching against <dd class="bookmarks">
		bookmarksMatches := workNode.Find("dd.bookmarks  > a")
		if len(bookmarksMatches.Nodes) == 1 {
			work.bookmarks, err = strconv.Atoi(bookmarksMatches.First().Text())
			if err != nil {
				return nil, NewError(http.StatusUnprocessableEntity, "parsing work bookmarks count failed")
			}
		}

		// Extract the optional hits count by matching against <dd class="hits">
		hitsMatches := workNode.Find("dd.hits")
		if len(hitsMatches.Nodes) == 1 {
			work.hits, err = strconv.Atoi(hitsMatches.First().Text())
			if err != nil {
				return nil, NewError(http.StatusUnprocessableEntity, "parsing work hits count failed")
			}
		}

		// Extract the optional series
		seriesMatches := workNode.Find(".series > li")
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

			work.series.slug = seriesValueMatches[2]
			work.series.name = seriesValueMatches[3]
			work.series.part, err = strconv.Atoi(seriesValueMatches[1])
			if err != nil {
				return nil, NewError(http.StatusUnprocessableEntity, "parsing work series part failed")
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
		headingMatches := workNode.Find(".header.module > h4.heading")
		if len(headingMatches.Nodes) < 1 {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work header with goquery failed")
		}
		headingNode := headingMatches.First()

		titleLinkMatches := headingNode.Find("a")
		if len(titleLinkMatches.Nodes) < 2 {
			return nil, NewError(http.StatusUnprocessableEntity, "parsing work title header with goquery failed")
		}

		// Extract the title from the header of the box
		work.title = titleLinkMatches.First().Text()

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
			work.authors = []Link{{
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
					work.authors = append(work.authors, user)
				} else {
					work.recipients = append(work.recipients, user)
				}
			}
		}

		works = append(works, work)
	}

	return works, nil
}
