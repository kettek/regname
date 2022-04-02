package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"

	"github.com/martinlindhe/inputbox"
	"github.com/sqweek/dialog"
)

type Config struct {
	Regex   string `json:"regex"`
	Fmt     string `json:"fmt"`
	Lastdir string `json:"lastdir"`
}

func saveConfig(config Config) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile("regname.json", bytes, 0666)
	return err
}

func main() {
	var config Config
	// Create our config.
	bytes, err := os.ReadFile("regname.json")
	if err != nil {
		f, err := os.Create("regname.json")
		if err != nil {
			dialog.Message("%s", err).Error()
			log.Fatal(err)
		}
		f.WriteString(`{
	"regex": "(\\d+)\\s-\\s\\d+\\s([^.]*)(.*)",
	"fmt": "myfile %[2]s %[1]s%[3]s",
	"lastdir": ""
}`)
		f.Close()
		bytes, err = os.ReadFile("regname.json")
		if err != nil {
			dialog.Message("%s", err).Error()
			log.Fatal(err)
		}
	}
	if err := json.Unmarshal(bytes, &config); err != nil {
		dialog.Message("%s", err).Error()
		log.Fatal(err)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		dirpath, err := dialog.Directory().Title("Select a directory to regex rename").SetStartDir(config.Lastdir).Browse()
		if err != nil {
			log.Fatal(errors.New("filepath required"))
		}
		args = append(args, dirpath)
		config.Lastdir = dirpath
	}
	saveConfig(config)

	var ok bool

	config.Regex, ok = inputbox.InputBox("Regular Expression", "Specify the RegEx pattern", config.Regex)
	if !ok {
		log.Fatal("user bailed")
		return
	}
	saveConfig(config)

	re, err := regexp.Compile(config.Regex)
	if err != nil {
		log.Fatal(err)
	}

	config.Fmt, ok = inputbox.InputBox("String Format", "Specify the fmt.Sprintf rename string", config.Fmt)
	if !ok {
		log.Fatal("user bailed")
		return
	}
	saveConfig(config)

	var allFiles []string
	renameMap := make(map[string]string)

	var recurse func(fullpath, localpath string)
	recurse = func(fullpath, localpath string) {
		fpath := path.Join(fullpath, localpath)
		files, err := ioutil.ReadDir(fpath)
		if err != nil {
			log.Println("couldn't read dir", err, fpath)
			return
		}
		for _, file := range files {
			if file.IsDir() {
				recurse(fpath, file.Name())
			} else {
				results := re.FindAllStringSubmatch(file.Name(), -1)
				if len(results) > 0 {
					results := results[0][1:]
					lpath := path.Join(fpath, file.Name())
					var values []interface{}
					for _, v := range results {
						values = append(values, v)
					}
					npath := fmt.Sprintf(config.Fmt, values...)
					renameMap[lpath] = path.Join(fpath, npath)
					allFiles = append(allFiles, lpath)
				}
			}
		}
	}

	// Let's collect our population.
	recurse("", args[0])

	// Let's purge the non-believers.
	var renameString string
	for k, v := range renameMap {
		renameString = renameString + "\n" + k + " => " + v
	}
	ok = dialog.Message("Renaming %d files:\n%s", len(allFiles), renameString).Title("Confirm Rename").YesNo()
	if !ok {
		return
	}
	var errs []error
	var processed []string
	for k, v := range renameMap {
		if err := os.Rename(k, v); err != nil {
			errs = append(errs, err)
		} else {
			processed = append(processed, k+" => "+v)
		}
	}

	if len(errs) > 0 {
		var errorString string
		for _, e := range errs {
			errorString = errorString + "\n" + fmt.Sprintf("%s", e)
		}
		dialog.Message("%s", errorString).Error()
	}
	var processedString string
	for _, s := range processed {
		processedString = processedString + "\n" + s
	}
	dialog.Message("%d out of %d files processed \n%s", len(processed), len(allFiles), processedString).Info()
}
