package tests_test

import (
	"fmt"
	"github.com/jmhbh/argo-app-orchestrator/app_orchestrator/utils"
	"os"
	"testing"
)

func TestTemplateYaml(t *testing.T) {
	yaml, err := os.ReadFile("testdata/testappset.yaml")
	if err != nil {
		t.Errorf("error reading file: %v", err)
	}

	userNames := []string{"luigi", "bowser jr.", "yoshi"}
	buf, err := utils.TemplateYaml(userNames, yaml)
	if err != nil {
		t.Errorf("error templating yaml: %v", err)
	}
	fmt.Printf("succeeded in templating yaml output - %s\n", buf.String())
}
