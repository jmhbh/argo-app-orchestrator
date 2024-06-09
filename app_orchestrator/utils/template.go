package utils

import (
	"bytes"
	"html/template"
)

func TemplateYaml(userNames []string, yaml []byte) (*bytes.Buffer, error) {
	temp := template.Must(template.New("yaml").Parse(string(yaml)))
	buffer := new(bytes.Buffer)
	err := temp.Execute(buffer, map[string]interface{}{"UserNames": userNames, "AppName": "{{user}}-app"})
	if err != nil {
		return nil, err
	}
	return buffer, nil
}
