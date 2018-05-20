package ao3_go

import (
	"github.com/gocolly/colly"
	"log"
	"regexp"
	"errors"
)

type FandomCategory struct {
	name string
	slug string
}

func getFandomCategories() ([]FandomCategory, error) {
	// Matcher constants
	const endpoint = "media"
	const baseSelector = ".medium.listbox.group > h3 > a"
	const slugRegex = "^\\/media\\/(.+)\\/fandoms$"
	r := regexp.MustCompile(slugRegex)

	var fandomCategory []FandomCategory
	var asyncErrors []error

	// Set up collector handlers
	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		asyncErrors = append(asyncErrors, err)
	})

	c.OnHTML(baseSelector, func(e *colly.HTMLElement) {
		// Attempt to extract slug from category URL
		slug := r.FindStringSubmatch(e.Attr("href"))
		if len(slug) != 2 {
			asyncErrors = append(asyncErrors, errors.New("unable to match slug to URL "+e.Attr("href")))
		}

		fandomCategory = append(fandomCategory, FandomCategory{
			name: e.Text,
			slug: slug[1],
		})
	})

	// Run the collector
	if err := c.Visit(baseURL + endpoint); err != nil {
		return nil, err
	}

	c.Wait()
	if len(asyncErrors) != 0 {
		for err := range asyncErrors {
			log.Println("Error handling fandom categories: ", err)
		}
		return nil, errors.New("fandom category handler error occurred")
	}

	return fandomCategory, nil
}
