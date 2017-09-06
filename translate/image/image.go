package image

// Image one image
type Image struct {
	Title string
	Link  string
}

// SearchProvider interface for search provider
type SearchProvider interface {
	Search(query string) ([]Image, error)
}
