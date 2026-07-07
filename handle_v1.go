package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v6"
)

func (a *App) handleV1ImagesGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV1Error(w, GenerateRequestID(), http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
		return
	}
	requestID := GenerateRequestID()
	var request dalle.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeV1Error(w, requestID, http.StatusBadRequest, dalle.ErrInvalidInput, "invalid JSON request")
		return
	}
	result, err := a.Engine.Generate(request)
	if err != nil {
		writeV1EngineError(w, requestID, err)
		return
	}
	WriteSuccessResponse(w, result, requestID)
}

func (a *App) handleV1ImagesPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV1Error(w, GenerateRequestID(), http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
		return
	}
	requestID := GenerateRequestID()
	var request dalle.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeV1Error(w, requestID, http.StatusBadRequest, dalle.ErrInvalidInput, "invalid JSON request")
		return
	}
	result, err := a.Engine.Preview(request)
	if err != nil {
		writeV1EngineError(w, requestID, err)
		return
	}
	WriteSuccessResponse(w, result, requestID)
}

func (a *App) handleV1Images(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, GenerateRequestID(), http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
		return
	}
	requestID := GenerateRequestID()
	records, err := a.Engine.ListImages(dalle.ImageFilter{Series: r.URL.Query().Get("series")})
	if err != nil {
		writeV1EngineError(w, requestID, err)
		return
	}
	WriteSuccessResponse(w, records, requestID)
}

func (a *App) handleV1Image(w http.ResponseWriter, r *http.Request) {
	requestID := GenerateRequestID()
	id := strings.TrimPrefix(r.URL.Path, "/v1/images/")
	if strings.HasSuffix(id, "/regenerate") {
		if r.Method != http.MethodPost {
			writeV1Error(w, requestID, http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
			return
		}
		id = strings.TrimSuffix(id, "/regenerate")
		result, err := a.Engine.RegenerateImage(id)
		if err != nil {
			writeV1EngineError(w, requestID, err)
			return
		}
		WriteSuccessResponse(w, result, requestID)
		return
	}
	if strings.HasSuffix(id, "/export") {
		if r.Method != http.MethodPost {
			writeV1Error(w, requestID, http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
			return
		}
		id = strings.TrimSuffix(id, "/export")
		var options dalle.ExportImageOptions
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&options); err != nil {
				writeV1Error(w, requestID, http.StatusBadRequest, dalle.ErrInvalidInput, "invalid JSON request")
				return
			}
		}
		result, err := a.Engine.ExportImage(id, options)
		if err != nil {
			writeV1EngineError(w, requestID, err)
			return
		}
		WriteSuccessResponse(w, result, requestID)
		return
	}
	if r.Method == http.MethodDelete {
		if err := a.Engine.DeleteImage(id); err != nil {
			writeV1EngineError(w, requestID, err)
			return
		}
		WriteSuccessResponse(w, map[string]bool{"deleted": true}, requestID)
		return
	}
	if r.Method != http.MethodGet {
		writeV1Error(w, requestID, http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
		return
	}
	record, err := a.Engine.GetImage(id)
	if err != nil {
		writeV1EngineError(w, requestID, err)
		return
	}
	WriteSuccessResponse(w, record, requestID)
}

func (a *App) handleV1Series(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, GenerateRequestID(), http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
		return
	}
	requestID := GenerateRequestID()
	filter := dalle.SeriesFilter{
		IncludeHidden: r.URL.Query().Get("includeHidden") == "1",
		OnlyHidden:    r.URL.Query().Get("onlyHidden") == "1",
	}
	series, err := a.Engine.ListSeries(filter)
	if err != nil {
		writeV1EngineError(w, requestID, err)
		return
	}
	WriteSuccessResponse(w, series, requestID)
}

