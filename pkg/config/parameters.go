// Copyright Project Contour Authors
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

package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// ServerType is the name of a xDS server implementation.
type ServerType string

const ContourServerType ServerType = "contour"
const EnvoyServerType ServerType = "envoy"

// Validate the xDS server type.
func (s ServerType) Validate() error {
	switch s {
	case ContourServerType, EnvoyServerType:
		return nil
	default:
		return fmt.Errorf("invalid xDS server type %q", s)
	}
}

// ServerVersion is the version of a xDS server.
type ServerVersion string

const XDSv2 ServerVersion = "v2"
const XDSv3 ServerVersion = "v3"

// Validate the xDS server version.
func (s ServerVersion) Validate() error {
	switch s {
	case XDSv2, XDSv3:
		return nil
	default:
		return fmt.Errorf("invalid xDS version %q", s)
	}
}

// ClusterDNSFamilyType is the Ip family to use for resolving DNS
// names in an Envoy cluster configuration.
type ClusterDNSFamilyType string

func (c ClusterDNSFamilyType) Validate() error {
	switch c {
	case AutoClusterDNSFamily, IPv4ClusterDNSFamily, IPv6ClusterDNSFamily:
		return nil
	default:
		return fmt.Errorf("invalid cluster DNS lookup family %q", c)
	}
}

const AutoClusterDNSFamily ClusterDNSFamilyType = "auto"
const IPv4ClusterDNSFamily ClusterDNSFamilyType = "v4"
const IPv6ClusterDNSFamily ClusterDNSFamilyType = "v6"

// AccessLogType is the name of a supported access logging mechanism.
type AccessLogType string

func (a AccessLogType) Validate() error {
	switch a {
	case EnvoyAccessLog, JSONAccessLog:
		return nil
	default:
		return fmt.Errorf("invalid access log format %q", a)
	}
}

const EnvoyAccessLog AccessLogType = "envoy"
const JSONAccessLog AccessLogType = "json"

// HTTPVersionType is the name of a supported HTTP version.
type HTTPVersionType string

func (h HTTPVersionType) Validate() error {
	switch h {
	case HTTPVersion1, HTTPVersion2:
		return nil
	default:
		return fmt.Errorf("invalid HTTP version %q", h)
	}
}

const HTTPVersion1 HTTPVersionType = "http/1.1"
const HTTPVersion2 HTTPVersionType = "http/2"

// NamespacedName defines the namespace/name of the Kubernetes resource referred from the configuration file.
// Used for Contour configuration YAML file parsing, otherwise we could use K8s types.NamespacedName.
type NamespacedName struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// Validate that both name fields are present, or neither are.
func (n NamespacedName) Validate() error {
	if len(strings.TrimSpace(n.Name)) == 0 && len(strings.TrimSpace(n.Namespace)) == 0 {
		return nil
	}

	if len(strings.TrimSpace(n.Namespace)) == 0 {
		return errors.New("namespace must be defined")
	}

	if len(strings.TrimSpace(n.Name)) == 0 {
		return errors.New("name must be defined")
	}

	return nil
}

// TLSParameters holds configuration file TLS configuration details.
type TLSParameters struct {
	MinimumProtocolVersion string `yaml:"minimum-protocol-version"`

	// FallbackCertificate defines the namespace/name of the Kubernetes secret to
	// use as fallback when a non-SNI request is received.
	FallbackCertificate NamespacedName `yaml:"fallback-certificate,omitempty"`

	// ClientCertificate defines the namespace/name of the Kubernetes
	// secret containing the client certificate and private key
	// to be used when establishing TLS connection to upstream
	// cluster.
	ClientCertificate NamespacedName `yaml:"envoy-client-certificate,omitempty"`
}

// ServerParameters holds the configuration for the Contour xDS server.
type ServerParameters struct {
	// Defines the XDSServer to use for `contour serve`.
	// Defaults to "contour"
	XDSServerType ServerType `yaml:"xds-server-type,omitempty"`

	// Defines the XDS Server Version to use for `contour serve`
	// Defaults to "v2"
	XDSServerVersion ServerVersion `yaml:"xds-server-version"`
}

// LeaderElectionParameters holds the config bits for leader election
// inside the  configuration file.
type LeaderElectionParameters struct {
	LeaseDuration time.Duration `yaml:"lease-duration,omitempty"`
	RenewDeadline time.Duration `yaml:"renew-deadline,omitempty"`
	RetryPeriod   time.Duration `yaml:"retry-period,omitempty"`
	Namespace     string        `yaml:"configmap-namespace,omitempty"`
	Name          string        `yaml:"configmap-name,omitempty"`
}

