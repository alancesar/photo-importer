package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/alancesar/photo-importer/cloud"
	"github.com/alancesar/photo-importer/database"
	"github.com/alancesar/photo-importer/file"
	"github.com/alancesar/photo-importer/photo"
	"github.com/alancesar/photo-importer/prompt"
	"github.com/alancesar/tidy-file/command"
	"github.com/alancesar/tidy-file/mime"
	"github.com/alancesar/tidy-file/path"
	"github.com/alancesar/tidy-photo/exif"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	workdir = ".photo-importer"

	_                   = iota
	errHomeDir          = iota
	errInitiateWorkdir  = iota
	errSqlConnection    = iota
	errStartRepository  = iota
	errListVolumes      = iota
	errPromptFailed     = iota
	errInitiateProvider = iota
	errGetProviderPath  = iota

	defaultPhotosPath = "Photos"
)

var (
	photosPath *string
	repository photo.Repository
	volumes    []string

	commands = []command.Command{command.MkDir, command.CopyFile}
	handler  = file.NewHandler(path.Exists, func(path string) (string, error) {
		raw, err := exif.NewReader(path).Extract()
		if err != nil {
			return "", err
		}

		return exif.NewParser(raw).GetChecksum(), nil
	})
)

func init() {
	photosPath = flag.String("-o", defaultPhotosPath, "output photos location")
	flag.Parse()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		os.Exit(errHomeDir)
	}

	if err := os.MkdirAll(workdir, os.ModePerm); err != nil {
		log.Println(err)
		os.Exit(errInitiateWorkdir)
	}

	dbPath := filepath.Join(home, workdir, "photos.db")
	db, err := database.NewSQLiteConnection(dbPath)
	if err != nil {
		log.Println(err)
		os.Exit(errSqlConnection)
	}

	repository, err = photo.NewSQLiteRepository(db)
	if err != nil {
		log.Println(err)
		os.Exit(errStartRepository)
	}

	volumes, err = file.ListVolumes()
	if err != nil {
		fmt.Println(err)
		os.Exit(errListVolumes)
	}
}

func main() {
	device, err := prompt.SelectDevices(volumes)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(errPromptFailed)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	paths := make(chan []string)
	go func(paths chan []string, group *sync.WaitGroup) {
		completeSourcePath := filepath.Join(file.VolumesDir, device, file.PhotosDir)
		paths <- path.LookFor(completeSourcePath, mime.ImageType, mime.ApplicationOctetStreamType)
		wg.Done()
	}(paths, &wg)

	providerName, err := prompt.ProviderNames(cloud.Providers)
	if err != nil {
		fmt.Println(err)
		os.Exit(errPromptFailed)
	}

	provider, err := cloud.NewProvider(providerName)
	if err != nil {
		log.Println(err)
		os.Exit(errInitiateProvider)
	}

	providerPath, err := provider.Location()
	if err != nil {
		log.Println(err)
		os.Exit(errGetProviderPath)
	}

	if exists, err := path.Exists(providerPath); err != nil {
		log.Println(err)
		os.Exit(errGetProviderPath)
	} else if !exists {
		log.Println(fmt.Errorf("%s directory has not been found", providerName))
		os.Exit(errGetProviderPath)
	}

	wg.Wait()
	total := len(<-paths)

	for index, source := range <-paths {
		_, filename := filepath.Split(source)
		raw, err := exif.NewReader(source).Extract()
		if err != nil {
			message := fmt.Sprintf("[ error ] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}
		checksum := fmt.Sprintf("%x", md5.Sum(raw))

		p, err := repository.Get(filename, checksum, providerName)
		if err != nil {
			message := fmt.Sprintf("[ error ] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		if p.Exists() {
			message := fmt.Sprintf("[skipped] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		p.Filename = filename
		p.Provider = providerName
		p.Checksum = checksum

		parser := exif.NewParser(raw)
		t, err := parser.GetDateTime()
		if err != nil {
			message := fmt.Sprintf("[ error ] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		prefix := file.DateToPath(t)
		destination := filepath.Join(providerPath, *photosPath, prefix, filename)
		destination = filepath.Clean(destination)
		duplicated, destination, err := handler.IsDuplicated(destination, checksum)
		if err != nil {
			message := fmt.Sprintf("[ error ] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		if duplicated {
			message := fmt.Sprintf("[skipped] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		wg.Add(2)
		copyErr := make(chan error)
		saveErr := make(chan error)

		go func(err chan error, wg *sync.WaitGroup) {
			err <- command.NewExecutor(source, destination).Execute(commands...)
			wg.Done()
		}(copyErr, &wg)

		go func(err chan error, wg *sync.WaitGroup) {
			err <- repository.Save(&p)
		}(saveErr, &wg)

		wg.Wait()

		if <-copyErr != nil {
			_ = repository.Delete(filename, checksum, providerName)
			message := fmt.Sprintf("[ error ] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		if <-saveErr != nil {
			message := fmt.Sprintf("[ error ] %s (%d of %d)", filename, index+1, total)
			fmt.Println(message)
			continue
		}

		message := fmt.Sprintf("[success] %s (%d of %d)", filename, index+1, total)
		fmt.Println(message)
	}
}
