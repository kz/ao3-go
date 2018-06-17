package ao3

import (
	"testing"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"strings"
)

func TestGetTaggedWorks(t *testing.T) {
	const tag = "Action*s*Adventure"

	client := InitAO3Client(nil)
	_, err := client.GetTaggedWorks(tag)
	if err != nil {
		t.Error(err.Error())
	}
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