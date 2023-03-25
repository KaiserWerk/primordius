# primordius
Primordius is a Go library to read configuration values from different sources (YAML, JSON, TOML, env vars).

## Installation

Just your typical ``go get github.com/KaiserWerk/primordius``.

## Usage

First, define your configuration struct. Can be any flat struct you like, e.g.

```golang
type Config struct {
    BaseURL      string `json:"base_url" yaml:"base_url" env:"PROXY_BASEURL"`
    Key          string `json:"key" yaml:"key" env:"KEY"`
    NumBackups   uint   `yaml:"num_backups" env:"NUM_BACKUPS"`
    ProxyEnabled bool   `env:"PROXY_ENABLED"`
}
```

Make sure you define the tags as required. Use the tag ``env`` for values that should be
read from environment variables.

Then, create an instance of your configuration struct and maybe set some default values: 

```golang
c := Config{
    BaseURL: "https://my.greenhouse.lan/",
    Key:     "mysupersecretkey",
}
```

Then, a new Primordius instance:

```golang
pr := primordius.New(&c)
```

It is important to note that you MUST supply a pointer to a struct as target.

The next step is to set up the desired sources. There are some default sources you can add directly
on the Primordius struct:

```golang
// Reads from a JSON block, maybe obtained by an external service
pr.FromJSON([]byte(`{"key": "some-key"}`))
// Reads from the supplied file
pr.FromJSONFile(`C:\Local\app.prod.json`)
// Reads from an io.Reader
pr.FromJSONReader(r.Body)
// Reads from a YAML block, maybe obtained by an external service
pr.FromYAML([]byte(`num_backups: 16`))
// Reads from the supplied file
pr.FromYAMLFile(`/opt/local/app.yaml`)
// Reads from an io.Reader
pr.FromYAMLReader(strings.NewReader(`base_url: "http://some-url"`))
// Reads from a YAML block, maybe obtained by an external service
pr.FromTOML([]byte(`data = [ ["delta", "phi"], [3.14] ]`))
// Reads from the supplied file
pr.FromTOMLFile("C:\\Users\\SomeUser\\AppData\\Local\\my-app\\config.prod.toml")
// Reads from an io.Reader
pr.FromTOMLReader(resp.Body)
// Reads from the env vars defined in the 'env' tag combined with the supplied
// prefix. If you don't need a prefix, supply an empty string.
pr.FromEnv("MY_APP_")
```

Sources are processed in the order they were registered meaning the last source has the highest
priority.

Lastly, call ``pr.Process()``:

```golang
err := pr.Process()
if err != nil {
    log.Fatal(err)
}
```

Now your configuration is populated with the values read from the sources and ready to be used!

### Custom sources

You have a different resource you want to read configuration values from? 
All sources must implement the ``primordius.Source`` interface which defines just
a single method:

```golang
// Source defines an origin writing found configuration values into t.
Source interface {
    // ToTarget writes configuration values into t. t MUST be a pointer to a struct.
    ToTarget(t any) error
}
```

You can add your custom ``Source`` like this:

```golang
type mySource struct {}
func (ms *mySource) ToTarget(target any) error { /* TODO implement */ }
s := &mySource{}
pr.AddSource(s)
```