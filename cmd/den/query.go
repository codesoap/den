package main

import (
	"flag"
	"log"
	"time"

	"github.com/codesoap/den/database"
)

func listPictures() {
	if flag.NArg() != 1 {
		log.Fatalln("Too many arguments.")
	}
	if dFlag {
		if err := db.PrintPicturesDetails(); err != nil {
			log.Fatalf("Could not query file statistics: %s\n", err)
		}
	} else {
		f := database.PictureFilter{
			FileFilter: createFilter(),
			Camera:     cameraFlag,
		}
		if err := db.PrintPicturesPaths(f); err != nil {
			log.Fatalf("Could not query files: %s\n", err)
		}
	}
}

func listVideos() {
	if flag.NArg() != 1 {
		log.Fatalln("Too many arguments.")
	}
	if dFlag {
		if err := db.PrintVideosDetails(); err != nil {
			log.Fatalf("Could not query file statistics: %s\n", err)
		}
	} else {
		f := database.VideoFilter{
			FileFilter:  createFilter(),
			MinDuration: durminFlag,
			MaxDuration: durmaxFlag,
			Camera:      cameraFlag,
			MinYear:     recordedFromYear,
			MaxYear:     recordedUntilYear,
		}
		if err := db.PrintVideosPaths(f); err != nil {
			log.Fatalf("Could not query files: %s\n", err)
		}
	}
}

func listAudio() {
	if flag.NArg() != 1 {
		log.Fatalln("Too many arguments.")
	}
	if dFlag {
		if err := db.PrintAudiosDetails(); err != nil {
			log.Fatalf("Could not query file statistics: %s\n", err)
		}
	} else {
		f := database.AudioFilter{
			FileFilter:  createFilter(),
			MinDuration: durminFlag,
			MaxDuration: durmaxFlag,
			Author:      authorFlag,
			MinYear:     recordedFromYear,
			MaxYear:     recordedUntilYear,
		}
		if err := db.PrintAudiosPaths(f); err != nil {
			log.Fatalf("Could not query files: %s\n", err)
		}
	}
}

func listDocuments() {
	if flag.NArg() != 1 {
		log.Fatalln("Too many arguments.")
	}
	if dFlag {
		if err := db.PrintDocumentsDetails(); err != nil {
			log.Fatalf("Could not query file statistics: %s\n", err)
		}
	} else {
		f := database.DocumentFilter{
			FileFilter: createFilter(),
			TxtOnly:    txtFlag,
		}
		if err := db.PrintDocumentsPaths(f); err != nil {
			log.Fatalf("Could not query files: %s\n", err)
		}
	}
}

func listOther() {
	if flag.NArg() != 1 {
		log.Fatalln("Too many arguments.")
	}
	if dFlag {
		if err := db.PrintOthersDetails(); err != nil {
			log.Fatalf("Could not query file statistics: %s\n", err)
		}
	} else {
		if err := db.PrintOthersPaths(createFilter()); err != nil {
			log.Fatalf("Could not query files: %s\n", err)
		}
	}
}

func listAll() {
	if flag.NArg() != 1 {
		log.Fatalln("Too many arguments.")
	}
	if dFlag {
		if err := db.PrintAllDetails(); err != nil {
			log.Fatalf("Could not query file statistics: %s\n", err)
		}
	} else {
		if err := db.PrintAllPaths(createFilter()); err != nil {
			log.Fatalf("Could not query files: %s\n", err)
		}
	}
}

func createFilter() database.FileFilter {
	f := database.FileFilter{}
	if createdFromYear != nil {
		since := time.Date(*createdFromYear, time.January, 0, 0, 0, 0, 0, time.Local)
		f.CreatedSince = &since
	}
	if createdUntilYear != nil {
		until := time.
			Date(*createdUntilYear+1, time.January, 0, 0, 0, 0, 0, time.Local).
			Add(-time.Nanosecond)
		f.CreatedUntil = &until
	}
	return f
}
