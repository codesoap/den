package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codesoap/den"
	"github.com/codesoap/den/database"
)

var (
	db database.DB

	createdFromYear, createdUntilYear   *int
	dFlag                               bool
	cameraFlag                          string
	durminFlag, durmaxFlag              *time.Duration
	recordedFromYear, recordedUntilYear *int
	authorFlag                          string
	txtFlag                             bool
)

func init() {
	log.SetFlags(0)
	initDB()
	parseFlags()
}

func initDB() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Could not find directory for database:", err)
	}
	// FIXME: Use more conventional dir on Windows.
	dbDir := filepath.Join(home, ".local", "share", "den")
	if err = os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalln("Could not create database directory:", err)
	}
	dbPath := filepath.Join(dbDir, "db.sqlite3")
	_, err = os.Stat(dbPath)
	db, err = database.NewDB(dbPath)
	if err != nil {
		log.Fatalln("Could not get database:", err)
	}
}

func parseFlags() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage:
    den (t|track) <PATH>
        Track all files within PATH.
    den (l|list)
        List tracked paths.
    den (u|untrack) <PATH>
    	Stop tracking PATH.
    den (r|rescan)
    	Update database for deleted, changed or added files within all
    	tracked paths.
    den [-d] [FILTER...] (p|picture)
        Print the paths of tracked pictures.
    den [-d] [FILTER...] (v|video)
        Print the paths of tracked videos.
    den [-d] [FILTER...] (a|audio)
        Print the paths of tracked audio files.
    den [-d] [FILTER...] (d|document)
        Print the paths of tracked documents.
    den [-d] [FILTER...] (o|other)
        Print the paths of tracked other files.
    den [-d] [FILTER...] all
        Print the paths of tracked files, regardless of category.

    Within tracked paths, only non-hidden files are considered. When
    printing files, paths will be sorted by modification date, last
    modified first.
Options:
    -d
    	Show details about the gathered metadata for the given file type.
Filters:
    -c <YEAR>
        The year in which a file was created. Ranges like 1990-1999 are also
        acceptable.
    -camera <CAMERA>
    	The camera a video or picture has been taken with.
    -durmin <DURATION>
    	The minimum duration of a video or audio file, e.g. 10m or 30s.
    -durmax <DURATION>
    	The maximum duration of a video or audio file, e.g. 10m or 30s.
    -year <YEAR>
    	The year a video or audio file was recorded in. Ranges like
    	1990-1999 are also acceptable.
    -author <AUTHOR>
    	The author (e.g. band) of an audio file.
    -txt
        Show only plain text files when listing documents. These are files
        suitable for editing with a text editor.
