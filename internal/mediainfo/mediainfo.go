package mediainfo

import (
	"strconv"
	"strings"
	"time"

	"github.com/codesoap/jkls-go-mediainfo"
)

type Info struct {
	Type Type

	// If Type is TypeVideo or TypeAudio, look at these fields:
	Duration *time.Duration
	Year     *int

	// If Type is TypeAudio, look at this field:
	Author string

	// If Type is TypeVideo or TypePicture, look at this field:
	Camera string
}

type Type int

const (
	TypeOther   Type = 0
	TypeVideo   Type = 1
	TypeAudio   Type = 2
	TypePicture Type = 3
)

func MediaInfo(path string) (Info, error) {
	i := Info{}
	mi := mediainfo.New()
	if err := mi.Open(path); err != nil {
		return i, err
	}
	if mi.Count(mediainfo.StreamVideo) > 0 {
		i.Type = TypeVideo
		if mi.Count(mediainfo.StreamGeneral) > 0 {
			// There is never more than one general stream, right?
			seconds := mi.Get(mediainfo.StreamGeneral, 0, "Duration")
			if seconds != "" {
				if s, err := strconv.Atoi(seconds); err == nil {
					d := time.Millisecond * time.Duration(s)
					i.Duration = &d
				}
			}
			i.Camera = mi.Get(mediainfo.StreamGeneral, 0, "Encoded_Hardware/String")
			if i.Camera == "" {
				i.Camera = getCamera(mi.Get(mediainfo.StreamGeneral, 0, "Inform"))
			}
			date := mi.Get(mediainfo.StreamGeneral, 0, "Recorded_Date")
			if len(date) >= 4 {
				if y, err := strconv.Atoi(date[:4]); err == nil {
					i.Year = &y
				}
			}
		}
	} else if mi.Count(mediainfo.StreamAudio) > 0 {
		i.Type = TypeAudio
		if mi.Count(mediainfo.StreamGeneral) > 0 {
			// There is never more than one general stream, right?
			seconds := mi.Get(mediainfo.StreamGeneral, 0, "Duration")
			if seconds != "" {
				if s, err := strconv.Atoi(seconds); err == nil {
					d := time.Millisecond * time.Duration(s)
					i.Duration = &d
				}
			}
			i.Author = mi.Get(mediainfo.StreamGeneral, 0, "Performer")
			date := mi.Get(mediainfo.StreamGeneral, 0, "Recorded_Date")
			if len(date) >= 4 {
				if y, err := strconv.Atoi(date[:4]); err == nil {
					i.Year = &y
				}
			}
		}
	} else if mi.Count(mediainfo.StreamImage) > 0 {
		i.Type = TypePicture
		if mi.Count(mediainfo.StreamGeneral) > 0 {
			// There is never more than one general stream, right?
			i.Camera = mi.Get(mediainfo.StreamGeneral, 0, "Encoded_Hardware/String")
			if i.Camera == "" {
				i.Camera = getCamera(mi.Get(mediainfo.StreamGeneral, 0, "Inform"))
			}
		}
	}
	defer mi.Close()
	return i, nil
}

// getCamera tries to find the camera inside manufacturer specific
// fields of the "Inform" value.
func getCamera(informVal string) string {
	for line := range strings.Lines(informVal) {
		s := strings.SplitN(line, ":", 2)
		if len(s) == 2 && strings.TrimSpace(s[0]) == "COM.APPLE.QUICKTIME.MODEL" {
			return strings.TrimSpace(s[1])
		}
	}
	return ""
}
