den is a tool for finding files easily. It allows you to quickly
remember which documents you last worked on, list all pictures you took
in 2019, identify full length movies on your disk or all music files of
Daft Punk in your library.

# Installation
den has been tested with Linux and OpenBSD, but will probably also work
with other Unix-like operating systems.

Make sure you have the C library "mediainfo" installed. On Debian, it
can be added with `sudo apt install libmediainfo-dev`. It can be removed
after den has been installed.

Once that's done, den can be installed with this command:

```
go install github.com/codesoap/den/cmd/den@latest
```

# Usage
Here is are a few examples to get started quickly:
```console
$ # Track (non-hidden) files within ~/Documents, ~/Pictures and ~/Music:
$ den track ~/Documents
Indexing 100% (3934/3934)... done
$ den track ~/Music
Indexing 100% (768/768)... done
$ den track ~/Pictures
Indexing 100% (320/320)... done

$ # Check which paths are tracked by den:
$ den list
/home/richard/Documents
/home/richard/Music
/home/richard/Pictures

$ # Find the most recently modified documents:
$ den document | head
/home/richard/Documents/todo.txt
/home/richard/Documents/ephemeral/call_notes.docx
/home/richard/Documents/invoices/invoice_12345.pdf
/home/richard/Music/to_buy.txt
/home/richard/Documents/finance/budget_planning_2025.xlsx
...

$ # Show available filters for pictures:
$ den -d picture
Pictures can be filtered by the year of creation:
        `-c 2017` lists all pictures created in 2017 (8 pictures).
        `-c 2018` lists all pictures created in 2018 (16 pictures).
        `-c 2019` lists all pictures created in 2019 (74 pictures).
        `-c 2020` lists all pictures created in 2020 (21 pictures).
        `-c 2022` lists all pictures created in 2022 (24 pictures).
        `-c 2023` lists all pictures created in 2023 (81 pictures).
        `-c 2024` lists all pictures created in 2024 (91 pictures).
Pictures can be filtered by the camera they were created with:
        `-camera 'CanoScan LiDE 100'` lists all pictures created with that camera (34 pictures).
        `-camera 'Canon EOS 5D'` lists all pictures created with that camera (70 pictures).
        `-camera 'Canon PowerShot A340'` lists all pictures created with that camera (129 pictures).
        `-camera 'Olympus E-330'` lists all pictures created with that camera (18 pictures).
        `-camera 'Apple iPhone 14 Pro'` lists all pictures created with that camera (12 pictures).

$ # Find pictures taken in 2019; use fzf to further filter the results:
$ den -c 2019 picture | fzf
/home/richard/Pictures/Spain/chapel.jpg

$ # Count the number of pictures within the ~/Music directory:
$ den picture ~/Music | wc -l
      12

$ # Whenever significant changes happened in the tracked directories,
$ # you can update the database like this:
$ den rescan
Looking for changes ca. 100% (5024/ca. 5022)... done
(Re-)Indexing 100% (12/12)... done
```

Full usage information:
```console
$ den -h
Usage:
    den (t|track) <PATH>
        Track all files within PATH.
    den (l|list)
        List tracked paths.
    den (u|untrack) <PATH>
    	Stop tracking PATH.
    den (r|rescan)
    	Update database for deleted, changed or added files within all
    	tracked paths.
    den [-d] [FILTER...] (p|picture) [<PREFIX>]
        Print the paths of tracked pictures.
    den [-d] [FILTER...] (v|video) [<PREFIX>]
        Print the paths of tracked videos.
    den [-d] [FILTER...] (a|audio) [<PREFIX>]
        Print the paths of tracked audio files.
    den [-d] [FILTER...] (d|document) [<PREFIX>]
        Print the paths of tracked documents.
    den [-d] [FILTER...] (o|other) [<PREFIX>]
        Print the paths of tracked other files.
    den [-d] [FILTER...] all [<PREFIX>]
        Print the paths of tracked files, regardless of category.

    Within tracked paths, only non-hidden files are considered. When
    printing files, paths will be sorted by modification date, last
    modified first.

    <PREFIX> is an optional filter for the path. E.g. 'den all .' will
    only list files in the current directory.
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
```

# Tips
To use den more easily, you could define small shell functions around
it in your `~/.bshrc`/`~/.zshrc`/etc. E.g. with this function you could
easily find and edit recently modified plain text documents by typing
`recent` in you terminal:

```
recent() {
	f="$(den -txt document | fzf -e --no-sort)"
	test -n "$f" && pushd "$(dirname "$f")" && vim "$(basename "$f")"
}
```

If you end up using this a lot, you might even want to define a key
binding for it. in Zsh, you can bind it to ctrl+Z like this:

```
recent() {
	f="$(den -txt document | fzf -e --no-sort)"
	test -n "$f" && pushd "$(dirname "$f")" && hx "$(basename "$f")"
}
recent-widget() {
	recent
	zle reset-prompt
}
zle -N recent-widget
bindkey '^Z' recent-widget
```

If you want to rescan automatically, you could execute `crontab -e` and
add this line, to rescan every hour (replace `<username>`):

```
0 * * * * /home/<username>/go/bin/den rescan
```
