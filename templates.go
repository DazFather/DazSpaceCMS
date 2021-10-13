package main

import (
	"html/template"
	"io"
	"os"
	"path"
)

type Data struct {
	Site                   SiteSettings
	StylesPath, ScriptPath string
	Value                  interface{}
}

func LoadTemplates(folderPath string) error {
	var (
		names      []string
		files, err = os.ReadDir(folderPath)
	)
	if err != nil {
		return err
	}

	for _, file := range files {
		names = append(names, path.Join(folderPath, file.Name()))
	}

	Cache.SavedTemplates = template.Must(template.ParseFiles(names...))
	return nil
}

func Pack(obj interface{}) Data {
	return Data{
		Site:       SITE,
		StylesPath: "/" + STYLES_FOLDER + "/",
		ScriptPath: "/" + SCRIPTS_FOLDER + "/",
		Value:      obj,
	}
}

func Compose(w io.Writer, templateName string, rawDatas interface{}) error {
	return Cache.SavedTemplates.ExecuteTemplate(w, templateName, Pack(rawDatas))
}
