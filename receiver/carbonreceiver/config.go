// Copyright 2019, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package carbonreceiver

import (
	"time"

	"github.com/open-telemetry/opentelemetry-collector/config/configmodels"
)

const (
	// TCPIdleTimeoutDefault is the default timeout for idle TCP connections.
	TCPIdleTimeoutDefault = 30 * time.Second
)

// Config defines configuration for the Carbon receiver.
type Config struct {
	configmodels.ReceiverSettings `mapstructure:",squash"`

	// TODO: Remove comments below. They are temporary while developing to
	// record some investigations.

	// TODO: Remove this comment?
	// ServerAcceptDeadline puts a deadline on Accept for TCP connections. In
	// principle the receiver doesn't need that: it should be accepting connections
	// and letting the Close get rid of any blocked/waiting Accept.
	// See https://github.com/signalfx/gateway/blob/master/protocol/carbon/carbonlistener.go#L206
	// In principle this does not seem necessary.
	// ServerAcceptDeadline time.Duration

	// TODO: Remove this comment?
	// ConnectionTimeout is used for two distinct purposes. The first one is
	// identical to the purpose of ServerAcceptDeadline above but for UDP connections.
	// The second one is to use it as an idle timeout for the TCP connections, that
	// will be used in the call the waits for actual data from the client.
	// ConnectionTimeout    time.Duration

	// TODO: Understand what types of descontructors are really needed.
	// Will choose an arbitrary one to get started.
	// MetricDeconstructor  metricdeconstructor.MetricDeconstructor

	// TODO: This is actually Carbon-Line receiver since Gateway doesn't support pickle format.
	//		consider renaming? Or future option for it?

	// Protocol is actually "transport"" ie.: UDP or TCP.
	// 	Protocol             string
	// The file below has the default carbon settings
	// https://github.com/graphite-project/carbon/blob/master/conf/carbon.conf.example

	// Transport is either "tcp" or "udp".
	Transport string `mapstructure:"transport"`

	// TCPIdleTimeout is the timout for idle TCP connections, it is ignored
	// if transport being used is UDP.
	TCPIdleTimeout time.Duration `mapstructure:"tcp_idle_timeout"`
}
