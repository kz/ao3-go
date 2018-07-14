package ao3

import (
	"testing"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

// TestIndexedWorkNode is a general integration test to test the majority of
// fields. Edge cases and missed fields are tested in separate functions.
func TestIndexedWorkNode(t *testing.T) {
	const endpoint = "/works/search?utf8=%E2%9C%93&work_search%5Btitle%5D=No+Bats+In+The+Belfry+%28A+Gotham+High+School+AU%29&work_search%5Bcreators%5D=Airawyn"

	// The actual fields must be equal to the expected fields
	expectedEqual := IndexedWork{
		Title:       "No Bats In The Belfry (A Gotham High School AU)",
		Slug:        "168604",
		LastUpdated: "01 Aug 2011",

		IsAnonymous: false,
		Authors:     []Link{{Text: "Airawyn", Slug: "Airawyn"}},
		Recipients:  []Link{},

		Rating:   "Teen And Up Audiences",
		Warnings: "No Archive Warnings Apply",
		Category: "Gen",
		Status:   "Work in Progress",

		FandomTags: []Link{
			{Text: "DCU - Comicverse", Slug: "DCU%20-%20Comicverse"},
			{Text: "Teen Titans (comic)", Slug: "Teen%20Titans%20(comic)"},
			{Text: "Batman (Comics)", Slug: "Batman%20(Comics)"},
		},
		WarningTags: []Link{
			{Text: "No Archive Warnings Apply", Slug: "No%20Archive%20Warnings%20Apply"},
		},
		RelationshipTags: []Link{
			{Text: "Dick Grayson/Koriand&#39;r", Slug: "Dick%20Grayson*s*Koriand&#39;r"},
			{Text: "Roy Harper/Donna Troy", Slug: "Roy%20Harper*s*Donna%20Troy"},
			{Text: "Jason Todd/Rose Wilson", Slug: "Jason%20Todd*s*Rose%20Wilson"},
		},
		CharacterTags: []Link{
			{Text: "Dick Grayson", Slug: "Dick%20Grayson"},
			{Text: "Tim Drake", Slug: "Tim%20Drake"},
			{Text: "Conner Kent", Slug: "Conner%20Kent"},
			{Text: "Koriand&#39;r", Slug: "Koriand&#39;r"},
			{Text: "Jason Todd", Slug: "Jason%20Todd"},
			{Text: "Damian Wayne", Slug: "Damian%20Wayne"},
			{Text: "Stephanie Brown", Slug: "Stephanie%20Brown"},
			{Text: "Cassandra Cain", Slug: "Cassandra%20Cain"},
			{Text: "Roy Harper", Slug: "Roy%20Harper"},
			{Text: "Kon-El", Slug: "Kon-El"},
		},
		FreeformTags: []Link{
			{Text: "Alternate Universe - High School", Slug: "Alternate%20Universe%20-%20High%20School"},
			{Text: "High School AU", Slug: "High%20School%20AU"},
		},

		IsSeries: true,
		Series: Link{
			Text: "No Bats In The Belfry (A Gotham High School AU)",
			Slug: "6909",
		},
		SeriesPart: 2,

		Summary:  "High school AU featuring the Batkids and various Titans. In part one, Conner Kent arrives at his new school and meets his classmates.",
		Language: "English",
		Words:    4793,
		Chapters: "4/?",
	}

	// The actual fields must be greater than or equal to the expected fields
	expectedMin := Work{
		Comments:  12,
		Kudos:     132,
		Bookmarks: 15,
		Hits:      4659,
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

	// Perform equality tests
	equalityTests := []struct {
		expected interface{}
		actual   interface{}
	}{
		{expectedEqual.Title, work.Title},
		{expectedEqual.Slug, work.Slug},
		{expectedEqual.LastUpdated, work.LastUpdated},
		{expectedEqual.IsAnonymous, work.IsAnonymous},
		{expectedEqual.Authors, work.Authors},
		{expectedEqual.Recipients, work.Recipients},
		{expectedEqual.Rating, work.Rating},
		{expectedEqual.Warnings, work.Warnings},
		{expectedEqual.Category, work.Category},
		{expectedEqual.Status, work.Status},
		{expectedEqual.FandomTags, work.FandomTags},
		{expectedEqual.WarningTags, work.WarningTags},
		{expectedEqual.RelationshipTags, work.RelationshipTags},
		{expectedEqual.CharacterTags, work.CharacterTags},
		{expectedEqual.FreeformTags, work.FreeformTags},
		{expectedEqual.IsSeries, work.IsSeries},
		{expectedEqual.Series, work.Series},
	}

	for _, test := range equalityTests {
		assert.Equal(t, test.expected, test.actual)
	}

	// Perform min tests
	moreThanEqualToTests := []struct {
		expected int
		actual   int
	}{
		{expectedMin.Comments, work.Comments},
		{expectedMin.Kudos, work.Kudos},
		{expectedMin.Bookmarks, work.Bookmarks},
		{expectedMin.Hits, work.Hits},
	}

	for _, test := range moreThanEqualToTests {
		if test.actual < test.expected {
			t.Errorf("Expected not greater than or equal to actual: \n"+
				"expected: %d\n"+
				"actual  : %d\n", test.expected, test.actual)
		}
	}
}

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

func TestIndexedWorkNodeWithSeries(t *testing.T) {
	const endpoint = "/works/search?utf8=%E2%9C%93&work_search%5Btitle%5D=Winnipeg&work_search%5Bcreators%5D=Molly&work_search%5Bfandom_names%5D=Highlander%3A+The+Series&work_search%5Bwarning_ids%5D%5B%5D=16&work_search%5Bcategory_ids%5D%5B%5D=23"

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

	assert.Equal(t, true, work.IsSeries)
	assert.Equal(t, Link{Slug: "62", Text: "Canadian Shack"}, work.Series)
	assert.Equal(t, 1, work.SeriesPart)
}
