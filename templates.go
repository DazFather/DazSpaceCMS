package main

import (
	"html/template"
	"io"
	"os"
	"path"
)

type Data struct {
	StylesPath, ScriptPath string
	Value                  interface{}
}

var (
	SavedTemplates  *template.Template
	TEMPLATE_FOLDER = path.Join(RESOURCE_DIR, "templates")
)

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

	SavedTemplates = template.Must(template.ParseFiles(names...))
	return nil
}

func Pack(obj interface{}) Data {
	return Data{
		StylesPath: "/" + STYLES_FOLDER + "/",
		ScriptPath: "/" + SCRIPTS_FOLDER + "/",
		Value:      obj,
	}
}

func Compose(w io.Writer, templateName string, rawDatas interface{}) error {
	return SavedTemplates.ExecuteTemplate(w, templateName, Pack(rawDatas))
}
