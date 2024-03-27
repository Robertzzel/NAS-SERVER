package models

type File struct {
	Name    string
	Size    int64
	IsDir   bool
	Type    string
	Created int64
}
