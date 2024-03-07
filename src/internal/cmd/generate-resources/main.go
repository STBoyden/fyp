//go:build ignore
// +build ignore

package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"
)

var timestamp = time.Now()

func generateImageResourcesFile() {
	type templateStruct struct {
		Timestamp               time.Time
		ImageFiles              []string
		ImageFilesFunctionified []string
	}

	t := template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
/*
This file was generated at {{ .Timestamp }}.
*/

package resources

import (
	"image"
	_ "image/png"
	_ "image/jpeg"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

{{range $i, $n := .ImageFilesFunctionified }}
/*
GetImg{{$n}} loads "resources/images/{{ index $.ImageFiles $i }}" as an ebiten-compatible image.
*/
func GetImg{{$n}}() (*ebiten.Image, image.Image, error) {
	return ebitenutil.NewImageFromFile("resources/images/{{ index $.ImageFiles $i }}")
}
{{end}}
`))

	entries, err := os.ReadDir("./images")
	if err != nil {
		log.Fatalf("an error occurred: %s", err.Error())
	}

	sourceFile, err := os.Create("resources_images.go")
	if err != nil {
		log.Fatalf("an error occurred: %s", err.Error())
	}
	defer sourceFile.Close()

	fileNames := []string{}
	fileNamesFunctionified := []string{}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		fileNames = append(fileNames, entry.Name())

		functionified := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		capitalise := func(words []string) string {
			for i, word := range words {
				firstRune, size := utf8.DecodeRuneInString(word)
				if firstRune == utf8.RuneError && size < 2 {
					continue
				}

				firstLetter := string(firstRune)
				after := string([]rune(word)[1:])

				words[i] = strings.ToUpper(firstLetter) + after

			}

			return strings.Join(words, "")
		}

		functionified = capitalise([]string{functionified})
		functionified = capitalise(strings.Split(functionified, " "))
		functionified = capitalise(strings.Split(functionified, "_"))
		functionified = capitalise(strings.Split(functionified, "-"))

		fileNamesFunctionified = append(fileNamesFunctionified, functionified)
	}

	err = t.Execute(sourceFile, templateStruct{
		Timestamp:               timestamp,
		ImageFiles:              fileNames,
		ImageFilesFunctionified: fileNamesFunctionified,
	})
	if err != nil {
		log.Printf("an error occurred: %s", err.Error())
	}
}

func generateSfxResourcesFile() {
	type templateStruct struct {
		Timestamp               time.Time
		SoundFiles              []string
		SoundFilesFunctionified []string
	}

	t := template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
/*
This file was generated at {{ .Timestamp }}.
*/

package resources

import (
	"os"

	_ "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

{{range $i, $n := .SoundFilesFunctionified }}
/*
GetSfx{{$n}} decodes "resources/sfx/{{ index $.SoundFiles $i }}" as playable in an ebiten audio.Context.
*/
func GetSfx{{$n}}() (*wav.Stream, error) {
	file, err := os.Open("resources/sfx/{{ index $.SoundFiles $i }}")
	if err != nil {
		return nil, err
	}

	stream, err := wav.DecodeWithoutResampling(file)
	if err != nil {
		return nil, err
	}

	return stream, err
}
{{end}}
`))

	entries, err := os.ReadDir("./sfx")
	if err != nil {
		log.Fatalf("an error occurred: %s", err.Error())
	}

	sourceFile, err := os.Create("resources_sfx.go")
	if err != nil {
		log.Fatalf("an error occurred: %s", err.Error())
	}
	defer sourceFile.Close()

	fileNames := []string{}
	fileNamesFunctionified := []string{}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		fileNames = append(fileNames, entry.Name())

		functionified := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		capitalise := func(words []string) string {
			for i, word := range words {
				firstRune, size := utf8.DecodeRuneInString(word)
				if firstRune == utf8.RuneError && size < 2 {
					continue
				}

				firstLetter := string(firstRune)
				after := string([]rune(word)[1:])

				words[i] = strings.ToUpper(firstLetter) + after

			}

			return strings.Join(words, "")
		}

		functionified = capitalise([]string{functionified})
		functionified = capitalise(strings.Split(functionified, " "))
		functionified = capitalise(strings.Split(functionified, "_"))
		functionified = capitalise(strings.Split(functionified, "-"))

		fileNamesFunctionified = append(fileNamesFunctionified, functionified)
	}

	err = t.Execute(sourceFile, templateStruct{
		Timestamp:               timestamp,
		SoundFiles:              fileNames,
		SoundFilesFunctionified: fileNamesFunctionified,
	})
	if err != nil {
		log.Printf("an error occurred: %s", err.Error())
	}
}

func generateMusicResourcesFile() {
	type templateStruct struct {
		Timestamp               time.Time
		MusicFiles              []string
		MusicFilesFunctionified []string
	}

	t := template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
/*
This file was generated at {{ .Timestamp }}.
*/

package resources

import (
	"os"

	_ "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

{{range $i, $n := .MusicFilesFunctionified }}
/*
GetMusic{{$n}} decodes "resources/music/{{ index $.MusicFiles $i }}" as playable in an ebiten audio.Context.
*/
func GetMusic{{$n}}() (*vorbis.Stream, error) {
	file, err := os.Open("resources/music/{{ index $.MusicFiles $i }}")
	if err != nil {
		return nil, err
	}

	stream, err := vorbis.DecodeWithoutResampling(file)
	if err != nil {
		return nil, err
	}

	return stream, err
}
{{end}}
`))

	entries, err := os.ReadDir("./music")
	if err != nil {
		log.Fatalf("an error occurred: %s", err.Error())
	}

	sourceFile, err := os.Create("resources_music.go")
	if err != nil {
		log.Fatalf("an error occurred: %s", err.Error())
	}
	defer sourceFile.Close()

	fileNames := []string{}
	fileNamesFunctionified := []string{}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		fileNames = append(fileNames, entry.Name())

		functionified := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		capitalise := func(words []string) string {
			for i, word := range words {
				firstRune, size := utf8.DecodeRuneInString(word)
				if firstRune == utf8.RuneError && size < 2 {
					continue
				}

				firstLetter := string(firstRune)
				after := string([]rune(word)[1:])

				words[i] = strings.ToUpper(firstLetter) + after

			}

			return strings.Join(words, "")
		}

		functionified = capitalise([]string{functionified})
		functionified = capitalise(strings.Split(functionified, " "))
		functionified = capitalise(strings.Split(functionified, "_"))
		functionified = capitalise(strings.Split(functionified, "-"))

		fileNamesFunctionified = append(fileNamesFunctionified, functionified)
	}

	err = t.Execute(sourceFile, templateStruct{
		Timestamp:               timestamp,
		MusicFiles:              fileNames,
		MusicFilesFunctionified: fileNamesFunctionified,
	})
	if err != nil {
		log.Printf("an error occurred: %s", err.Error())
	}
}

func main() {
	generateImageResourcesFile()
	generateSfxResourcesFile()
	generateMusicResourcesFile()
}