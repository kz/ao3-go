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
		t.Error(ao3Err.Error())
	}

	res, err := client.HttpClient.Get(baseURL + endpoint)
		if err != nil {
		t.Error(err.Error())
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Error("fetching page returned a non-200 status code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		t.Error(err.Error())
	}

	workNodes := doc.Find("div.work.listbox.group > ul > li.work.blurb.group")
	if len(workNodes.Nodes) < expectedWorkCount {
		t.Error("unable to find work nodes")
	}

	work, err := client.parseIndexedWorkNode(workNodes.First())
	if err != nil {
		t.Error(err.Error())
	}

	assert.Equal(t, false, work.IsAnonymous)
	assert.Equal(t, expectedAuthors, work.Authors)
}
