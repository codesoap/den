package database

import (
	"fmt"
	"sync"
)

type intcount struct{ num, count int }
type strcount struct {
	str   string
	count int
}

func (db DB) PrintPicturesDetails() error {
	q := `SELECT ` +
		`strftime('%Y', datetime(created_guess, 'unixepoch', 'localtime')) AS year, ` +
		`COUNT(*) ` +
		`FROM file f ` +
		`WHERE EXISTS (SELECT 1 FROM picture p WHERE p.file = f.id) ` +
		`GROUP BY year ` +
		`ORDER BY year `
	res := make(chan intcount)
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("Pictures can be filtered by the year of creation:")
			}
			f := "\t`-c %d` lists all pictures created in %d (%d pictures).\n"
			fmt.Printf(f, r.num, r.num, r.count)
		}
	})
	if err := db.getIntCounts(q, res); err != nil {
		return err
	}
	wg.Wait()

	q = `SELECT camera, COUNT(*) FROM picture ` +
		`WHERE camera IS NOT NULL ` +
		`GROUP BY camera ` +
		`ORDER BY camera `
	res2 := make(chan strcount)
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res2
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("Pictures can be filtered by the camera they were created with:")
			}
			f := "\t`-camera '%s'` lists all pictures created with that camera (%d pictures).\n"
			fmt.Printf(f, r.str, r.count)
		}
	})
	if err := db.getStringCounts(q, res2); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func (db DB) PrintVideosDetails() error {
	q := `SELECT ` +
		`strftime('%Y', datetime(created_guess, 'unixepoch', 'localtime')) AS year, ` +
		`COUNT(*) ` +
		`FROM file f ` +
		`WHERE EXISTS (SELECT 1 FROM video v WHERE v.file = f.id) ` +
		`GROUP BY year ` +
		`ORDER BY year `
	res := make(chan intcount)
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("Videos can be filtered by the year of file creation:")
			}
			f := "\t`-c %d` lists all videos created in %d (%d videos).\n"
			fmt.Printf(f, r.num, r.num, r.count)
		}
	})
	if err := db.getIntCounts(q, res); err != nil {
		return err
	}
	wg.Wait()

	fmt.Printf("Videos can be filtered by their length:\n")
	row := db.d.QueryRow(`SELECT COUNT(*) FROM video WHERE seconds <= 600`)
	var cnt int
	if err := row.Scan(&cnt); err != nil {
		return fmt.Errorf("could not get video count: %s", err)
	}
	fmt.Printf("\t`-durmax 10m` lists all videos up to 10 minutes long (%d videos).\n", cnt)
	row = db.d.QueryRow(`SELECT COUNT(*) FROM video WHERE seconds >= 600`)
	if err := row.Scan(&cnt); err != nil {
		return fmt.Errorf("could not get video count: %s", err)
	}
	fmt.Printf("\t`-durmin 10m` lists all videos at least 10 minutes long (%d videos).\n", cnt)

	q = `SELECT camera, COUNT(*) FROM video ` +
		`WHERE camera IS NOT NULL ` +
		`GROUP BY camera ` +
		`ORDER BY camera `
	res2 := make(chan strcount)
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res2
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("Videos can be filtered by the camera they were created with:")
			}
			f := "\t`-camera '%s'` lists all videos created with that camera (%d videos).\n"
			fmt.Printf(f, r.str, r.count)
		}
	})
	if err := db.getStringCounts(q, res2); err != nil {
		return err
	}
	wg.Wait()

	q = `SELECT ` +
		`CAST((year/10)*10 AS TEXT) || '-' || CAST(((year/10)*10+9) AS TEXT) AS decade,` +
		`COUNT(*) ` +
		`FROM video ` +
		`WHERE year IS NOT NULL ` +
		`GROUP BY decade ` +
		`ORDER BY decade `
	res2 = make(chan strcount)
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res2
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("Videos can be filtered by the year they were recorded:")
			}
			f := "\t`-year %s` lists all videos recorded in that decade (%d videos).\n"
			fmt.Printf(f, r.str, r.count)
		}
	})
	if err := db.getStringCounts(q, res2); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func (db DB) PrintAudiosDetails() error {
	q := `SELECT ` +
		`strftime('%Y', datetime(created_guess, 'unixepoch', 'localtime')) AS year, ` +
		`COUNT(*) ` +
		`FROM file f ` +
		`WHERE EXISTS (SELECT 1 FROM audio a WHERE a.file = f.id) ` +
		`GROUP BY year ` +
		`ORDER BY year `
	res := make(chan intcount)
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("Audio files can be filtered by the year of file creation:")
			}
			f := "\t`-c %d` lists all audio files created in %d (%d files).\n"
			fmt.Printf(f, r.num, r.num, r.count)
		}
	})
	if err := db.getIntCounts(q, res); err != nil {
		return err
	}
	wg.Wait()

	fmt.Printf("Audio files can be filtered by their length:\n")
	row := db.d.QueryRow(`SELECT COUNT(*) FROM audio WHERE seconds <= 600`)
	var cnt int
	if err := row.Scan(&cnt); err != nil {
		return fmt.Errorf("could not get audio file count: %s", err)
	}
	fmt.Printf("\t`-durmax 10m` lists all audio files up to 10 minutes long (%d files).\n", cnt)
	row = db.d.QueryRow(`SELECT COUNT(*) FROM audio WHERE seconds >= 600`)
	if err := row.Scan(&cnt); err != nil {
		return fmt.Errorf("could not get audio file count: %s", err)
	}
	fmt.Printf("\t`-durmin 10m` lists all audio files at least 10 minutes long (%d files).\n", cnt)

	q = `SELECT author, COUNT(*) FROM audio ` +
		`WHERE author IS NOT NULL ` +
		`GROUP BY author ` +
		`ORDER BY author ` +
		`LIMIT 9 `
	res2 := make(chan strcount)
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res2
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("Audio files can be filtered by their author:")
			}
			if i < 8 {
				f := "\t`-author '%s'` lists all files created by that author (%d files).\n"
				fmt.Printf(f, r.str, r.count)
			} else {
				fmt.Println("\t...")
			}
		}
	})
	if err := db.getStringCounts(q, res2); err != nil {
		return err
	}
	wg.Wait()

	q = `SELECT ` +
		`CAST((year/10)*10 AS TEXT) || '-' || CAST(((year/10)*10+9) AS TEXT) AS decade,` +
		`COUNT(*) ` +
		`FROM audio ` +
		`WHERE year IS NOT NULL ` +
		`GROUP BY decade ` +
		`ORDER BY decade `
	res2 = make(chan strcount)
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res2
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("Audio files can be filtered by the year they were recorded:")
			}
			f := "\t`-year %s` lists all files recorded in that decade (%d files).\n"
			fmt.Printf(f, r.str, r.count)
		}
	})
	if err := db.getStringCounts(q, res2); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func (db DB) PrintDocumentsDetails() error {
	q := `SELECT ` +
		`strftime('%Y', datetime(created_guess, 'unixepoch', 'localtime')) AS year, ` +
		`COUNT(*) ` +
		`FROM file f ` +
		`WHERE EXISTS (SELECT 1 FROM document d WHERE d.file = f.id) ` +
		`GROUP BY year ` +
		`ORDER BY year `
	res := make(chan intcount)
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("Documents can be filtered by the year of creation:")
			}
			f := "\t`-c %d` lists all documents created in %d (%d files).\n"
			fmt.Printf(f, r.num, r.num, r.count)
		}
	})
	if err := db.getIntCounts(q, res); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func (db DB) PrintOthersDetails() error {
	q := `SELECT ` +
		`strftime('%Y', datetime(created_guess, 'unixepoch', 'localtime')) AS year, ` +
		`COUNT(*) ` +
		`FROM file f ` +
		`WHERE NOT EXISTS (SELECT 1 FROM picture p WHERE p.file = f.id) ` +
		`AND NOT EXISTS (SELECT 1 FROM video v WHERE v.file = f.id) ` +
		`AND NOT EXISTS (SELECT 1 FROM audio a WHERE a.file = f.id) ` +
		`AND NOT EXISTS (SELECT 1 FROM document d WHERE d.file = f.id) ` +
		`GROUP BY year ` +
		`ORDER BY year `
	res := make(chan intcount)
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("Other files can be filtered by the year of creation:")
			}
			f := "\t`-c %d` lists all files created in %d (%d files).\n"
			fmt.Printf(f, r.num, r.num, r.count)
		}
	})
	if err := db.getIntCounts(q, res); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func (db DB) PrintAllDetails() error {
	q := `SELECT ` +
		`strftime('%Y', datetime(created_guess, 'unixepoch', 'localtime')) AS year, ` +
		`COUNT(*) ` +
		`FROM file f ` +
		`GROUP BY year ` +
		`ORDER BY year `
	res := make(chan intcount)
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := 0; true; i++ {
			r, ok := <-res
			if !ok {
				return
			} else if i == 0 {
				fmt.Println("All files can be filtered by the year of creation:")
			}
			f := "\t`-c %d` lists all files created in %d (%d files).\n"
			fmt.Printf(f, r.num, r.num, r.count)
		}
	})
	if err := db.getIntCounts(q, res); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func (db DB) getIntCounts(query string, out chan intcount) error {
	defer close(out)
	rows, err := db.d.Query(query)
	if err != nil {
		return fmt.Errorf("could not query database: %s", err)
	}
	for rows.Next() {
		var num, count int
		err := rows.Scan(&num, &count)
		if err != nil {
			return fmt.Errorf("could not read from database: %s", err)
		}
		out <- intcount{num: num, count: count}
	}
	return rows.Err()
}

func (db DB) getStringCounts(query string, out chan strcount) error {
	defer close(out)
	rows, err := db.d.Query(query)
	if err != nil {
		return fmt.Errorf("could not query database: %s", err)
	}
	for rows.Next() {
		var str string
		var count int
		err := rows.Scan(&str, &count)
		if err != nil {
			return fmt.Errorf("could not read from database: %s", err)
		}
		out <- strcount{str: str, count: count}
	}
	return rows.Err()
}
