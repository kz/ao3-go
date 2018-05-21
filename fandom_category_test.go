package ao3_go

import (
	"testing"
	"reflect"
)

// func TestRegex(t *testing.T) {
// 	const countRegex = "(?s)^.*\\((\\S+)\\)[\\n\\r\\s]*$"
// 	countMatcher := regexp.MustCompile(countRegex)
// 	text := "\n              07-Ghost\n                (202)\n            "
// 	result := countMatcher.FindStringSubmatch(text)
// 	fmt.Println(result)
//
// }

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

func TestGetFandomCategory(t *testing.T) {
	const exampleCategory = "Anime%20*a*%20Manga"
	const expectedNumExampleFandoms = 1000
	category, err := getFandomCategory(exampleCategory)
	if err != nil {
		t.Fatal("Error while fetching fandom category")
	}

	if len(category) < expectedNumExampleFandoms {
		t.Error("Number of fandoms in the example category is not as expected")
	}
}
