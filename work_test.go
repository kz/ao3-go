package ao3

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// TestGetWork is an integration test handling a general work. Edge cases which
// are not caught include some multiple optional tags, series, anonymous and
// multiple authors and all HTML tags in summaries.
func TestGetWork(t *testing.T) {
	const workId = "5191202"

	// The actual fields must be equal to the expected fields
	expectedEqual := Work{
		Title:       "A Complete Guide to 'Limited HTML' on AO3",
		IsAnonymous: false,
		Authors:     []Link{{Text: "CodenameCarrot", Slug: "CodenameCarrot"}},

		RatingTags:    []Link{{Text: "General Audiences", Slug: "General%20Audiences"}},
		FandomTags:    []Link{{Text: "No Fandom", Slug: "No%20Fandom"}},
		WarningTags:   []Link{{Text: "No Archive Warnings Apply", Slug: "No%20Archive%20Warnings%20Apply"}},
		CategoryTags:  []Link{{Text: "Gen", Slug: "Gen"}},
		CharacterTags: []Link{},
		FreeformTags: []Link{
			{Text: "HTML", Slug: "HTML"},
			{Text: "Guide", Slug: "Guide"},
			{Text: "cheat sheet", Slug: "cheat%20sheet"},
			{Text: "How-to", Slug: "How-to"},
			{Text: "no story here", Slug: "no%20story%20here"},
			{Text: "reference", Slug: "reference"},
			{Text: "Fanwork Research & Reference Guides", Slug: "Fanwork%20Research%20*a*%20Reference%20Guides"},
		},

		IsSeries: false,

		Language:  "English",
		Published: "2015-11-11",
		Updated:   "2015-11-23",
		Words:     2642,
		Chapters:  "3/4",
	}

	// The actual fields must be greater than or equal to the expected fields
	expectedMin := Work{
		Comments:  126,
		Kudos:     337,
		Bookmarks: 424,
		Hits:      18530,
	}

	// The expected fields must contain the actual fields
	expectedContains := Work{
		Summary:          "a <b><em>comprehensive</em></b> guide, dividing all of the available tags into the following categories:</p><p>1. Text Formatting (in-line HTML)<br/>",
		HTMLDownloadSlug: "Co/CodenameCarrot/5191202/A%20Complete%20Guide%20to%20Limited.html?updated_at=",
	}

	// Fetch the work
	client, err := InitAO3Client(nil, AO3Policy)
	if err != nil {
		t.Fatal(err.Error())
	}

	work, err := client.GetWork(workId)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Perform equality tests
	equalityTests := []struct {
		expected interface{}
		actual   interface{}
	}{
		{expectedEqual.Title, work.Title},
		{expectedEqual.IsAnonymous, work.IsAnonymous},
		{expectedEqual.Authors, work.Authors},
		{expectedEqual.RatingTags, work.RatingTags},
		{expectedEqual.FandomTags, work.FandomTags},
		{expectedEqual.WarningTags, work.WarningTags},
		{expectedEqual.CategoryTags, work.CategoryTags},
		{expectedEqual.CharacterTags, work.CharacterTags},
		{expectedEqual.FreeformTags, work.FreeformTags},
		{expectedEqual.IsSeries, work.IsSeries},
		{expectedEqual.Language, work.Language},
		{expectedEqual.Published, work.Published},
		{expectedEqual.Updated, work.Updated},
		{expectedEqual.Words, work.Words},
		{expectedEqual.Chapters, work.Chapters},
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

	// Perform contains tests
	containsTests := []struct {
		expected interface{}
		actual   interface{}
	}{
		{expectedContains.Summary, work.Summary},
		{expectedContains.HTMLDownloadSlug, work.HTMLDownloadSlug},
	}

	for _, test := range containsTests {
		assert.Contains(t, test.actual, test.expected)
	}
}

// TestGetWorkExtractsSeries checks whether series links are correctly processed
func TestGetWorkExtractsSeries(t *testing.T) {
	const workId = "639"
	expectedEqual := Work {
		IsSeries: true,
		Series: Link{Text:"Canadian Shack", Slug:"62"},
		SeriesPart: 1,
	}

	client, err := InitAO3Client(nil, AO3Policy)
	if err != nil {
		t.Fatal(err.Error())
	}

	work, err := client.GetWork(workId)
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, expectedEqual.IsSeries, work.IsSeries)
	assert.Equal(t, expectedEqual.Series, work.Series)
	assert.Equal(t, expectedEqual.SeriesPart, work.SeriesPart)
}

// TestGetWorkDetectsAnonymousAuthor ensures works with an anonymous author are
// processed correctly
func TestGetWorkDetectsAnonymousAuthor(t *testing.T) {
	const workId = "62903"

	client, err := InitAO3Client(nil, AO3Policy)
	if err != nil {
		t.Fatal(err.Error())
	}

	work, err := client.GetWork(workId)
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, true, work.IsAnonymous)
}

func TestGetWorkExtractsMultipleAuthors(t *testing.T) {
	const workId = "4664616"

	client, err := InitAO3Client(nil, AO3Policy)
	if err != nil {
		t.Fatal(err.Error())
	}

	work, err := client.GetWork(workId)
	if err != nil {
		t.Fatal(err.Error())
	}

	expectedAuthors := []Link{
		{Text: "Afflitto", Slug:"Afflitto"},
		{Text: "Cassbuttstiels", Slug: "Cassbuttstiels"},
		{Text: "HetaliaFanficNetwork", Slug: "HetaliaFanficNetwork"},
		{Text: "Inharborlights", Slug: "Inharborlights"},
		{Text: "Lumeilleur (orphan_account)", Slug: "orphan_account"},
		{Text: "neonferriswheels", Slug: "neonferriswheels"},
		{Text: "orphan_account", Slug: "orphan_account"},
		{Text: "soillse", Slug: "soillse"},
	}

	assert.Equal(t, false, work.IsAnonymous)
	assert.Equal(t, expectedAuthors, work.Authors)
}