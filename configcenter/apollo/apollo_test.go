package apollo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApolloConf_Validate(t *testing.T) {
	tests := []struct {
		name    string
		conf    ApolloConf
		wantErr bool
	}{
		{
			name: "valid config",
			conf: ApolloConf{
				AppID:    "test-app",
				MetaAddr: "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "empty meta addr",
			conf: ApolloConf{
				AppID: "test-app",
			},
			wantErr: true,
		},
		{
			name: "empty app id",
			conf: ApolloConf{
				MetaAddr: "http://localhost:8080",
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			conf:    ApolloConf{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conf.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildApolloConfig(t *testing.T) {
	conf := ApolloConf{
		AppID:          "test-app",
		Cluster:        "prod",
		NamespaceName:  "application.json",
		IP:             "127.0.0.1",
		MetaAddr:       "http://localhost:8080",
		Secret:         "test-secret",
		IsBackupConfig: true,
		BackupPath:     "/tmp/apollo",
		MustStart:      false,
	}

	apolloConf := buildApolloConfig(conf)

	assert.Equal(t, "test-app", apolloConf.AppID)
	assert.Equal(t, "prod", apolloConf.Cluster)
	assert.Equal(t, "application.json", apolloConf.NamespaceName)
	assert.Equal(t, "127.0.0.1", apolloConf.IP)
	assert.Equal(t, "test-secret", apolloConf.Secret)
	assert.Equal(t, true, apolloConf.IsBackupConfig)
	assert.Equal(t, "/tmp/apollo", apolloConf.BackupConfigPath)
	assert.Equal(t, false, apolloConf.MustStart)
}

func TestBuildApolloConfig_Defaults(t *testing.T) {
	conf := ApolloConf{
		AppID:    "test-app",
		MetaAddr: "http://localhost:8080",
	}

	apolloConf := buildApolloConfig(conf)

	assert.Equal(t, "test-app", apolloConf.AppID)
	assert.Equal(t, "", apolloConf.Cluster)       // Will use default in actual usage
	assert.Equal(t, "", apolloConf.NamespaceName) // Will use default in actual usage
	assert.Equal(t, "http://localhost:8080", apolloConf.IP)
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: "",
		},
		{
			name:     "string value",
			input:    "test",
			expected: "test",
		},
		{
			name:     "int value",
			input:    123,
			expected: "123",
		},
		{
			name:     "bool value",
			input:    true,
			expected: "true",
		},
		{
			name:     "map value",
			input:    map[string]string{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "empty string value",
			input:    "",
			expected: "",
		},
		{
			name:     "float value",
			input:    3.14,
			expected: "3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildApolloConfig_IPOverride(t *testing.T) {
	// Test that IP field overrides MetaAddr when both are provided
	conf := ApolloConf{
		AppID:    "test-app",
		MetaAddr: "http://localhost:8080",
		IP:       "192.168.1.100",
	}

	apolloConf := buildApolloConfig(conf)

	// IP should override MetaAddr
	assert.Equal(t, "192.168.1.100", apolloConf.IP)
}

func TestBuildApolloConfig_MetaAddrAsDefault(t *testing.T) {
	// Test that MetaAddr is used when IP is not provided
	conf := ApolloConf{
		AppID:    "test-app",
		MetaAddr: "http://localhost:8080",
	}

	apolloConf := buildApolloConfig(conf)

	// MetaAddr should be used as default for IP
	assert.Equal(t, "http://localhost:8080", apolloConf.IP)
}

func TestBuildApolloConfig_IPFieldAcceptsURL(t *testing.T) {
	// Test that IP field accepts full URL (not just bare IP)
	// This is supported by Apollo SDK as shown in official examples
	// Reference: https://github.com/apolloconfig/agollo README.md
	conf := ApolloConf{
		AppID:    "test-app",
		MetaAddr: "http://config.example.com:8080",
	}

	apolloConf := buildApolloConfig(conf)

	// IP field can hold full URL with scheme and port
	assert.Equal(t, "http://config.example.com:8080", apolloConf.IP)

	// Verify with HTTPS
	confHTTPS := ApolloConf{
		AppID:    "test-app",
		MetaAddr: "https://config.example.com",
	}

	apolloConfHTTPS := buildApolloConfig(confHTTPS)
	assert.Equal(t, "https://config.example.com", apolloConfHTTPS.IP)
}
