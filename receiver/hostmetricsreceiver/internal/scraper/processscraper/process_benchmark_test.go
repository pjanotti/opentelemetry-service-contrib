// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package processscraper

import (
    "context"
    "testing"
)

func BenchmarkGetProcessHandlesInternal(b *testing.B) {
    ctx := context.Background()

    for i := 0; i < b.N; i++ {
        _, err := getProcessHandlesInternal(ctx)
        if err != nil {
            b.Fatalf("Failed to get process handles: %v", err)
        }
    }
}
