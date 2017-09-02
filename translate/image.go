package main

type ImageSearchProvider interface {
	getOneImage(query string)
}
