package ao3

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// TestGetSeries is an integration test handling a general series. Edge cases
// including multiple authors/anonymous authors are tested separately.
func TestGetSeries(t *testing.T) {
	const seriesId = "3487"

	// The actual fields must be equal to the expected fields
	expectedEqual := Series{
		Title:       "Kidnapped - Original",
		IsAnonymous: false,
		Creators: []Link{
			{Slug: "EnolaRaven", Text: "EnolaRaven"},
		},
		Begun:      "2003-03-30",
		Updated:    "2010-05-23",
		Words:      101388,
		NumWorks:   2,
		IsComplete: false,
	}

	// The expected fields must contain the actual fields
	expectedContains := Series{
		Description: "Peter Pan doesn&#39;t fear death",
		Notes:       "is for historical archiving only",
	}

	// The actual fields must be greater than or equal to the expected fields
	const expectedMinBookmarks = 4

	// Fetch the work
	client, err := InitAO3Client(nil, AO3Policy)
	if err != nil {
		t.Fatal(err.Error())
	}

	series, err := client.GetSeries(seriesId)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Perform equality tests
	equalityTests := []struct {
		expected interface{}
		actual   interface{}
	}{
		{expectedEqual.Title, series.Title},
		{expectedEqual.IsAnonymous, series.IsAnonymous},
		{expectedEqual.Creators, series.Creators},
		{expectedEqual.Begun, series.Begun},
		{expectedEqual.Updated, series.Updated},
		{expectedEqual.Words, series.Words},
		{expectedEqual.NumWorks, series.NumWorks},
		{expectedEqual.IsComplete, series.IsComplete},
	}

	for _, test := range equalityTests {
		assert.Equal(t, test.expected, test.actual)
	}

	// Perform contains tests
	containsTests := []struct {
		expected interface{}
		actual   interface{}
	}{
		{expectedContains.Description, series.Description},
		{expectedContains.Notes, series.Notes},
	}

	for _, test := range containsTests {
		assert.Contains(t, test.actual, test.expected)
	}

	// Perform min tests
	if series.Bookmarks < expectedMinBookmarks {
		t.Errorf("Expected not greater than or equal to actual: \n"+
			"expected: %d\n"+
			"actual  : %d\n", expectedMinBookmarks, series.Bookmarks)
	}
}

func TestGetSeriesWithMultipleCreators(t *testing.T) {
	const seriesId = "251362"

	expectedCreators := []Link{
		{Slug: "RaijiMagiwind", Text: "RaijiMagiwind"},
		{Slug: "VoiceOfDeath0AyaKnight", Text: "VoiceOfDeath0AyaKnight"},
	}

	// Fetch the work
	client, err := InitAO3Client(nil, AO3Policy)
	if err != nil {
		t.Fatal(err.Error())
	}

	series, err := client.GetSeries(seriesId)
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, false, series.IsAnonymous)
	assert.Equal(t, expectedCreators, series.Creators)
}

func TestGetSeriesWithAnonymousCreator(t *testing.T) {
	const seriesId = "258439"

	// Fetch the work
	client, err := InitAO3Client(nil, AO3Policy)
	if err != nil {
		t.Fatal(err.Error())
	}

	series, err := client.GetSeries(seriesId)
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, true, series.IsAnonymous)
}