func (a *App) handleV1SeriesItem(w http.ResponseWriter, r *http.Request) {
	requestID := GenerateRequestID()
	path := strings.TrimPrefix(r.URL.Path, "/v1/series/")
	if strings.HasSuffix(path, "/hidden") {
		if r.Method != http.MethodPost {
			writeV1Error(w, requestID, http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
			return
		}
		name := strings.TrimSuffix(path, "/hidden")
		var request struct {
			Hidden bool `json:"hidden"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeV1Error(w, requestID, http.StatusBadRequest, dalle.ErrInvalidInput, "invalid JSON request")
			return
		}
		series, err := a.Engine.SetSeriesHidden(name, request.Hidden)
		if err != nil {
			writeV1EngineError(w, requestID, err)
			return
		}
		WriteSuccessResponse(w, series, requestID)
		return
	}
	switch r.Method {
	case http.MethodGet:
		series, err := a.Engine.GetSeries(path)
		if err != nil {
			writeV1EngineError(w, requestID, err)
			return
		}
		WriteSuccessResponse(w, series, requestID)
	case http.MethodPut:
		var series dalle.Series
		if err := json.NewDecoder(r.Body).Decode(&series); err != nil {
			writeV1Error(w, requestID, http.StatusBadRequest, dalle.ErrInvalidInput, "invalid JSON request")
			return
		}
		if strings.TrimSpace(series.Suffix) == "" {
			series.Suffix = path
		}
		saved, err := a.Engine.SaveSeries(series)
		if err != nil {
			writeV1EngineError(w, requestID, err)
			return
		}
		WriteSuccessResponse(w, saved, requestID)
	default:
		writeV1Error(w, requestID, http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
	}
}

func (a *App) handleV1Databases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, GenerateRequestID(), http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
		return
	}
	requestID := GenerateRequestID()
	archives, err := a.Engine.ListDatabaseArchives()
	if err != nil {
		writeV1EngineError(w, requestID, err)
		return
	}
	WriteSuccessResponse(w, archives, requestID)
}

func (a *App) handleV1Database(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, GenerateRequestID(), http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
		return
	}
	requestID := GenerateRequestID()
	path := strings.TrimPrefix(r.URL.Path, "/v1/databases/")
	if strings.Contains(path, "/records/") {
		parts := strings.SplitN(path, "/records/", 2)
		name := ""
		if len(parts) == 2 {
			name = parts[1]
		}
		limit := 200
		if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
			parsed, err := strconv.Atoi(rawLimit)
			if err != nil {
				writeV1Error(w, requestID, http.StatusBadRequest, dalle.ErrInvalidInput, "invalid limit")
				return
			}
			limit = parsed
		}
		result, err := a.Engine.ListDatabaseRecords(name, limit)
		if err != nil {
			writeV1EngineError(w, requestID, err)
			return
		}
		WriteSuccessResponse(w, result, requestID)
		return
	}
	version := path
	archive, err := a.Engine.GetDatabaseArchive(version)
	if err != nil {
		writeV1EngineError(w, requestID, err)
		return
	}
	WriteSuccessResponse(w, archive, requestID)
}

func (a *App) handleV1Validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV1Error(w, GenerateRequestID(), http.StatusMethodNotAllowed, dalle.ErrInvalidInput, "method not allowed")
		return
	}
	requestID := GenerateRequestID()
	if err := a.Engine.Validate(); err != nil {
		writeV1EngineError(w, requestID, err)
		return
	}
	WriteSuccessResponse(w, map[string]bool{"valid": true}, requestID)
}

func writeV1EngineError(w http.ResponseWriter, requestID string, err error) {
	code := dalle.ErrorCodeOf(err)
	status := http.StatusInternalServerError
	switch code {
	case dalle.ErrInvalidInput, dalle.ErrSeriesInvalid:
		status = http.StatusBadRequest
	case dalle.ErrSeriesNotFound, dalle.ErrArtifactMissing, dalle.ErrDatabaseVersionUnavailable:
		status = http.StatusNotFound
	case dalle.ErrRegenerationRefused, dalle.ErrDatabaseHashMismatch:
		status = http.StatusConflict
	case dalle.ErrProviderUnavailable:
		status = http.StatusServiceUnavailable
	case dalle.ErrProviderFailed:
		status = http.StatusBadGateway
	}
	writeV1Error(w, requestID, status, code, err.Error())
}

func writeV1Error(w http.ResponseWriter, requestID string, status int, code dalle.ErrorCode, message string) {
	if code == "" {
		code = dalle.ErrInvalidInput
	}
	WriteErrorResponse(w, NewAPIError(string(code), message, "").WithRequestID(requestID), status)
}
