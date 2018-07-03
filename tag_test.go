package ao3

import (
	"testing"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"strings"
	"sync"
)

// TestGetTaggedWorks is an integration test to ensure that no errors are raised
// while traversing a sufficient sample set of live tagged works. No comparison
// is made against a hardcoded work in this test.
func TestGetTaggedWorks(t *testing.T) {
	const tag = "No%20Archive%20Warnings%20Apply"
	const testPages = 25

	client, err := InitAO3Client(nil, AO3Policy)
	if err != nil {
		t.Error(err.Error())
	}

	var wg sync.WaitGroup

	for i := 0; i < testPages; i++ {
		wg.Add(1)

		go func(i int) {
			_, err := client.GetTaggedWorks(tag, i)
			if err != nil {
				t.Errorf("error occurred on page %d - %v\n", i, err.Error())
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestAuthors(t *testing.T) {
	res, _ := http.Get("https://archiveofourown.org/users/dairesfanficrefuge_archivist/")
	defer res.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(res.Body)

	workMatches := doc.Find(".work.blurb.group .header.module > h4.heading")
	for i := range workMatches.Nodes {
		workNode := workMatches.Eq(i)
		fmt.Printf("%v///", strings.Contains(workNode.Text(), "[archived by"))
	}
}
