# ao3-go

ao3-go is a Go client for [Archive of our Own](https://archiveofourown.org). **Work in progress.** 

Due to the absence of a HTTP API, this package uses [goquery](https://github.com/PuerkitoBio/goquery) to scrape from the website. As a result, the reliability of the package is tested using integration tests which compare processsed live data against expected values.

This package is designed to be the backend API for the [fanficowl](https://github.com/fanficowl) project. As a result, the API endpoints are tailored towards fanficowl's requirements.

## Supported Endpoints

- [x] `GetFandomCategories` retrieves the list of fandom categories
    - Actual endpoint: `https://archiveofourown.org/media`
- [x] `GetFandomCategory` retrieves the fandoms under a category
    - Actual endpoint: `https://archiveofourown.org/media/[category]/fandoms`
- [ ] `GetTaggedWorks` retrieves a paginated list of works for a tag with optional search parameters
    - Actual endpoint: `https://archiveofourown.org/tags/[tag]/works?page=[page]`
- [ ] `GetTagSearchOptions` retrieves the possible search options for a tag's works
    - Actual endpoint: `https://archiveofourown.org/tags/[tag]/works`
- [ ] `GetAuthorWorks` retrieves a list of works for a author with optional search parameters
    - Actual endpoint: `https://archiveofourown.org/users/[author]/works`
- [ ] `GetAuthorSearchOptions` retrieves the possible search options for a author's works
    - Actual endpoint: `https://archiveofourown.org/users/[author]/works`
- [x] `GetSeriesWorks` retrieves a series' works and its metadata
    - Actual endpoint: `https://archiveofourown.org/series/[series]`
- [x] `GetWork` retrieves the details for a work
    - Actual endpoint: `https://archiveofourown.org/works/[work]?view_adult=true`
- [ ] `DownloadWork` downloads the entire work and returns a byte array
    - Actual endpoint: `https://archiveofourown.org/downloads/[path]`
- [ ] `Authenticate` authenticates the user and retrieves the session cookie
    - Actual endpoint: `https://archiveofourown.org/user_sessions`
    - An initial GET request is required by the scraper in order to obtain the authenticity (CSRF) token
- [ ] `AddKudos` adds kudos to a work
    - Actual endpoint: `https://archiveofourown.org/works/[work]/kudos`
- [ ] `SearchWorks` searches works
    - Actual endpoint: `https://archiveofourown.org/works/search`

## Error Handling
See `ao3_error.go` for the format of all errors handled by this package.

## Known Issues
| Priority | Affected                  | Description                                                                                                                                                                                                                                                                  |
| -------- | --------------            | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| High     | Lists of works            | Listing works for a tag may return a different work count depending on whether the user is logged in. Solution to implement: add option to authenticate users.                                                                                                               |
| Medium   | `IndexedWorkNode`         | `IndexedWorkNode` is missing integration tests. As a fundamental part of this codebase, extensive tests should be written.                                                                                                                                                   |
| Low      | Author links              | Authors which are orphan accounts (e.g., `Lumeilleur` at https://archiveofourown.org/works/4664616) will link to the `orphan_account` user as pseudonyms are ignored.                                                                                                        |
| Low      | Works part of collections | On the website, if a work is part of a collection, its tags' links are prepended with `/collections/[collection]/`. This package ignores these prefixes, so slugs point to the main tags instead, i.e., `/collections/[collection]/tags/[tag]/works` => `/tags/[tag]/works`. |
| Low      | `GetWork`                 | GetWork does not link to collections.                                                                                                                                                                                                                                        |