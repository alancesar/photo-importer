package main

import (
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
	"time"
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

	skippedLabel = "[skipped]"
	errorLabel   = "[ error ]"
	successLabel = "[success]"
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

	workdirPath := filepath.Join(home, workdir)
	if err := os.MkdirAll(workdirPath, os.ModePerm); err != nil {
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

	paths := make(chan []string)
	go func() {
		completeSourcePath := filepath.Join(file.VolumesDir, device, file.PhotosDir)
		paths <- path.LookFor(completeSourcePath, mime.ImageType, mime.ApplicationOctetStreamType)
	}()

	var (
		provider     cloud.Provider
		providerName cloud.ProviderName
		providerPath string
	)

	if providerName, err = prompt.ProviderNames(cloud.Providers); err != nil {
		fmt.Println(err)
		os.Exit(errPromptFailed)
	}

	if provider, err = cloud.NewProvider(providerName); err != nil {
		log.Println(err)
		os.Exit(errInitiateProvider)
	}

	if providerPath, err = provider.Location(); err != nil {
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

	sources := <-paths
	total := len(sources)
	for index, source := range sources {
		_, filename := filepath.Split(source)

		logger := func(label string) {
			message := fmt.Sprintf("%s %s (%d of %d)", label, filename, index+1, total)
			fmt.Println(message)
		}

		raw, err := exif.NewReader(source).Extract()
		if err != nil {
			logger(errorLabel)
			continue
		}
		parser := exif.NewParser(raw)
		checksum := parser.GetChecksum()

		var p photo.Photo
		if p, err = repository.Get(filename, checksum, providerName); err != nil {
			logger(errorLabel)
			continue
		} else if p.Exists() {
			logger(skippedLabel)
			continue
		}

		p.Filename = filename
		p.Provider = providerName
		p.Checksum = checksum

		var t time.Time
		if t, err = parser.GetDateTime(); err != nil {
			logger(errorLabel)
			continue
		}

		prefix := file.DateToPath(t)
		destination := filepath.Join(providerPath, *photosPath, prefix, filename)
		destination = filepath.Clean(destination)

		var duplicated bool
		if duplicated, destination, err = handler.IsDuplicated(destination, checksum); err != nil {
			logger(errorLabel)
			continue
		} else if duplicated {
			logger(skippedLabel)
			continue
		}

		copyErr := make(chan error)
		saveErr := make(chan error)

		go func() {
			copyErr <- command.NewExecutor(source, destination).Execute(commands...)
		}()

		go func() {
			saveErr <- repository.Save(&p)
		}()

		if <-copyErr != nil {
			_ = repository.Delete(p)
			logger(errorLabel)
			continue
		}

		if <-saveErr != nil {
			logger(errorLabel)
			continue
		}

		logger(successLabel)
	}
}
