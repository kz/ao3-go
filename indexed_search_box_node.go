package ao3

import "github.com/PuerkitoBio/goquery"

type SearchOption struct {
	Key  string
	Text string
}

type SearchOptions struct {
	SortBy []SearchOption // Dropdown (default: Date Updated)

	IncludeRatings         []SearchOption // Check
	IncludeWarnings        []SearchOption // Check
	IncludeCategories      []SearchOption // Check
	IncludeFandoms         []SearchOption // Check
	IncludeCharacters      []SearchOption // Check
	IncludeRelationships   []SearchOption // Check
	IncludedAdditionalTags []SearchOption // Check

	ExcludeRatings        []SearchOption // Check
	ExcludeWarnings       []SearchOption // Check
	ExcludeCategories     []SearchOption // Check
	ExcludeFandoms        []SearchOption // Check
	ExcludeCharacters     []SearchOption // Check
	ExcludeRelationships  []SearchOption // Check
	ExcludeAdditionalTags []SearchOption // Check

	Crossovers       []SearchOption // Radio (default: Include crossovers)
	CompletionStatus []SearchOption // Radio (default: All works)
	Language         []SearchOption // Dropdown (default: not selected)
}

const SortByKey = "work_search[sort_column]"

const IncludeRatingsKey = "include_work_search[rating_ids][]"
const IncludeWarningsKey = "include_work_search[warning_ids][]"
const IncludeCategoriesKey = "include_work_search[category_ids][]"
const IncludeFandomsKey = "include_work_search[fandom_ids][]"
const IncludeCharactersKey = "include_work_Search[character_ids][]"
const IncludeRelationshipsKey = "include_work_search[relationship_ids][]"
const IncludedAdditionalTagsKey = "include_work_search[freeform_ids][]"
const IncludeOtherTagsKey = "work_search[other_tag_names]"

const ExcludeRatingsKey = "exclude_work_search[rating_ids][]"
const ExcludeWarningsKey = "exclude_work_search[warning_ids][]"
const ExcludeCategoriesKey = "exclude_work_search[category_ids][]"
const ExcludeFandomsKey = "exclude_work_search[fandom_ids][]"
const ExcludeCharactersKey = "exclude_work_Search[character_ids][]"
const ExcludeRelationshipsKey = "exclude_work_search[relationship_ids][]"
const ExcludeAdditionalTagsKey = "exclude_work_search[freeform_ids][]"
const ExcludeOtherTagsKey = "work_search[excluded_tag_names]"

const CrossoversKey = "work_search[crossover]"
const CompletionStatusKey = "work_search[complete]"
const WordCountFromKey = "work_search[words_from]"
const WordCountToKey = "work_search[words_to]"
const DateUpdatedFromKey = "work_search[date_from]"
const DateUpdatedToKey = "work_search[date_to]"
const DateSearchWithinResultsKey = "work_search[query]"
const LanguageKey = "work_search[language_id]"

type SearchQuery struct {
	SortBy []string

	IncludeRatings         []string
	IncludeWarnings        []string
	IncludeCategories      []string
	IncludeFandoms         []string
	IncludeCharacters      []string
	IncludeRelationships   []string
	IncludedAdditionalTags []string
	IncludeOtherTags       string // Comma separated

	ExcludeRatings        []string
	ExcludeWarnings       []string
	ExcludeCategories     []string
	ExcludeFandoms        []string
	ExcludeCharacters     []string
	ExcludeRelationships  []string
	ExcludeAdditionalTags []string
	ExcludeOtherTags      string // Comma separated

	Crossovers              []string
	CompletionStatus        []string
	WordCountFrom           string
	WordCountTo             string
	DateUpdatedFrom         string
	DateUpdatedTo           string
	DateSearchWithinResults string
	Language                []string
}

// parseIndexedSearchBoxNode parses the "Sort and Filter" box on lists of tagged
// works and author works.
func (client *AO3Client) parseIndexedSearchBoxNode(node *goquery.Selection) (*SearchOptions, error) {
	options := SearchOptions{}

	// Hardcode appropriate options in order to reduce runtime. Each hardcoded
	// case must be explicitly tested for.
	options.SortBy = []SearchOption{
		{Key: "authors_to_sort_on", Text: "Author"},
		{Key: "title_to_sort_on", Text: "Title"},
		{Key: "created_at", Text: "Date Posted"},
		{Key: "revised_at", Text: "Date Updated"},
		{Key: "word_count", Text: "Word Count"},
		{Key: "hits", Text: "Hits"},
		{Key: "kudos_count", Text: "Kudos"},
		{Key: "comments_count", Text: "Comments"},
		{Key: "bookmarks_count", Text: "Bookmarks"},
	}

	options.Crossovers = []SearchOption{
		{Key: "", Text: "Include crossovers"},
		{Key: "F", Text: "Exclude crossovers"},
		{Key: "T", Text: "Show only crossovers"},
	}

	options.CompletionStatus = []SearchOption{
		{Key: "", Text: "All works"},
		{Key: "T", Text: "Complete works only"},
		{Key: "F", Text: "Works in progress only"},
	}

	return nil, nil
}
