// Package mimecat can categorize MIME types into these categories:
//   - Other
//   - Video
//   - Picture
//   - Video
//   - Audio
//   - Document
package mimecat

import "strings"

type Category int

const (
	Other = iota
	Picture
	Video
	Audio
	Document
)

func MIMEToCategory(mime string) Category {
	switch mime {
	case "application/ogg":
		return Audio
	case "application/json",
		"application/msword",
		"application/pdf",
		"application/rtf",
		"application/vnd.oasis.opendocument.presentation ",
		"application/vnd.oasis.opendocument.spreadsheet",
		"application/vnd.oasis.opendocument.text",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/x-openscad",
		"application/x-perl",
		"application/x-php",
		"application/x-powershell",
		"application/x-ruby",
		"application/x-troff-man",
		"application/xml",
		"application/yaml":
		return Document
	}
	firstSegment, _, _ := strings.Cut(mime, "/")
	switch firstSegment {
	case "image":
		return Picture
	case "video":
		return Video
	case "audio":
		return Audio
	case "message", "text":
		return Document
	}
	return Other
}
