package primordius

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
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
		m       *sync.RWMutex
		t       *time.Ticker
		ctx     context.Context
		cf      context.CancelFunc
	}
	yamlFileSource struct {
		name string
	}
	yamlContentSource struct {
		content []byte
	}
	jsonFileSource struct {
		name string
	}
	jsonContentSource struct {
		content []byte
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
		//fmt.Printf("%d: %s %s = %v\n", i, t.Field(i).Name, f.Type(), f.Interface())
		//tag := f.Tag.Get(tagName)

		if f.IsValid() {
			f.SetString(os.Getenv(es.prefix + t.Field(i).Tag.Get(tagName)))
		}
	}

	return nil
}

// New allocates and returns a new instance of Primordius with the supplied target.
// target MUST be a pointer to a struct.
func New(target any) *Primordius {
	return &Primordius{
		target: target,
		m:      new(sync.RWMutex),
	}
}

// NewWithReload is like new, but sets up an interval at which the configuration is re-read into target.
func NewWithReload(target any, d time.Duration) *Primordius {
	p := Primordius{
		target: target,
		m:      new(sync.RWMutex),
		t:      time.NewTicker(d),
	}
	p.ctx, p.cf = context.WithCancel(context.Background())

	return &p
}

// Process calls all registered Sources to write values into pr.target.
func (pr *Primordius) Process() error {
	pr.m.Lock()
	defer pr.m.Unlock()

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

// FromJSONFile adds a Source to pr which reads values from a JSON file.
func (pr *Primordius) FromJSONFile(name string) {
	pr.AddSource(&jsonFileSource{name: name})
}

// FromJSON adds a Source to pr which reads values from a JSON block.
func (pr *Primordius) FromJSON(content []byte) {
	pr.AddSource(&jsonContentSource{content: content})
}

// FromEnv adds a Source to pr which reads values from environment variables.
func (pr *Primordius) FromEnv(prefix string) {
	pr.AddSource(&envSource{prefix: prefix})
}

// AddSource adds a Source to to pr to obtain arbitrary configuration values from.
func (pr *Primordius) AddSource(s Source) {
	pr.m.Lock()
	pr.sources = append(pr.sources, s)
	pr.m.Unlock()
}
