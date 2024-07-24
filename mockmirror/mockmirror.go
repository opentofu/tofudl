// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package mockmirror

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/opentofu/tofudl"
	"github.com/opentofu/tofudl/branding"
)

// New returns a mirror serving a fake archive signed with a GPG key for testing purposes.
func New(
	t *testing.T,
) Mirror {
	return NewFromBinary(t, buildFake(t))
}

// NewFromBinary returns a mirror serving a binary passed and signed with a GPG key for testing purposes.
func NewFromBinary(
	t *testing.T,
	binary []byte,
) Mirror {
	key, err := crypto.GenerateKey(branding.ProductName+" Test", "noreply@example.org", "rsa", 2048)
	if err != nil {
		panic(err)
	}
	pubKey, err := key.GetArmoredPublicKey()
	if err != nil {
		t.Fatalf("Failed to get public key (%v)", err)
	}

	builder, err := tofudl.NewReleaseBuilder(key)
	if err != nil {
		t.Fatalf("Failed to create release builder (%v)", err)
	}
	if err := builder.PackageBinary(tofudl.PlatformAuto, tofudl.ArchitectureAuto, binary, map[string][]byte{}); err != nil {
		t.Fatalf("Failed to package binary (%v)", err)
	}

	storage, err := tofudl.NewFilesystemStorage(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create storage (%v)", err)
	}

	tofudlMirror, err := tofudl.NewMirror(tofudl.MirrorConfig{}, storage, nil)
	if err != nil {
		t.Fatalf("Failed to create mirror (%v)", err)
	}

	if err := builder.Build(context.Background(), "1.0.0", tofudlMirror); err != nil {
		t.Fatalf("Failed to build release (%v)", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to open listen socket for mock mirror (%v)", err)
	}
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      tofudlMirror,
	}
	go func() {
		_ = srv.Serve(ln)
	}()
	t.Cleanup(func() {
		_ = srv.Shutdown(context.Background())
		_ = ln.Close()
	})
	return &mirror{
		addr:   ln.Addr().(*net.TCPAddr),
		pubKey: pubKey,
	}
}

// Mirror is a mock mirror for testing purposes holding a single version with a single binary for the current platform
// and architecture.
type Mirror interface {
	GPGKey() string
	APIURL() string
	DownloadMirrorURLTemplate() string
}

type mirror struct {
	addr   *net.TCPAddr
	pubKey string
}

func (m mirror) GPGKey() string {
	return m.pubKey
}

func (m mirror) APIURL() string {
	return "http://127.0.0.1:" + strconv.Itoa(m.addr.Port) + "/api.json"
}

func (m mirror) DownloadMirrorURLTemplate() string {
	return "http://127.0.0.1:" + strconv.Itoa(m.addr.Port) + "/v{{ .Version }}/{{ .Artifact }}"
}
