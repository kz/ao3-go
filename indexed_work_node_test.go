package ao3

import (
	"testing"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestIndexedWorkNodeWithArchivist(t *testing.T) {
	const endpoint = "/users/dairesfanficrefuge_archivist"
	const expectedWorkCount = 5
	expectedAuthors := []Link{
		{
			Text: "Ria [archived by dairesfanficrefuge_archivist]",
			Slug: "dairesfanficrefuge_archivist",
		},
	}

	// Fetch the work
	client, ao3Err := InitAO3Client(nil, AO3Policy)
	if ao3Err != nil {
		t.Fatal(ao3Err.Error())
	}

	res, err := client.HttpClient.Get(baseURL + endpoint)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatal("fetching page returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	workNodes := doc.Find("div.work.listbox.group > ul > li.work.blurb.group")
	if len(workNodes.Nodes) < expectedWorkCount {
		t.Fatal("unable to find work nodes")
	}

	work, err := client.parseIndexedWorkNode(workNodes.First())
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, false, work.IsAnonymous)
	assert.Equal(t, expectedAuthors, work.Authors)
}

func TestIndexedWorkNodeWithMultipleAuthorsAndRecipients(t *testing.T) {
	const endpoint = "/works/search?utf8=%E2%9C%93&work_search%5Btitle%5D=Across+the+Stars&work_search%5Bcreators%5D=+jacksqueen16"

	expectedAuthors := []Link{
		{
			Text: "jacksqueen16",
			Slug: "jacksqueen16",
		},
		{
			Text: "TC (thecollective)",
			Slug: "thecollective",
		},
	}

	expectedRecipients := []Link{
		{
			Text: "Aceriee",
			Slug: "Aceriee",
		},
		{
			Text: "Ignisentis",
			Slug: "Ignisentis",
		},
	}

	// Fetch the work
	client, ao3Err := InitAO3Client(nil, AO3Policy)
	if ao3Err != nil {
		t.Fatal(ao3Err.Error())
	}

	res, err := client.HttpClient.Get(baseURL + endpoint)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatal("fetching page returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	workNodeMatches := doc.Find(".work.blurb.group")
	if len(workNodeMatches.Nodes) != 1 {
		t.Fatal("number of work results is not one")
	}

	work, err := client.parseIndexedWorkNode(workNodeMatches.First())
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, false, work.IsAnonymous)
	assert.Equal(t, expectedAuthors, work.Authors)
	assert.Equal(t, expectedRecipients, work.Recipients)
}

func TestIndexedWorkNodeWithAnonymous(t *testing.T) {
	const endpoint = "/works/search?utf8=%E2%9C%93&work_search%5Btitle%5D=Serious+Business&work_search%5Bcreators%5D=Anonymous&work_search%5Bfandom_names%5D=Doctor+Who&work_search%5Bwarning_ids%5D%5B%5D=16&work_search%5Bcategory_ids%5D%5B%5D=21"

	// Fetch the work
	client, ao3Err := InitAO3Client(nil, AO3Policy)
	if ao3Err != nil {
		t.Fatal(ao3Err.Error())
	}

	res, err := client.HttpClient.Get(baseURL + endpoint)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatal("fetching page returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	workNodeMatches := doc.Find(".work.blurb.group")
	if len(workNodeMatches.Nodes) != 1 {
		t.Fatal("number of work results is not one")
	}

	work, err := client.parseIndexedWorkNode(workNodeMatches.First())
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, true, work.IsAnonymous)
	assert.Equal(t, 0, len(work.Authors))
}
