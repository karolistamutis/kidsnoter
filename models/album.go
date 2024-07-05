package models

type AlbumPage struct {
	count    int    `json:"count,omitempty"`
	next     string `json:"next,omitempty"`
	previous string `json:"previous,omitempty"`
}

type Album struct {
	ID                  int      `json:"id,omitempty"`
	GeneratedFolderName string   `json:"-"`
	Date                string   `json:"created,omitempty"`
	Title               string   `json:"title,omitempty"`
	Content             string   `json:"content,omitempty"`
	Video               *Video   `json:"attached_video,omitempty"`
	Images              []*Image `json:"attached_images,omitempty"`
}

type Video struct {
	ID           int    `json:"id,omitempty"`
	FileName     string `json:"original_file_name,omitempty"`
	FileSize     int    `json:"file_size,omitempty"`
	DownloadLink string `json:"high,omitempty"`
}

type Image struct {
	ID           int    `json:"id,omitempty"`
	FileSize     int    `json:"file_size,omitempty"`
	DownloadLink string `json:"original,omitempty"`
}
