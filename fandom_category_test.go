package ao3

import (
	"testing"
)

func TestGetFandomCategory(t *testing.T) {
	const exampleCategory = "Books%20*a*%20Literature"

	const expectedMinFandomsCount = 1000
	expectedFandom := Fandom{
		name: "Artemis Fowl - Eoin Colfer",
		slug: "Artemis%20Fowl%20-%20Eoin%20Colfer",
	}
	const expectedMinFandomCount = 300

	client := InitAO3Client(nil)
	category, err := client.GetFandomCategory(exampleCategory)
	if err != nil {
		t.Fatal("Error while fetching fandom category:" + err.Error())
	}

	if len(category) < expectedMinFandomsCount {
		t.Fatal("Number of fandoms in the example category is not as expected")
	}

	// Look for an expected fandom, ensuring its count meets the expected minimum
	hasExpectedFandom := false
	for _, fandom := range category {
		if fandom.name != expectedFandom.name {
			continue
		}

		if fandom.slug != expectedFandom.slug {
			t.Fatal("Slug of given fandom does not match expected slug")
		}

		if fandom.count < expectedMinFandomCount {
			t.Fatal("Count of given fandom is less than expected count")
		}

		hasExpectedFandom = true
		break
	}

	if !hasExpectedFandom {
		t.Fatal("Expected fandom not found")
	}
}
