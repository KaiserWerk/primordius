package primordius

import (
	"context"
	"errors"
	"os"
	"reflect"
	"sync"
	"time"
)

const tagName = "env"

var ErrInvalidSpecification = errors.New("specification must be a struct pointer")

type (
	Source interface {
		ToTarget(t any) error
	}
	Primordius struct {
		target  any
		sources []Source
		m       *sync.RWMutex
		t       *time.Ticker
		ctx     context.Context
		cf      context.CancelFunc
	}
	EnvSource struct {
		prefix string
	}
)

func (es *EnvSource) ToTarget(spec any) error {
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

func New(target any) *Primordius {
	return &Primordius{
		target: target,
		m:      new(sync.RWMutex),
	}
}

func NewWithReload(target any, d time.Duration) *Primordius {
	p := Primordius{
		target: target,
		m:      new(sync.RWMutex),
		t:      time.NewTicker(d),
	}
	p.ctx, p.cf = context.WithCancel(context.Background())

	return &p
}

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

func (pr *Primordius) FromYAMLFile(name string) {}
func (pr *Primordius) FromYAML(content []byte)  {}
func (pr *Primordius) FromJSONFile(name string) {}
func (pr *Primordius) FromJSON(content []byte)  {}
func (pr *Primordius) FromEnv(prefix string) {
	pr.m.Lock()
	pr.sources = append(pr.sources, &EnvSource{prefix: prefix})
	pr.m.Unlock()
}
