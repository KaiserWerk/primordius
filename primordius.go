package primordius

import (
	"encoding/json"
	"errors"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"reflect"
	"strconv"
)

const tagName = "env"

var ErrInvalidSpecification = errors.New("specification must be a struct pointer")

type (
	// Source defines an origin writing found configuration values into t.
	Source interface {
		// ToTarget writes configuration values into t. t MUST be a pointer to a struct.
		ToTarget(t any) error
	}
	// Primordius manages sources and processes them into the set target.
	Primordius struct {
		target  any
		sources []Source
	}
	yamlFileSource struct {
		name string
	}
	yamlContentSource struct {
		content []byte
	}
	yamlReaderSource struct {
		r io.Reader
	}
	jsonFileSource struct {
		name string
	}
	jsonContentSource struct {
		content []byte
	}
	jsonReaderSource struct {
		r io.Reader
	}
	tomlFileSource struct {
		name string
	}
	tomlContentSource struct {
		content []byte
	}
	tomlReaderSource struct {
		r io.Reader
	}
	envSource struct {
		prefix string
	}
)

func (y *yamlFileSource) ToTarget(t any) error {
	cont, err := os.ReadFile(y.name)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(cont, t)
}

func (y *yamlContentSource) ToTarget(t any) error {
	return yaml.Unmarshal(y.content, t)
}

func (y *yamlReaderSource) ToTarget(t any) error {
	cont, err := io.ReadAll(y.r)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(cont, t)
}

func (j *jsonFileSource) ToTarget(t any) error {
	cont, err := os.ReadFile(j.name)
	if err != nil {
		return err
	}

	return json.Unmarshal(cont, t)
}

func (j *jsonContentSource) ToTarget(t any) error {
	return json.Unmarshal(j.content, t)
}

func (y *jsonReaderSource) ToTarget(t any) error {
	cont, err := io.ReadAll(y.r)
	if err != nil {
		return err
	}

	return json.Unmarshal(cont, t)
}

func (to *tomlFileSource) ToTarget(t any) error {
	_, err := toml.DecodeFile(to.name, t)
	return err
}

func (to *tomlContentSource) ToTarget(t any) error {
	_, err := toml.Decode(string(to.content), t)
	return err
}

func (to *tomlReaderSource) ToTarget(t any) error {
	cont, err := io.ReadAll(to.r)
	if err != nil {
		return err
	}

	_, err = toml.Decode(string(cont), t)
	return err
}

func (es *envSource) ToTarget(spec any) error {
	valueOf := reflect.ValueOf(spec)

	if valueOf.Kind() != reflect.Pointer {
		return ErrInvalidSpecification
	}
	s := valueOf.Elem()
	if s.Kind() != reflect.Struct {
		return ErrInvalidSpecification
	}

	t := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		if !f.IsValid() {
			continue
		}
		tagVal := t.Field(i).Tag.Get(tagName)
		if tagVal == "" || tagVal == "-" {
			continue
		}
		val, exists := os.LookupEnv(es.prefix + tagVal)
		if !exists {
			continue
		}

		switch f.Kind() {
		case reflect.String:
			f.SetString(val)
		case reflect.Int:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			f.SetInt(v)
		case reflect.Uint:
			fallthrough
		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			v, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}
			f.SetUint(v)
		case reflect.Bool:
			v, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			f.SetBool(v)
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return err
			}
			f.SetFloat(v)
		case reflect.Slice:
			f.SetBytes([]byte(val))
		}

	}

	return nil
}

// New allocates and returns a new instance of Primordius with the supplied target.
// target MUST be a pointer to a struct.
func New(target any) *Primordius {
	return &Primordius{
		target: target,
	}
}

// Process calls all registered Sources to write values into pr.target.
// Registered sources are processed in the order they were initially added.
func (pr *Primordius) Process() error {
	for _, s := range pr.sources {
		if err := s.ToTarget(pr.target); err != nil {
			return err
		}
	}

	return nil
}

// FromYAMLFile adds a Source to pr which reads values from a YAML file.
func (pr *Primordius) FromYAMLFile(name string) {
	pr.AddSource(&yamlFileSource{name: name})
}

// FromYAML adds a Source to pr which reads values from a YAML block.
func (pr *Primordius) FromYAML(content []byte) {
	pr.AddSource(&yamlContentSource{content: content})
}

// FromYAMLReader adds a Source to pr which reads JSON content from r.
func (pr *Primordius) FromYAMLReader(r io.Reader) {
	pr.AddSource(&yamlReaderSource{r: r})
}

// FromJSONFile adds a Source to pr which reads values from a JSON file.
func (pr *Primordius) FromJSONFile(name string) {
	pr.AddSource(&jsonFileSource{name: name})
}

// FromJSON adds a Source to pr which reads values from a JSON block.
func (pr *Primordius) FromJSON(content []byte) {
	pr.AddSource(&jsonContentSource{content: content})
}

// FromJSONReader adds a Source to pr which reads YAML content from r.
func (pr *Primordius) FromJSONReader(r io.Reader) {
	pr.AddSource(&jsonReaderSource{r: r})
}

func (pr *Primordius) FromTOMLFile(name string) {
	pr.AddSource(&tomlFileSource{name: name})
}

func (pr *Primordius) FromTOML(content []byte) {
	pr.AddSource(&tomlContentSource{content: content})
}

func (pr *Primordius) FromTOMLReader(r io.Reader) {
	pr.AddSource(&tomlReaderSource{r: r})
}

// FromEnv adds a Source to pr which reads values from environment variables.
func (pr *Primordius) FromEnv(prefix string) {
	pr.AddSource(&envSource{prefix: prefix})
}

// AddSource adds a Source s to pr to obtain arbitrary configuration values from.
// Can also be used to add a custom Source.
func (pr *Primordius) AddSource(s Source) {
	pr.sources = append(pr.sources, s)
}

// ResetSources empties the internal list of registered Sources.
func (pr *Primordius) ResetSources() {
	pr.sources = make([]Source, 0, 5)
}
