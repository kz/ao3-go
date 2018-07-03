# ao3-go

A Go client for Archive of Our Own. 

## Known Bugs

- Listing works for a tag may return a different work count depending on whether the user is logged in. Solution to implement: add option to authenticate users
- Works where tags link to a collection (e.g., have /collections/\[collection\]/ appended to /tags/\[tag\]/works) return the main tags instead of a subset of works in the collection  