// TimeoutParameters holds various configurable proxy timeout values.
type TimeoutParameters struct {
	// RequestTimeout sets the client request timeout globally for Contour. Note that
	// this is a timeout for the entire request, not an idle timeout. Omit or set to
	// "infinity" to disable the timeout entirely.
	//
	// See https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/filter/network/http_connection_manager/v2/http_connection_manager.proto#envoy-api-field-config-filter-network-http-connection-manager-v2-httpconnectionmanager-request-timeout
	// for more information.
	RequestTimeout string `yaml:"request-timeout,omitempty"`

	// ConnectionIdleTimeout defines how long the proxy should wait while there are
	// no active requests (for HTTP/1.1) or streams (for HTTP/2) before terminating
	// an HTTP connection. Set to "infinity" to disable the timeout entirely.
	//
	// See https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/core/protocol.proto#envoy-api-field-core-httpprotocoloptions-idle-timeout
	// for more information.
	ConnectionIdleTimeout string `yaml:"connection-idle-timeout,omitempty"`

	// StreamIdleTimeout defines how long the proxy should wait while there is no
	// request activity (for HTTP/1.1) or stream activity (for HTTP/2) before
	// terminating the HTTP request or stream. Set to "infinity" to disable the
	// timeout entirely.
	//
	// See https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/filter/network/http_connection_manager/v2/http_connection_manager.proto#envoy-api-field-config-filter-network-http-connection-manager-v2-httpconnectionmanager-stream-idle-timeout
	// for more information.
	StreamIdleTimeout string `yaml:"stream-idle-timeout,omitempty"`

	// MaxConnectionDuration defines the maximum period of time after an HTTP connection
	// has been established from the client to the proxy before it is closed by the proxy,
	// regardless of whether there has been activity or not. Omit or set to "infinity" for
	// no max duration.
	//
	// See https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/core/protocol.proto#envoy-api-field-core-httpprotocoloptions-max-connection-duration
	// for more information.
	MaxConnectionDuration string `yaml:"max-connection-duration,omitempty"`

	// ConnectionShutdownGracePeriod defines how long the proxy will wait between sending an
	// initial GOAWAY frame and a second, final GOAWAY frame when terminating an HTTP/2 connection.
	// During this grace period, the proxy will continue to respond to new streams. After the final
	// GOAWAY frame has been sent, the proxy will refuse new streams.
	//
	// See https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/filter/network/http_connection_manager/v2/http_connection_manager.proto#envoy-api-field-config-filter-network-http-connection-manager-v2-httpconnectionmanager-drain-timeout
	// for more information.
	ConnectionShutdownGracePeriod string `yaml:"connection-shutdown-grace-period,omitempty"`
}

// Validate the timeout parameters.
func (t TimeoutParameters) Validate() error {
	// We can't use `timeout.Parse` for validation here because
	// that would make an exported package depend on an internal
	// package.
	v := func(str string) error {
		switch str {
		case "", "infinity", "infinite":
			return nil
		default:
			_, err := time.ParseDuration(str)
			return err
		}
	}

	if err := v(t.RequestTimeout); err != nil {
		return fmt.Errorf("invalid request timeout %q: %w", t.RequestTimeout, err)
	}

	if err := v(t.ConnectionIdleTimeout); err != nil {
		return fmt.Errorf("connection idle timeout %q: %w", t.RequestTimeout, err)
	}

	if err := v(t.StreamIdleTimeout); err != nil {
		return fmt.Errorf("stream idle timeout %q: %w", t.RequestTimeout, err)
	}

	if err := v(t.MaxConnectionDuration); err != nil {
		return fmt.Errorf("max connection duration %q: %w", t.RequestTimeout, err)
	}

	if err := v(t.ConnectionShutdownGracePeriod); err != nil {
		return fmt.Errorf("connection shutdown grace period %q: %w", t.RequestTimeout, err)
	}

	return nil
}

// ClusterParameters holds various configurable cluster values.
type ClusterParameters struct {
	// DNSLookupFamily defines how external names are looked up
	// When configured as V4, the DNS resolver will only perform a lookup
	// for addresses in the IPv4 family. If V6 is configured, the DNS resolver
	// will only perform a lookup for addresses in the IPv6 family.
	// If AUTO is configured, the DNS resolver will first perform a lookup
	// for addresses in the IPv6 family and fallback to a lookup for addresses
	// in the IPv4 family.
	// Note: This only applies to externalName clusters.
	//
	// See https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/cluster.proto#enum-cluster-dnslookupfamily
	// for more information.
	DNSLookupFamily ClusterDNSFamilyType `yaml:"dns-lookup-family"`
}

