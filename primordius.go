package primordius

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strconv"
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

		if !f.IsValid() {
			continue
		}
		val, exists := os.LookupEnv(es.prefix + t.Field(i).Tag.Get(tagName))
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
		m:      new(sync.RWMutex),
		cf:     func() {}, // make sure cf is never nil to prevent panic
	}
}

// NewWithReload is like New, but sets up an interval at which the configuration is re-read into target.
func NewWithReload(target any, d time.Duration) *Primordius {
	var ctx context.Context
	p := &Primordius{
		target: target,
		m:      new(sync.RWMutex),
		t:      time.NewTicker(d),
	}
	ctx, p.cf = context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.t.C:
				_ = p.Process()
			}
		}
	}()
	return p
}

// Stop stops the automatic configuration reloading started via NewWithReload.
// Calls to Stop for instances created via New() or repeated calls do nothing.
func (pr *Primordius) Stop() {
	pr.cf()
}

// Process calls all registered Sources to write values into pr.target.
// Registered sources are processed in the order they were initially added.
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
// Can also be used to add a custom Source.
func (pr *Primordius) AddSource(s Source) {
	pr.m.Lock()
	pr.sources = append(pr.sources, s)
	pr.m.Unlock()
}

// ResetSources empties the internal list of registered Sources.
func (pr *Primordius) ResetSources() {
	pr.m.Lock()
	pr.sources = make([]Source, 0, 5)
	pr.m.Unlock()
}
