package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	var out bytes.Buffer

	if err := run(&out); err != nil {
		t.Fatalf("run() retornou erro: %v", err)
	}

	if !strings.Contains(out.String(), serviceName) {
		t.Errorf("saída %q não contém o nome do serviço %q", out.String(), serviceName)
	}
}