// Parameters contains the configuration file parameters for the
// Contour ingress controller.
type Parameters struct {
	// Enable debug logging
	Debug bool

	// Kubernetes client parameters.
	InCluster  bool   `yaml:"incluster,omitempty"`
	Kubeconfig string `yaml:"kubeconfig,omitempty"`

	// Server contains parameters for the xDS server.
	Server ServerParameters `yaml:"server,omitempty"`

	// Address to be placed in status.loadbalancer field of Ingress objects.
	// May be either a literal IP address or a host name.
	// The value will be placed directly into the relevant field inside the status.loadBalancer struct.
	IngressStatusAddress string `yaml:"ingress-status-address,omitempty"`

	// AccessLogFormat sets the global access log format.
	// Valid options are 'envoy' or 'json'
	AccessLogFormat AccessLogType `yaml:"accesslog-format,omitempty"`

	// AccessLogFields sets the fields that JSON logging will
	// output when AccessLogFormat is json.
	AccessLogFields []string `yaml:"json-fields,omitempty"`

	// TLS contains TLS policy parameters.
	TLS TLSParameters `yaml:"tls,omitempty"`

	// DisablePermitInsecure disables the use of the
	// permitInsecure field in HTTPProxy.
	DisablePermitInsecure bool `yaml:"disablePermitInsecure,omitempty"`

	// LeaderElection contains leader election parameters.
	LeaderElection LeaderElectionParameters `yaml:"leaderelection,omitempty"`

	// Timeouts holds various configurable timeouts that can
	// be set in the config file.
	Timeouts TimeoutParameters `yaml:"timeouts,omitempty"`

	// Namespace of the envoy service to inspect for Ingress status details.
	EnvoyServiceNamespace string `yaml:"envoy-service-namespace,omitempty"`

	// Name of the envoy service to inspect for Ingress status details.
	EnvoyServiceName string `yaml:"envoy-service-name,omitempty"`

	// DefaultHTTPVersions defines the default set of HTTPS
	// versions the proxy should accept. HTTP versions are
	// strings of the form "HTTP/xx". Supported versions are
	// "HTTP/1.1" and "HTTP/2".
	//
	// If this field not specified, all supported versions are accepted.
	DefaultHTTPVersions []HTTPVersionType `yaml:"default-http-versions"`

	// Cluster holds various configurable Envoy cluster values that can
	// be set in the config file.
	Cluster ClusterParameters `yaml:"cluster,omitempty"`
}

// Validate verifies that the parameter values do not have any syntax errors.
func (p *Parameters) Validate() error {
	if err := p.Cluster.DNSLookupFamily.Validate(); err != nil {
		return err
	}

	if err := p.Server.XDSServerType.Validate(); err != nil {
		return err
	}

	if err := p.Server.XDSServerVersion.Validate(); err != nil {
		return err
	}

	if err := p.AccessLogFormat.Validate(); err != nil {
		return err
	}

	for _, f := range p.AccessLogFields {
		if _, ok := JSONFields[f]; !ok {
			return fmt.Errorf("invalid JSON log field name %s", f)
		}
	}

	// Check TLS secret names.
	if err := p.TLS.FallbackCertificate.Validate(); err != nil {
		return fmt.Errorf("invalid TLS fallback certificate: %w", err)
	}

	if err := p.TLS.ClientCertificate.Validate(); err != nil {
		return fmt.Errorf("invalid TLS client certificate: %w", err)
	}

	if err := p.Timeouts.Validate(); err != nil {
		return err
	}

	for _, v := range p.DefaultHTTPVersions {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Defaults returns the default set of parameters.
func Defaults() Parameters {
	contourNamespace := GetenvOr("CONTOUR_NAMESPACE", "projectcontour")

	return Parameters{
		Debug:      false,
		InCluster:  false,
		Kubeconfig: filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		Server: ServerParameters{
			XDSServerType:    ContourServerType,
			XDSServerVersion: XDSv2,
		},
		IngressStatusAddress:  "",
		AccessLogFormat:       DEFAULT_ACCESS_LOG_TYPE,
		AccessLogFields:       DefaultFields,
		TLS:                   TLSParameters{},
		DisablePermitInsecure: false,
		LeaderElection: LeaderElectionParameters{
			LeaseDuration: time.Second * 15,
			RenewDeadline: time.Second * 10,
			RetryPeriod:   time.Second * 2,
			Name:          "leader-elect",
			Namespace:     contourNamespace,
		},
		Timeouts: TimeoutParameters{
			// This is chosen as a rough default to stop idle connections wasting resources,
			// without stopping slow connections from being terminated too quickly.
			ConnectionIdleTimeout: "60s",
		},
		EnvoyServiceName:      "envoy",
		EnvoyServiceNamespace: contourNamespace,
		DefaultHTTPVersions:   []HTTPVersionType{},
		Cluster: ClusterParameters{
			DNSLookupFamily: AutoClusterDNSFamily,
		},
	}
}

// Parse reads parameters from a YAML input stream. Any parameters
// not specified by the input are according to Defaults().
func Parse(in io.Reader) (*Parameters, error) {
	conf := Defaults()
	decoder := yaml.NewDecoder(in)

	decoder.SetStrict(true)

	if err := decoder.Decode(&conf); err != nil {
		// The YAML decoder will return EOF if there are
		// no YAML nodes in the results. In this case, we just
		// want to succeed and return the defaults.
		if err != io.EOF {
			return nil, fmt.Errorf("failed to parse configuration: %w", err)
		}
	}

	// Force the version string to match the lowercase version
	// constants (assuming that it will match).
	for i, v := range conf.DefaultHTTPVersions {
		conf.DefaultHTTPVersions[i] = HTTPVersionType(strings.ToLower(string(v)))
	}

	return &conf, nil
}

// GetenvOr reads an environment or return a default value
func GetenvOr(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
