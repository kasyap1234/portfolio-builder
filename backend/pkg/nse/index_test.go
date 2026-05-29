package nse

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveIndexName(t *testing.T) {
	name, ok := ResolveIndexName("^NSEI")
	assert.True(t, ok)
	assert.Equal(t, nifty50Index, name)

	_, ok = ResolveIndexName("RELIANCE.NS")
	assert.False(t, ok)
}

func TestFetchIndexPE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/allIndices", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": [
				{"index": "NIFTY 50", "pe": "20.58", "pb": "3.26"},
				{"index": "NIFTY BANK", "pe": "13.85", "pb": "1.87"}
			]
		}`))
	}))
	defer server.Close()

	orig := allIndicesURL
	allIndicesURL = server.URL + "/api/allIndices"
	t.Cleanup(func() { allIndicesURL = orig })

	pe, err := FetchIndexPE("NIFTY 50")
	require.NoError(t, err)
	assert.InDelta(t, 20.58, pe, 0.01)

	pe, err = FetchPEForSymbol("^NSEI")
	require.NoError(t, err)
	assert.InDelta(t, 20.58, pe, 0.01)

	_, err = FetchPEForSymbol("RELIANCE.NS")
	assert.ErrorIs(t, err, ErrNotIndexSymbol)
}

func TestFetchIndexPE_missingIndex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"index":"NIFTY BANK","pe":"13.85"}]}`))
	}))
	defer server.Close()

	orig := allIndicesURL
	allIndicesURL = server.URL + "/api/allIndices"
	t.Cleanup(func() { allIndicesURL = orig })

	_, err := FetchIndexPE("NIFTY 50")
	assert.Error(t, err)
}
