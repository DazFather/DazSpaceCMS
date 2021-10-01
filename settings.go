package main

import (
	"encoding/json"
	"os"
	"path"
)

type settings struct {
	ContentDirectoryPath  string `json:content`
	ResourceDirectoryPath string `json:resource`
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

/* The settings directory contains JSON's files that admin can use to edit
 * some server options. Edits will be applied on server reloading
 */
const SETTINGS_DIR = "settings"

// List of all setting's file inside the directory
var (
	PATH_SETTINGS = path.Join(SETTINGS_DIR, "path.JSON")
)

func AddDirPath(directoryName string, folders ...*string) {
	for _, folderPath := range folders {
		*folderPath = path.Join(directoryName, *folderPath)
	}
}

func CreateDefaultSettings() (err error) {
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
	var sets = settings{
		ResourceDirectoryPath: RESOURCE_DIR,
		ContentDirectoryPath:  CONTENT_DIR,
	}
	AddDirPath(CONTENT_DIR, &BLOG_FOLDER, &ARTICLE_FOLDER, &BACKUPS_FOLDER)
	AddDirPath(RESOURCE_DIR, &TEMPLATE_FOLDER, &IMAGES_FOLDER, &STYLES_FOLDER, &SCRIPTS_FOLDER)

	// Create the path settings file
	file, err = os.Create(PATH_SETTINGS)
	if err != nil {
		return
	}

	// Convert to json
	jsonContent, err = json.MarshalIndent(sets, "", "\t")
	if err != nil {
		return
	}

	// Write on file
	_, err = file.Write(jsonContent)
	return err
}

func LoadSettings() (err error) {
	var (
		jsonContent []byte
		sets        settings
	)

	jsonContent, err = os.ReadFile(PATH_SETTINGS)
	if os.IsNotExist(err) {
		err = CreateDefaultSettings()
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(jsonContent, &sets)
	if err != nil {
		return
	}

	CONTENT_DIR = sets.ContentDirectoryPath
	AddDirPath(CONTENT_DIR, &BLOG_FOLDER, &ARTICLE_FOLDER, &BACKUPS_FOLDER)
	RESOURCE_DIR = sets.ResourceDirectoryPath
	AddDirPath(RESOURCE_DIR, &TEMPLATE_FOLDER, &IMAGES_FOLDER, &STYLES_FOLDER, &SCRIPTS_FOLDER)

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
				return
			}
		} else if err != nil {
			return
		}
	}

	return
}
