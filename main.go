package main

import (
	"fmt"
	"github.com/alancesar/photo-importer/database"
	"github.com/alancesar/photo-importer/file"
	"github.com/alancesar/photo-importer/md5"
	"github.com/alancesar/photo-importer/photo"
	"github.com/alancesar/tidy-file/command"
	"github.com/alancesar/tidy-file/mime"
	"github.com/alancesar/tidy-file/path"
	"github.com/alancesar/tidy-photo/exif"
	"log"
	"os"
	"path/filepath"
)

const (
	workdir   = ".photo-importer"
	canonPath = "/Volumes/EOS_DIGITAL/DCIM/100CANON"

	_                    = iota
	errHomeDir           = iota
	errSqlConnection     = iota
	errStartRepository   = iota
	errMD5Checksum       = iota
	errGetFromRepository = iota
	errSaveInRepository  = iota
)

var (
	commands = []command.Command{command.MkDir, command.CopyFile}
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Print(err)
		os.Exit(errHomeDir)
	}

	dbPath := filepath.Join(home, workdir, "photos.db")
	db, err := database.NewSQLiteConnection(dbPath)
	if err != nil {
		log.Print(err)
		os.Exit(errSqlConnection)
	}

	repository, err := photo.NewSQLiteRepository(db)
	if err != nil {
		log.Print(err)
		os.Exit(errStartRepository)
	}

	paths := path.LookFor(canonPath, mime.ImageType, mime.ApplicationOctetStreamType)
	total := len(paths)

	for index, source := range paths {
		checksum, err := md5.CalculateMD5Checksum(source)
		if err != nil {
			log.Print(err)
			os.Exit(errMD5Checksum)
		}

		_, filename := filepath.Split(source)
		p, err := repository.Get(filename, checksum)
		if err != nil {
			log.Print(err)
			os.Exit(errGetFromRepository)
		}

		if p.ID != 0 {
			message := fmt.Sprintf("[skipped] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		raw, err := exif.NewReader(source).Extract()
		if err != nil {
			message := fmt.Sprintf("[error ] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		parser := exif.NewParser(raw)
		destination, err := file.BuildFilename(filename, parser)
		if err != nil {
			message := fmt.Sprintf("[error ] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		if err := command.NewExecutor(source, destination).Execute(commands...); err != nil {
			message := fmt.Sprintf("[error ] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		p.Filename = filename
		p.Checksum = checksum
		if err := repository.Save(&p); err != nil {
			log.Print(err)
			os.Exit(errSaveInRepository)
		}

		message := fmt.Sprintf("[success] %s (%d of %d)", filename, index+1, total)
		fmt.Println(message)
	}
}
