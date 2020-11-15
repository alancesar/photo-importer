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
	directories := map[string]bool{}

	for index, source := range paths {
		_, filename := filepath.Split(source)
		checksum := calculateMD5Checksum(source)
		p := getFromRepository(filename, checksum, repository)

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

		if err := copyFile(source, destination, directories); err != nil {
			message := fmt.Sprintf("[error ] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		p.Filename = filename
		p.Checksum = checksum
		saveInRepository(p, repository)

		message := fmt.Sprintf("[success] %s (%d of %d)", filename, index+1, total)
		fmt.Println(message)
	}
}

func calculateMD5Checksum(source string) string {
	checksum, err := md5.CalculateMD5Checksum(source)
	if err != nil {
		log.Print(err)
		os.Exit(errMD5Checksum)
	}

	return checksum
}

func getFromRepository(filename, checksum string, repository photo.Repository) photo.Photo {
	p, err := repository.Get(filename, checksum)
	if err != nil {
		log.Print(err)
		os.Exit(errGetFromRepository)
	}

	return p
}

func saveInRepository(p photo.Photo, repository photo.Repository) {
	if err := repository.Save(&p); err != nil {
		log.Print(err)
		os.Exit(errSaveInRepository)
	}
}

func copyFile(source, destination string, directories map[string]bool) error {
	output, _ := filepath.Split(destination)
	commands := createCommands(output, directories)
	if err := command.NewExecutor(source, destination).Execute(commands...); err != nil {
		return err
	}
	directories[output] = true
	return nil
}

func createCommands(output string, directories map[string]bool) []command.Command {
	var commands []command.Command
	if _, exist := directories[output]; !exist {
		commands = append(commands, command.MkDir)
	}
	commands = append(commands, command.CopyFile)
	return commands
}
