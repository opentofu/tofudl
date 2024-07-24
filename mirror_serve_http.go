package tofudl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func (m *mirror) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := context.Background()
	if request.RequestURI == "/api.json" {
		m.serveAPI(ctx, writer)
		return
	}
	m.serveAsset(ctx, writer, request)
}

func (m *mirror) serveAPI(ctx context.Context, writer http.ResponseWriter) {
	versionList, err := m.ListVersions(ctx)
	if err != nil {
		m.badGateway(writer)
		return
	}
	response := APIResponse{
		Versions: versionList,
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		m.badGateway(writer)
		return
	}
	writer.WriteHeader(200)
	writer.Header().Set("Content-Type", "application/json")
	_, _ = writer.Write(encoded)
}

func (m *mirror) serveAsset(ctx context.Context, writer http.ResponseWriter, request *http.Request) {
	if !strings.HasPrefix(request.RequestURI, "/") {
		m.badRequest(writer)
		return
	}
	parts := strings.Split(request.RequestURI, "/")
	if len(parts) != 3 {
		m.notFound(writer)
		return
	}
	version := Version(strings.TrimPrefix(parts[1], "v"))
	if err := version.Validate(); err != nil {
		m.notFound(writer)
		return
	}
	versions, err := m.ListVersions(ctx)
	if err != nil {
		m.badGateway(writer)
		return
	}
	var foundVersion *VersionWithArtifacts
	for _, ver := range versions {
		if ver.ID == version {
			foundVersion = &ver
			break
		}
	}
	if foundVersion == nil {
		m.notFound(writer)
		return
	}
	// TODO implement stream reading
	contents, err := m.DownloadArtifact(ctx, *foundVersion, parts[2])
	if err != nil {
		m.badGateway(writer)
		return
	}
	writer.WriteHeader(200)
	writer.Header().Set("Content-Type", "application/octet-stream")
	_, _ = writer.Write(contents)
}

func (m *mirror) badRequest(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusBadRequest)
	writer.Header().Set("Content-Type", "text/html")
	_, _ = writer.Write([]byte("<h1>Bad request</h1>"))
}

func (m *mirror) badGateway(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusBadGateway)
	writer.Header().Set("Content-Type", "text/html")
	_, _ = writer.Write([]byte("<h1>Bad gateway</h1>"))
}

func (m *mirror) notFound(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusNotFound)
	writer.Header().Set("Content-Type", "text/html")
	_, _ = writer.Write([]byte("<h1>Bad gateway</h1>"))
}
