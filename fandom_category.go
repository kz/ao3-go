package ao3_go

import (
	"github.com/gocolly/colly"
	"log"
	"regexp"
	"errors"
	"strconv"
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

func getFandomCategories() ([]FandomCategory, error) {
	// Matcher constants
	const endpoint = "media"
	const baseSelector = ".medium.listbox.group > h3 > a"
	const slugRegex = "^\\/media\\/(.+)\\/fandoms$"
	r := regexp.MustCompile(slugRegex)

	var fandomCategories []FandomCategory
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
			return
		}

		fandomCategories = append(fandomCategories, FandomCategory{
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
		return nil, errors.New("fandom categories handler error occurred")
	}

	return fandomCategories, nil
}

func getFandomCategory(category string) ([]Fandom, error) {
	endpoint := "media/" + category + "/fandoms"
	const baseSelector = "ol > .letter.listbox.group"
	const letterFandomSelector = "ul > li"
	const slugRegex = "^\\/tags\\/(.+)\\/works$"
	const letterRegex = "(\\S+)"
	const countRegex = "(?s)^.*\\((\\S+)\\)[\\n\\r\\s]*$"
	letterMatcher := regexp.MustCompile(letterRegex)
	slugMatcher := regexp.MustCompile(slugRegex)
	countMatcher := regexp.MustCompile(countRegex)

	var fandoms []Fandom
	var asyncErrors []error

	// Set up collector handlers
	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		asyncErrors = append(asyncErrors, err)
	})

	c.OnHTML(baseSelector, func(e *colly.HTMLElement) {
		// Attempt to extract letter from heading
		heading := e.ChildText("h3")
		letter := letterMatcher.FindStringSubmatch(heading)
		if len(letter) != 2 {
			asyncErrors = append(asyncErrors, errors.New("unable to match letter from heading "+heading))
			return
		}

		// Extract fandoms for letter
		e.ForEach(letterFandomSelector, func(_ int, childEl *colly.HTMLElement) {
			// Attempt to extract count from text
			count := countMatcher.FindStringSubmatch(childEl.Text)
			if len(count) < 2 {
				asyncErrors = append(asyncErrors, errors.New("unable to match count from count "+childEl.Text))
				return
			}
			countNum, err := strconv.Atoi(count[len(count) - 1])
			if err != nil {
				asyncErrors = append(asyncErrors, errors.New("unable to convert count to int "+count[len(count) - 1]))
			}

			// Attempt to extract name from text
			name := childEl.ChildText("a")

			// Attempt to extract slug from URL
			slug := slugMatcher.FindStringSubmatch(childEl.ChildAttr("a", "href"))
			if len(slug) != 2 {
				asyncErrors = append(asyncErrors, errors.New("unable to match slug to URL "+childEl.ChildAttr("a", "href")))
				return
			}

			// Append to fandoms
			fandoms = append(fandoms, Fandom{
				name:   name,
				letter: letter[1],
				count:  countNum,
				slug:   slug[1],
			})
		})
	})

	// Run the collector
	if err := c.Visit(baseURL + endpoint); err != nil {
		return nil, err
	}

	c.Wait()
	if len(asyncErrors) != 0 {
		for err := range asyncErrors {
			log.Println("Error handling fandom category: ", err)
		}
		return nil, errors.New("fandom category handler error occurred")
	}

	return fandoms, nil
}
