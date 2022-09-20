////go:build integration

package primordius

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_yamlFileSource_ToTarget(t *testing.T) {
	const path = "./Test_yamlFileSource_ToTarget"
	if err := os.MkdirAll(path, 0777); err != nil {
		t.Fatalf("failed to create temp dir: %s", err.Error())
	}
	defer os.RemoveAll(path)

	tests := []struct {
		name     string
		content  []byte
		filename string
		source   Source
		target   interface{}
		want     *testTarget
		wantErr  bool
	}{
		{
			"YAML file with 0 entries",
			[]byte(``),
			filepath.Join(path, "0.yaml"),
			&yamlFileSource{name: filepath.Join(path, "0.yaml")},
			&testTarget{},
			&testTarget{},
			false,
		},
		{
			"YAML file with 1 valid entry",
			[]byte(`a: hallo`),
			filepath.Join(path, "1.yaml"),
			&yamlFileSource{name: filepath.Join(path, "1.yaml")},
			&testTarget{},
			&testTarget{a: "hello"},
			false,
		},
		{
			"YAML file with 2 valid entries",
			[]byte(`a: hallo\nb:bye`),
			filepath.Join(path, "2.yaml"),
			&yamlFileSource{name: filepath.Join(path, "2.yaml")},
			&testTarget{},
			&testTarget{a: "hello", b: "bye"},
			false,
		},
		{
			"YAML file with 3 valid entries",
			[]byte(`a: hallo\nb:bye\nc:"how is it going?"`),
			filepath.Join(path, "3.yaml"),
			&yamlFileSource{name: filepath.Join(path, "3.yaml")},
			&testTarget{},
			&testTarget{a: "hello", b: "bye", c: "how is it going?"},
			false,
		},
		{
			"YAML file with 2 valid and 1 invalid entry",
			[]byte(`a: hallo\nb:bye\nc: Go: The Easy Way`),
			filepath.Join(path, "3.yaml"),
			&yamlFileSource{name: filepath.Join(path, "3.yaml")},
			&testTarget{},
			nil,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			if err := os.WriteFile(tc.filename, tc.content, 0777); err != nil {
				t.Fatalf("failed to write test file: %s", err.Error())
			}

			if err := tc.source.ToTarget(tc.target); (err != nil) != tc.wantErr {
				t.Errorf("ToTarget() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
