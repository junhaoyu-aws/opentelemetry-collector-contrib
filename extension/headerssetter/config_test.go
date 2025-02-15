// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package headerssetter

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id       config.ComponentID
		expected config.Extension
	}{
		{
			id:       config.NewComponentID(typeStr),
			expected: NewFactory().CreateDefaultConfig(),
		},
		{
			id: config.NewComponentIDWithName(typeStr, "1"),
			expected: &Config{
				ExtensionSettings: config.NewExtensionSettings(config.NewComponentID(typeStr)),
				HeadersConfig: []HeaderConfig{
					{
						Key:         stringp("X-Scope-OrgID"),
						FromContext: stringp("tenant_id"),
						Value:       nil,
					},
					{
						Key:         stringp("User-ID"),
						FromContext: stringp("user_id"),
						Value:       nil,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			require.NoError(t, err)
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()
			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, config.UnmarshalExtension(sub, cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		header      []HeaderConfig
		expectedErr error
	}{
		{
			"header value from config property",
			[]HeaderConfig{
				{
					Key:   stringp("name"),
					Value: stringp("from config"),
				},
			},
			nil,
		},
		{
			"header value from context",
			[]HeaderConfig{
				{
					Key:         stringp("name"),
					FromContext: stringp("from config"),
				},
			},
			nil,
		},
		{
			"missing header name for from value",
			[]HeaderConfig{
				{Value: stringp("test")},
			},
			errMissingHeader,
		},
		{
			"missing header name for from context",
			[]HeaderConfig{
				{FromContext: stringp("test")},
			},
			errMissingHeader,
		},
		{
			"header value from context and value",
			[]HeaderConfig{
				{
					Key:         stringp("name"),
					Value:       stringp("from config"),
					FromContext: stringp("from context"),
				},
			},
			errConflictingSources,
		},
		{
			"header value source is missing",
			[]HeaderConfig{
				{
					Key: stringp("name"),
				},
			},
			errMissingSource,
		},
		{
			"headers configuration is missing",
			nil,
			errMissingHeadersConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{HeadersConfig: tt.header}
			require.ErrorIs(t, cfg.Validate(), tt.expectedErr)
		})
	}
}
