package ao3_go

import (
	"testing"
	"reflect"
)

func TestGetFandomCategories(t *testing.T) {
	exampleExpectedCategory := FandomCategory{
		name: "Anime & Manga",
		slug: "Anime%20*a*%20Manga",
	}

	categories, err := getFandomCategories()
	if err != nil {
		t.Fatal("Error while fetching fandom categories")
	}

	if len(categories) == 0 {
		t.Error("Number of fandom categories is not greater than 0")
	}

	hasExampleExpectedCategory := false
	for _, category := range categories {
		if reflect.DeepEqual(category, exampleExpectedCategory) {
			hasExampleExpectedCategory = true
			break
		}
	}

	if !hasExampleExpectedCategory {
		t.Error("Expected example category does not exist")
	}
}