# Coliseu - YouTube library

Authors: Kailash Nadh, Ricardo Pereira

License: GPL v2


# Library

## Methods

### youtube.Get(youtube_video_id)
`video, err = youtube.Get(youtube_video_id)`

Initializes a `Video` object by fetching its metdata from Youtube. `Video` is a struct with the following structure

```go
{
	Id, Title, Author, Keywords, Thumbnail_url string
	Avg_rating float32
	View_count,	Length_seconds int
	Formats []Format
}
```

`Video.Formats` is an array of the `Format` struct, which looks like this:

```
type Format struct {
	Itag int
	Video_type, Quality, Url string
}
```

### youtube.Download(format_index, output_file, callback)
`format_index` is the index of the format listed in the `Video.Formats` array. Youtube offers a number of video formats (mp4, webm, 3gp etc.)

Callback function signature

```go
type HttpProgressCallback func(transferred int, total int)
```

### youtube.GetExtension(format_index)
Guesses the file extension (avi, 3gp, mp4, webm) based on the format chosen

## Example
```go
import (
	youtube "github.com/ricardopereira/coliseu-youtube"
)

func main() {
	// Get the video object (with metdata)
	video, err := youtube.Get("FTl0tl9BGdc")

	// Download the video and write to file
	video.download(0, "video.mp4", func(transferred int, total int) {
	   fmt.Println("Transferred bytes:\r", transferred)
	})
}
```
