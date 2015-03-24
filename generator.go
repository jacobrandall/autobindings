package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"io/ioutil"
	"log"
	"strings"
	"text/template"

	"go/parser"
	"go/token"
	"os"
)

type Mapping struct {
	Name           string
	Type           string
	JSONTags       string
	RestrictedTags string
	Required       bool
	Create         bool
	Update         bool
	MediaTypes     []string
}

func (m *Mapping) HasMediaType(mediaType string) bool {
	lm := strings.ToLower(mediaType)
	for _, c := range m.MediaTypes {
		if strings.ToLower(c) == lm {
			return true
		}
	}
	return false
}

var AllMediaTypes = []string{"audio", "subtitle", "video"}
var AllMediaTypesCaps = []string{"Audio", "Subtitle", "Video"}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		log.Println("Usage : autobindings {file_name} \n example: autobindings file.go")
		return
	}
	log.Println("autobindings started. file:", args[0])
	generateFieldMap(args[0])
	log.Println("autobindings finished. file:", args[0])
}

func generateFieldMap(fileName string) {
	fset := token.NewFileSet() // positions are relative to fset
	// Parse the file given in arguments
	f, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	structMap := map[string]*ast.FieldList{}
	// range over the structs and fill struct map
	for _, d := range f.Scope.Objects {
		ts, ok := d.Decl.(*ast.TypeSpec)
		if !ok {
			continue
		}
		switch ts.Type.(type) {
		case *ast.StructType:
			x, _ := ts.Type.(*ast.StructType)
			structMap[ts.Name.String()] = x.Fields
		}
	}
	// looping through each struct and creating a bindings file for it
	packageName := f.Name
	for structName, fields := range structMap {
		variableName := strings.ToLower(string(structName[0]))
		mappings := map[string]*Mapping{}
		for _, field := range fields.List {
			if len(field.Names) == 0 {
				continue
			}
			name := field.Names[0].String()
			fieldType := string(b[field.Type.Pos()-1 : field.Type.End()-1])
			mappings[name] = &Mapping{Type: fieldType, MediaTypes: AllMediaTypes}
			// if tag for field doesn't exists, create one
			if field.Tag == nil {
				mappings[name].JSONTags = name
				mappings[name].RestrictedTags = fmt.Sprintf("`json:\"%s\"`", name)
			} else if strings.Contains(field.Tag.Value, "json") || strings.Contains(field.Tag.Value, "binding") || strings.Contains(field.Tag.Value, "mediatypes") {
				tags := strings.Replace(field.Tag.Value, "`", "", -1)
				for _, tag := range strings.Split(tags, " ") {
					if strings.Contains(tag, "json") {
						mapping := strings.Replace(tag, "json:\"", "", -1)
						mapping = strings.Replace(mapping, "\"", "", -1)
						if mapping == "-" {
							continue
						}
						mappings[name].JSONTags = strings.TrimSuffix(mapping, ",omitempty")
						mappings[name].RestrictedTags = fmt.Sprintf("`json:\"%s\"`", mapping)
					} else if strings.Contains(tag, "binding") {
						mapping := strings.Replace(tag, "binding:\"", "", -1)
						mapping = strings.Replace(mapping, "\"", "", -1)
						if strings.Contains(mapping, "required") {
							mappings[name].Required = true
						}
						if strings.Contains(mapping, "create") {
							mappings[name].Create = true
						}
						if strings.Contains(mapping, "update") {
							mappings[name].Update = true
						}
					} else if strings.Contains(tag, "mediatypes") {
						mapping := strings.Replace(tag, "mediatypes:\"", "", -1)
						mapping = strings.Replace(mapping, "\"", "", -1)
						mediaTypes := strings.Split(mapping, ",")
						mappings[name].MediaTypes = make([]string, 0, 3)
						mappings[name].MediaTypes = append(mappings[name].MediaTypes, mediaTypes...)
					}
				}
			} else {
				// I will handle other cases later
				mappings[name].JSONTags = name
				mappings[name].RestrictedTags = fmt.Sprintf("`json:\"%s\"`", name)
			}
		}
		// opening file for writing content
		splitFileName := strings.Split(fileName, ".")
		bindingsFileName := fmt.Sprintf("%s_bindings.%s", splitFileName[0], splitFileName[1])
		log.Printf("Proccessed %s, writing bindings out to %s", fileName, bindingsFileName)
		writer, err := os.Create(bindingsFileName)
		if err != nil {
			log.Printf("Error opening file %v", err)
			panic(err)
		}
		defer writer.Close()
		content := new(bytes.Buffer)

		var bindingsFile string
		if structName == "Asset" {
			bindingsFile = bindingsFileAsset
		} else {
			bindingsFile = bindingsFileGeneric
		}
		t := template.Must(template.New("bindings").Parse(bindingsFile))
		err = t.Execute(content, map[string]interface{}{
			"packageName":  packageName,
			"variableName": variableName,
			"structName":   structName,
			"mappings":     mappings,
			"mediaTypes":   AllMediaTypesCaps,
		})
		if err != nil {
			panic(err)
		}
		finalContent, err := format.Source(content.Bytes())
		if err != nil {
			panic(err)
		}
		writer.WriteString(string(finalContent))
	}
}
