package youtube

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// YouTube Video meta source url
const URL_META = "http://www.youtube.com/get_video_info?&video_id="

var VideoFormats []string = []string{"3gp", "mp4", "flv", "webm", "avi"}

// Video struct
type Video struct {
	Id, Title, Author, Keywords, Thumbnail_url string
	Avg_rating                                 float32
	View_count, Length_seconds                 int
	Formats                                    []Format
}

type Format struct {
	Itag                     int
	Video_type, Quality, Url string
}

// Given a video id, get it's information from YouTube
func Get(video_id string) (Video, error) {
	// Fetch video meta from YouTube
	query_string, err := fetchMeta(video_id)
	if err != nil {
		return Video{}, err
	}

	meta, err := parseMeta(video_id, query_string)

	if err != nil {
		return Video{}, err
	}

	return meta, nil
}

// Download video
type HttpProgressCallback func(transferred int, total int)

type HttpProgress struct {
	io.ReadCloser
	io.Reader
	total       int
	transferred int
	callback    HttpProgressCallback
}

func (bf *HttpProgress) Read(p []byte) (int, error) {
	readed, err := bf.Reader.Read(p)
	bf.transferred += readed

	if err == nil {
		bf.callback(bf.transferred, bf.total)
	}

	return readed, err
}

func (video *Video) Download(index int, filename string, callback HttpProgressCallback) error {
	// Output file
	out, err := os.Create(filename)
	defer out.Close()

	if err != nil {
		return errors.New("Unable to create file " + filename)
	}

	// Download file
	resp, err := http.Get(video.Formats[index].Url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Unable to download video, status: " + resp.Status)
	}

	// Total size
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

	resp.Body = &HttpProgress{Reader: resp.Body, total: size, callback: callback}

	if err != nil {
		return errors.New("Unable to download video content from YouTube")
	}

	// Write chunks of data to output
	io.Copy(out, resp.Body)

	return nil
}

// Figure out the file extension from a codec string
func (video *Video) GetExtension(index int) string {
	for i := 0; i < len(VideoFormats); i++ {
		if strings.Contains(video.Formats[index].Video_type, VideoFormats[i]) {
			return VideoFormats[i]
		}
	}
	return "avi"
}

// Fetch video meta from http
func fetchMeta(video_id string) (string, error) {
	resp, err := http.Get(URL_META + video_id)

	// Fetch the meta information from http
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	query_string, _ := ioutil.ReadAll(resp.Body)

	return string(query_string), nil
}

// Parse youtube video metadata and return a Video object
func parseMeta(video_id, query_string string) (Video, error) {
	// Parse the query string
	u, _ := url.Parse("?" + query_string)

	// Parse url params
	query := u.Query()

	// No such video
	if query.Get("errorcode") != "" || query.Get("status") == "fail" {
		return Video{}, errors.New(query.Get("reason"))
	}

	// Collate the necessary params
	video := Video{
		Id:            video_id,
		Title:         query.Get("title"),
		Author:        query.Get("author"),
		Keywords:      query.Get("keywords"),
		Thumbnail_url: query.Get("thumbnail_url"),
	}

	v, _ := strconv.Atoi(query.Get("view_count"))
	video.View_count = v

	r, _ := strconv.ParseFloat(query.Get("avg_rating"), 32)
	video.Avg_rating = float32(r)

	l, _ := strconv.Atoi(query.Get("length_seconds"))
	video.Length_seconds = l

	// Further decode the format data
	format_params := strings.Split(query.Get("url_encoded_fmt_stream_map"), ",")

	// Every video has multiple format choices. collate the list.
	for _, f := range format_params {
		furl, _ := url.Parse("?" + f)
		fquery := furl.Query()

		itag, _ := strconv.Atoi(fquery.Get("itag"))

		video.Formats = append(video.Formats, Format{
			Itag:       itag,
			Video_type: fquery.Get("type"),
			Quality:    fquery.Get("quality"),
			Url:        fquery.Get("url") + "&signature=" + fquery.Get("sig"),
		})
	}

	return video, nil
}
