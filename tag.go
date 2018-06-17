package ao3

import (
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strings"
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

	works := []Work{}

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

			author := Link{
				text: archivistRegex.FindStringSubmatch(headingNode.Text())[1],
				slug: userSlugRegex.FindStringSubmatch(archivistLink)[1],
			}
			work.authors = []Link{author}
		} else {
			// Extract authors and recipients
			var authors []Link
			var recipients []Link

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
					authors = append(authors, user)
				} else {
					recipients = append(recipients, user)
				}
			}
		}

		// Next retrieve the

		works = append(works, work)
	}

	return works, nil
}
