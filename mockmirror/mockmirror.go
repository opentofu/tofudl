// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package mockmirror

import (
	"net"
	"net/http"
	"strconv"
	"testing"

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

	platform, err := tofudl.PlatformAuto.ResolveAuto()
	if err != nil {
		t.Fatalf("Failed to resolve platform (%v)", err)
	}
	arch, err := tofudl.ArchitectureAuto.ResolveAuto()
	if err != nil {
		t.Fatalf("Failed to resolve architecture (%v)", err)
	}
	version := "1.0.0"

	archiveName := branding.ArtifactPrefix + version + "_" + string(platform) + "_" + string(arch) + ".tar.gz"
	sumsName := branding.ArtifactPrefix + version + "_SHA256SUMS"
	sigName := branding.ArtifactPrefix + version + "_SHA256SUMS.gpgsig"

	apiResponse := buildAPI(t, version, []string{
		archiveName, sumsName, sigName,
	})
	archive := buildTarFile(t, binary)
	sums := buildSumsFile(map[string][]byte{
		archiveName: archive,
	})
	sig := signFile(t, sums, key)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to open listen socket for mock mirror (%v)", err)
	}
	addr := ln.Addr().(*net.TCPAddr)
	mirrorInstance := &mirror{
		addr:   addr,
		pubKey: pubKey,
		files: map[string][]byte{
			"/api.json":                        apiResponse,
			"/v" + version + "/" + archiveName: archive,
			"/v" + version + "/" + sumsName:    sums,
			"/v" + version + "/" + sigName:     sig,
		},
	}
	go func() {
		_ = http.Serve(ln, mirrorInstance)
	}()
	t.Cleanup(func() {
		_ = ln.Close()
	})
	return mirrorInstance
}

type Mirror interface {
	GPGKey() string
	APIURL() string
	DownloadMirrorURLTemplate() string
}

type mirror struct {
	addr   *net.TCPAddr
	pubKey string
	files  map[string][]byte
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

func (m mirror) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	contents, ok := m.files[request.RequestURI]
	if !ok {
		writer.WriteHeader(404)
		return
	}
	_, _ = writer.Write(contents)
}