`)
		os.Exit(1)
	}
	cFlag := flag.String("c", "", "")
	flag.BoolVar(&dFlag, "d", false, "")
	flag.StringVar(&cameraFlag, "camera", "", "")
	durminFlag = flag.Duration("durmin", 0, "")
	durmaxFlag = flag.Duration("durmax", 1_000_000*time.Hour, "")
	yearFlag := flag.String("year", "", "")
	flag.StringVar(&authorFlag, "author", "", "")
	flag.BoolVar(&txtFlag, "txt", false, "")
	flag.Parse()
	if *cFlag != "" {
		s := strings.Split(*cFlag, "-")
		switch len(s) {
		case 1:
			a, err := strconv.Atoi(s[0])
			if err != nil {
				flag.Usage()
			}
			createdFromYear, createdUntilYear = &a, &a
		case 2:
			a, err := strconv.Atoi(s[0])
			if err != nil {
				flag.Usage()
			}
			b, err := strconv.Atoi(s[1])
			if err != nil {
				flag.Usage()
			}
			createdFromYear, createdUntilYear = &a, &b
		default:
			flag.Usage()
		}
	}
	if *durminFlag == 0 {
		durminFlag = nil
	}
	if *durmaxFlag == 1_000_000*time.Hour {
		durmaxFlag = nil
	}
	if *yearFlag != "" {
		s := strings.Split(*yearFlag, "-")
		switch len(s) {
		case 1:
			a, err := strconv.Atoi(s[0])
			if err != nil {
				flag.Usage()
			}
			recordedFromYear, recordedUntilYear = &a, &a
		case 2:
			a, err := strconv.Atoi(s[0])
			if err != nil {
				flag.Usage()
			}
			b, err := strconv.Atoi(s[1])
			if err != nil {
				flag.Usage()
			}
			recordedFromYear, recordedUntilYear = &a, &b
		default:
			flag.Usage()
		}
	}
	if flag.NArg() < 1 {
		flag.Usage()
	}
}

func main() {
	switch flag.Arg(0) {
	case "t", "track":
		add()
	case "l", "list":
		listTracked()
	case "u", "untrack":
		delete()
	case "r", "rescan":
		rescan()
	case "p", "pic", "picture":
		listPictures()
	case "v", "vid", "video":
		listVideos()
	case "a", "audio":
		listAudio()
	case "d", "doc", "document":
		listDocuments()
	case "o", "other":
		listOther()
	case "all":
		listAll()
	default:
		flag.Usage()
	}
}

func add() {
	if flag.NArg() != 2 {
		log.Fatalln("Give exactly one argument to the add command.")
	}
	path := flag.Arg(1)
	progress := make(chan den.Progress)
	var wg sync.WaitGroup
	wg.Go(func() {
		for {
			prog, ok := <-progress
			if !ok {
				return
			}
			if prog.Total == 0 {
				fmt.Fprintf(os.Stderr, "\rIndexing 0%% (0/0)... ")
			} else {
				fmt.Fprintf(os.Stderr, "\rIndexing %d%% (%d/%d)... ",
					prog.Done*100/prog.Total, prog.Done, prog.Total)
			}
		}
	})
	if err := den.Add(path, db, progress); err != nil {
		log.Fatalln("Could not index dir:", err)
	}
	wg.Wait()
	fmt.Fprintf(os.Stderr, "done\n")
}

func listTracked() {
	if flag.NArg() != 1 {
		log.Fatalln("Got unexpected arguments for the list command.")
	}
	paths, err := den.List(db)
	if err != nil {
		log.Fatalln("Could not list tracked paths:", err)
	}
	slices.Sort(paths)
	for _, p := range paths {
		fmt.Println(p)
	}
}

func delete() {
	if flag.NArg() != 2 {
		log.Fatalln("Give exactly one argument to the delete command.")
	}
	path := flag.Arg(1)
	if err := den.Delete(path, db); err != nil {
		log.Fatalln("Could not delete path:", err)
	}
}

func rescan() {
	if flag.NArg() != 1 {
		log.Fatalln("Got unexpected arguments for the rescan command.")
	}
	checkProgress := make(chan den.Progress)
	indexProgress := make(chan den.Progress)
	var wg sync.WaitGroup
	wg.Go(func() {
		for {
			prog, ok := <-checkProgress
			if !ok {
				break
			}
			if prog.Total == 0 {
				fmt.Fprintf(os.Stderr, "\rLooking for changes ca. 100%% (0/ca. 0)... ")
			} else {
				percent := min(100, prog.Done*100/prog.Total)
				fmt.Fprintf(os.Stderr, "\rLooking for changes ca. %d%% (%d/ca. %d)... ",
					percent, prog.Done, prog.Total)
			}
		}
		checkDone := false
		for {
			prog, ok := <-indexProgress
			if !ok {
				break
			}
			if !checkDone {
				// Assume there was no error in the checking phase, if a (re-)index
				// update has been received.
				fmt.Fprintf(os.Stderr, "done\n")
				checkDone = true
			}
			if prog.Total == 0 {
				fmt.Fprintf(os.Stderr, "\r(Re-)Indexing 100%% (0/0)... ")
			} else {
				fmt.Fprintf(os.Stderr, "\r(Re-)Indexing %d%% (%d/%d)... ",
					prog.Done*100/prog.Total, prog.Done, prog.Total)
			}
		}
	})
	if err := den.Rescan(db, checkProgress, indexProgress); err != nil {
		log.Fatalf("Could not rescan: %s\n", err)
	}
	wg.Wait()
	fmt.Fprintf(os.Stderr, "done\n")
}
