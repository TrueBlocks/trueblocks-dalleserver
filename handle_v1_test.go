package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v6"
)

func newV1TestApp(t *testing.T) *App {
	t.Helper()
	engine, err := dalle.New(dalle.Config{DataDir: filepath.Join(t.TempDir(), "dalle-data")})
	if err != nil {
		t.Fatalf("New engine: %v", err)
	}
	return &App{Config: Config{}, Engine: engine}
}

func decodeAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder) APIResponse {
	t.Helper()
	var response APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode API response: %v\n%s", err, recorder.Body.String())
	}
	return response
}

func TestHandleV1ImagesPreview(t *testing.T) {
	app := newV1TestApp(t)
	body := bytes.NewBufferString(`{"input":"Person Tour Coordinates"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/images/preview", body)
	recorder := httptest.NewRecorder()

	app.handleV1ImagesPreview(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	response := decodeAPIResponse(t, recorder)
	if !response.Success {
		t.Fatalf("expected success response: %#v", response)
	}
	encoded, err := json.Marshal(response.Data)
	if err != nil {
		t.Fatalf("marshal response data: %v", err)
	}
	var result dalle.GenerateResult
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatalf("decode generate result: %v", err)
	}
	if result.MetadataPath == "" || result.Metadata.Prompts.Prompt == "" {
		t.Fatalf("expected preview metadata result: %#v", result)
	}
}

func TestHandleV1SeriesSaveAndShow(t *testing.T) {
	app := newV1TestApp(t)
	body := bytes.NewBufferString(`{"last":7,"purpose":"test"}`)
	putRequest := httptest.NewRequest(http.MethodPut, "/v1/series/Test%20Series", body)
	putRecorder := httptest.NewRecorder()

	app.handleV1SeriesItem(putRecorder, putRequest)

	if putRecorder.Code != http.StatusOK {
		t.Fatalf("expected save status 200, got %d: %s", putRecorder.Code, putRecorder.Body.String())
	}
	getRequest := httptest.NewRequest(http.MethodGet, "/v1/series/Test%20Series", nil)
	getRecorder := httptest.NewRecorder()
	app.handleV1SeriesItem(getRecorder, getRequest)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected show status 200, got %d: %s", getRecorder.Code, getRecorder.Body.String())
	}
	response := decodeAPIResponse(t, getRecorder)
	encoded, err := json.Marshal(response.Data)
	if err != nil {
		t.Fatalf("marshal response data: %v", err)
	}
	var series dalle.Series
	if err := json.Unmarshal(encoded, &series); err != nil {
		t.Fatalf("decode series: %v", err)
	}
	if series.Suffix != "test-series" || series.Last != 7 {
		t.Fatalf("unexpected series: %#v", series)
	}
}

func TestHandleV1ImageMissingMapsToNotFound(t *testing.T) {
	app := newV1TestApp(t)
	request := httptest.NewRequest(http.MethodGet, "/v1/images/missing", nil)
	recorder := httptest.NewRecorder()

	app.handleV1Image(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", recorder.Code, recorder.Body.String())
	}
	response := decodeAPIResponse(t, recorder)
	if response.Success || response.Error == nil || !strings.Contains(response.Error.Code, string(dalle.ErrArtifactMissing)) {
		t.Fatalf("expected artifact missing error: %#v", response)
	}
}

func TestHandleV1ImageDelete(t *testing.T) {
	app := newV1TestApp(t)
	result, err := app.Engine.Generate(dalle.GenerateRequest{Input: "Person Tour Coordinates"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	request := httptest.NewRequest(http.MethodDelete, "/v1/images/"+result.Metadata.ImageID, nil)
	recorder := httptest.NewRecorder()

	app.handleV1Image(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if _, err := app.Engine.GetImage(result.Metadata.ImageID); dalle.ErrorCodeOf(err) != dalle.ErrArtifactMissing {
		t.Fatalf("expected deleted image lookup to fail, got %v", err)
	}
}

func TestHandleV1DatabaseRecords(t *testing.T) {
	app := newV1TestApp(t)
	request := httptest.NewRequest(http.MethodGet, "/v1/databases/1.0.0/records/nouns?limit=2", nil)
	recorder := httptest.NewRecorder()

	app.handleV1Database(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	response := decodeAPIResponse(t, recorder)
	encoded, err := json.Marshal(response.Data)
	if err != nil {
		t.Fatalf("marshal response data: %v", err)
	}
	var result dalle.DatabaseRecordsResult
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatalf("decode records result: %v", err)
	}
	if result.Name != "nouns" || len(result.Records) != 2 {
		t.Fatalf("unexpected records result: %#v", result)
	}
}
