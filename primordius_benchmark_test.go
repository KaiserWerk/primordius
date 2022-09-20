package primordius

import "testing"

func Benchmark_EnvSource_ToTarget(b *testing.B) {
	tests := []struct {
		name   string
		source Source
		target interface{}
		env    map[string]string
	}{
		{
			"4 fields / 0 env vars",
			&envSource{},
			&testTarget{},
			map[string]string{},
		},
		{
			"4 fields, 1 env var",
			&envSource{},
			&testTarget{},
			map[string]string{"a": "abc"},
		},
		{
			"4 fields, 2 env var",
			&envSource{},
			&testTarget{},
			map[string]string{"a": "abc", "b": "abc"},
		},
		{
			"4 fields, 3 env var",
			&envSource{},
			&testTarget{},
			map[string]string{"a": "abc", "b": "abc", "c": "abc"},
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tc.source.ToTarget(tc.target)
			}
		})
	}
}
