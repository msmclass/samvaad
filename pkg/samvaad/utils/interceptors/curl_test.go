package interceptors

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
)

func TestCurlPrinter(t *testing.T) {
	var buf bytes.Buffer
	err := printCurl(&buf,
		"http://localhost:8080",
		"example", "Service", "Do",
		http.Header{"X-Test": {"true"}},
		&samvaad.Room{Name: "test"},
	)
	require.NoError(t, err)
	require.Equal(t, `curl -X POST \
	-H 'X-Test: true' \
	-H 'Content-Type: application/json' \
	--data '{"name":"test"}' \
	http://localhost:8080/twirp/example.Service/Do
`, buf.String())
}
