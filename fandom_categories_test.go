package ao3

import (
	"testing"
	"reflect"
)

func TestGetFandomCategories(t *testing.T) {
	exampleExpectedCategory := FandomCategory{
		name: "Anime & Manga",
		slug: "Anime%20*a*%20Manga",
	}

	client := InitAO3Client(nil)
	categories, err := client.GetFandomCategories()
	if err != nil {
		t.Fatal("Error while fetching fandom categories:" + err.Error())
	}

	if len(categories) == 0 {
		t.Fatal("Number of fandom categories is not greater than 0")
	}

	hasExampleExpectedCategory := false
	for _, category := range categories {
		if reflect.DeepEqual(category, exampleExpectedCategory) {
			hasExampleExpectedCategory = true
			break
		}
	}

	if !hasExampleExpectedCategory {
		t.Fatal("Expected example category does not exist")
	}
}