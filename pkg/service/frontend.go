// Copyright 2026 Samvaad Project
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

package service

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed frontend_out
var frontendFS embed.FS

func (s *SamvaadServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	// Support both /healthz and plain text requests to / for backward compatibility with orchestrators
	if r.URL.Path == "/healthz" || (r.URL.Path == "/" && r.Header.Get("Accept") == "text/plain") {
		s.healthCheck(w, r)
		return
	}

	subFS, err := fs.Sub(frontendFS, "frontend_out")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// For clean Next.js export routing:
	// If the path doesn't have an extension and is not "/",
	// check if a corresponding .html file exists, e.g. "/room" -> "/room.html"
	path := r.URL.Path
	if path != "/" && !strings.Contains(path, ".") {
		trimmedPath := strings.TrimPrefix(path, "/")
		if _, err := subFS.Open(trimmedPath + ".html"); err == nil {
			r.URL.Path = path + ".html"
		}
	}

	http.FileServer(http.FS(subFS)).ServeHTTP(w, r)
}
