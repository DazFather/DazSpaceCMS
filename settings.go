package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
)

/* The settings directory contains JSON's files that admin can use to edit
* some server options. Edits will be applied on server reloading
 */
const SETTINGS_DIR = "settings"

type setting interface {
	CreateDefault() error
	Load() error
}

func LoadSettings() (err error) {
	var path pathSettings

	err = path.Load()
	if os.IsNotExist(err) {
		err = path.CreateDefault()
	}
	if err != nil {
		return
	}

	err = SITE.Load()
	if os.IsNotExist(err) {
		err = SITE.CreateDefault()
	}
	if err != nil {
		return
	}

	return
}

func HealDirectories() (err error) {
	var allPaths = []string{
		RESOURCE_DIR,
		TEMPLATE_FOLDER, IMAGES_FOLDER, STYLES_FOLDER, SCRIPTS_FOLDER,
		CONTENT_DIR,
		BLOG_FOLDER, ARTICLE_FOLDER, BACKUPS_FOLDER,
	}

	for _, path := range allPaths {
		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			err = os.Mkdir(path, 0755)
			if err != nil {
				return errors.New(fmt.Sprint(err, "\"", path, "\""))
			}
		} else if err != nil {
			return
		}
	}

	if _, err = os.Stat(SITE_SETTINGS); os.IsNotExist(err) {
		file, e := os.Create(SITE_SETTINGS)
		defer file.Close()
		if e != nil {
			return e
		}
	}

	return
}

/* --- PATH SETTINGS --- */
var PATH_SETTINGS = path.Join(SETTINGS_DIR, "path.JSON")

type pathSettings struct {
	ContentDirectoryPath  string `json:"content"`
	ResourceDirectoryPath string `json:"resource"`
}

/* The resource directory holds all the file that will be served statically.
 * Doing so they can be easily be incorporated in the articles but it expose
 * the directory to the users that will be able to access every file or folder.
 * Be careful of what you put there
 */
var (
	RESOURCE_DIR = "resources"
	// folders inside the Resources directory
	TEMPLATE_FOLDER = "templates"
	IMAGES_FOLDER   = "imgs"
	STYLES_FOLDER   = "styles"
	SCRIPTS_FOLDER  = "scripts"
)

/* The content directory holds all the files that will be dynamically severd
 * This means that user can not have access to it but this also means that
 * also front-end can't incorporate stuffs and is the server that needs to
 * take care of that
 */
var (
	CONTENT_DIR = "contents"
	// folders inside the Contents directory
	BLOG_FOLDER    = "blog"
	ARTICLE_FOLDER = "articles"
	BACKUPS_FOLDER = "backups"
)

func AddDirPath(directoryName string, folders ...*string) {
	for _, folderPath := range folders {
		*folderPath = path.Join(directoryName, *folderPath)
	}
}

func (s pathSettings) CreateDefault() (err error) {
	var (
		file        *os.File
		jsonContent []byte
	)

	if _, e := os.Stat(SETTINGS_DIR); os.IsNotExist(e) {
		err = os.Mkdir(SETTINGS_DIR, 0755)
		if err != nil {
			return
		}
	}

	// Set the default paths values
	s.ResourceDirectoryPath = RESOURCE_DIR
	s.ContentDirectoryPath = CONTENT_DIR

	AddDirPath(CONTENT_DIR, &BLOG_FOLDER, &ARTICLE_FOLDER, &BACKUPS_FOLDER)
	AddDirPath(RESOURCE_DIR, &TEMPLATE_FOLDER, &IMAGES_FOLDER, &STYLES_FOLDER, &SCRIPTS_FOLDER)

	// Create the path settings file
	file, err = os.Create(PATH_SETTINGS)
	if err != nil {
		return
	}

	// Convert to json
	jsonContent, err = json.MarshalIndent(s, "", "\t")
	if err != nil {
		return
	}

	// Write on file
	_, err = file.Write(jsonContent)
	return err
}

func (s *pathSettings) Load() error {
	var jsonContent, err = os.ReadFile(PATH_SETTINGS)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonContent, &s)
	if err != nil {
		return err
	}

	CONTENT_DIR = s.ContentDirectoryPath
	AddDirPath(CONTENT_DIR, &BLOG_FOLDER, &ARTICLE_FOLDER, &BACKUPS_FOLDER)
	RESOURCE_DIR = s.ResourceDirectoryPath
	AddDirPath(RESOURCE_DIR, &TEMPLATE_FOLDER, &IMAGES_FOLDER, &STYLES_FOLDER, &SCRIPTS_FOLDER)

	return err
}

/* --- SITE SETTINGS --- */
var (
	SITE_SETTINGS = path.Join(SETTINGS_DIR, "site.JSON")
	SITE          SiteSettings
)

type SiteSettings struct {
	Domain   string `json:"domain"`
	Name     string `json:"name"`
	About    string `json:"about"`
	Language string `json:"language"`
	Owner    string `json:"owner"`
	Mail     string `json:"mail"`
}

func (s *SiteSettings) CreateDefault() (err error) {
	err = errors.New("Missing site info")
	return
}

func (s *SiteSettings) Load() (err error) {
	var jsonContent []byte

	jsonContent, err = os.ReadFile(SITE_SETTINGS)
	if err != nil {
		return
	}

	err = json.Unmarshal(jsonContent, &s)
	if err != nil {
		return
	}

	return
}
