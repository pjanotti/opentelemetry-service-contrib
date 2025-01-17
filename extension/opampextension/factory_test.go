// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.1

package opampextension

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensiontest"
)

func TestFactory_CreateDefaultConfig(t *testing.T) {
	cfg := NewFactory().CreateDefaultConfig()
	assert.Equal(t, createDefaultConfig().(*Config), cfg)

	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
	ext, err := createExtension(context.Background(), extensiontest.NewNopSettings(), cfg)
	require.NoError(t, err)
	require.NotNil(t, ext)
	require.NoError(t, ext.Shutdown(context.Background()))
}

func TestFactory_Create(t *testing.T) {
	cfg := NewFactory().CreateDefaultConfig()
	ext, err := createExtension(context.Background(), extensiontest.NewNopSettings(), cfg)
	require.NoError(t, err)
	require.NotNil(t, ext)
	require.NoError(t, ext.Shutdown(context.Background()))
}
