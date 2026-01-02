package controller

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hodlbook/internal/models"
	"hodlbook/pkg/integrations/prices"

	"github.com/gin-gonic/gin"
)

type RowError struct {
	Row     int             `json:"row"`
	Data    json.RawMessage `json:"data"`
	Field   string          `json:"field,omitempty"`
	Message string          `json:"message"`
}

type ImportResponse struct {
	ID           int64      `json:"id"`
	Imported     int        `json:"imported"`
	Failed       int        `json:"failed"`
	Total        int        `json:"total"`
	Status       string     `json:"status"`
	Errors       []RowError `json:"errors,omitempty"`
}

// ExportAssets exports all assets as CSV or JSON
// @Summary Export assets
// @Description Export all assets as CSV or JSON file
// @Tags data
// @Produce octet-stream
// @Param format query string true "Export format (csv or json)"
// @Success 200 {file} file
// @Failure 400 {object} APIError
// @Router /api/assets/export [get]
func (c *Controller) ExportAssets(ctx *gin.Context) {
	format := strings.ToLower(ctx.DefaultQuery("format", "csv"))
	if format != "csv" && format != "json" {
		badRequest(ctx, "format must be csv or json")
		return
	}

	assets, err := c.repo.GetAllAssets()
	if err != nil {
		internalError(ctx, "failed to fetch assets")
		return
	}

	filename := fmt.Sprintf("assets_%s.%s", time.Now().Format("2006-01-02"), format)

	if format == "json" {
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		ctx.Header("Content-Type", "application/json")
		ctx.JSON(http.StatusOK, assets)
		return
	}

	data := assetsToCSV(assets)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	ctx.Header("Content-Type", "text/csv")
	ctx.Data(http.StatusOK, "text/csv", data)
}

// ExportExchanges exports all exchanges as CSV or JSON
// @Summary Export exchanges
// @Description Export all exchanges as CSV or JSON file
// @Tags data
// @Produce octet-stream
// @Param format query string true "Export format (csv or json)"
// @Success 200 {file} file
// @Failure 400 {object} APIError
// @Router /api/exchanges/export [get]
func (c *Controller) ExportExchanges(ctx *gin.Context) {
	format := strings.ToLower(ctx.DefaultQuery("format", "csv"))
	if format != "csv" && format != "json" {
		badRequest(ctx, "format must be csv or json")
		return
	}

	exchanges, err := c.repo.GetAllExchanges()
	if err != nil {
		internalError(ctx, "failed to fetch exchanges")
		return
	}

	filename := fmt.Sprintf("exchanges_%s.%s", time.Now().Format("2006-01-02"), format)

	if format == "json" {
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		ctx.Header("Content-Type", "application/json")
		ctx.JSON(http.StatusOK, exchanges)
		return
	}

	data := exchangesToCSV(exchanges)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	ctx.Header("Content-Type", "text/csv")
	ctx.Data(http.StatusOK, "text/csv", data)
}

// ImportAssets imports assets from uploaded CSV or JSON file
// @Summary Import assets
// @Description Import assets from CSV or JSON file. Valid rows are imported immediately.
// @Tags data
// @Accept multipart/form-data
// @Produce json
// @Param format query string true "Import format (csv or json)"
// @Param file formData file true "File to import"
// @Success 200 {object} ImportResponse
// @Failure 400 {object} APIError
// @Router /api/assets/import [post]
func (c *Controller) ImportAssets(ctx *gin.Context) {
	format := strings.ToLower(ctx.DefaultQuery("format", "csv"))
	if format != "csv" && format != "json" {
		badRequest(ctx, "format must be csv or json")
		return
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		badRequest(ctx, "file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		badRequest(ctx, "failed to read file")
		return
	}

	var validAssets []models.Asset
	var rowErrors []RowError

	if format == "json" {
		validAssets, rowErrors = parseAssetsFromJSON(data)
	} else {
		validAssets, rowErrors = parseAssetsFromCSV(data)
	}

	// Validate symbols against price providers
	supportedSymbols, currentPrices := getSupportedSymbolsWithPrices()
	var finalAssets []models.Asset
	for i, asset := range validAssets {
		if _, ok := supportedSymbols[strings.ToUpper(asset.Symbol)]; !ok {
			rowData, _ := json.Marshal(asset)
			rowErrors = append(rowErrors, RowError{
				Row:     i + 1,
				Data:    rowData,
				Field:   "symbol",
				Message: fmt.Sprintf("symbol %q is not supported by price providers", asset.Symbol),
			})
		} else {
			finalAssets = append(finalAssets, asset)
		}
	}

	// Import valid assets
	imported := 0
	importNote := fmt.Sprintf("%s imported", format)
	for _, asset := range finalAssets {
		asset.Symbol = strings.ToUpper(asset.Symbol)
		if asset.Timestamp.IsZero() {
			asset.Timestamp = time.Now()
		}
		if asset.Notes == "" {
			asset.Notes = importNote
		} else {
			asset.Notes = asset.Notes + " (" + importNote + ")"
		}
		if err := c.repo.CreateAsset(&asset); err == nil {
			imported++
			// Create price entry at asset timestamp for USD value display
			if price, ok := currentPrices[asset.Symbol]; ok {
				c.repo.CreatePrice(&models.Price{
					Symbol:    asset.Symbol,
					Currency:  "USD",
					Price:     price,
					Timestamp: asset.Timestamp,
				})
			}
			if c.assetCreatedPub != nil {
				assetJSON, _ := json.Marshal(asset)
				c.assetCreatedPub.Publish(assetJSON)
			}
		}
	}

	// Determine status
	status := "completed"
	if imported == 0 && len(rowErrors) > 0 {
		status = "failed"
	} else if len(rowErrors) > 0 {
		status = "partial"
	}

	// Save import log
	failedDataJSON, _ := json.Marshal(rowErrors)
	importLog := &models.ImportLog{
		Filename:     header.Filename,
		Format:       format,
		EntityType:   "asset",
		TotalRows:    len(validAssets) + len(rowErrors),
		ImportedRows: imported,
		FailedRows:   len(rowErrors),
		Status:       status,
		FailedData:   string(failedDataJSON),
	}
	c.repo.CreateImportLog(importLog)

	ctx.JSON(http.StatusOK, ImportResponse{
		ID:       importLog.ID,
		Imported: imported,
		Failed:   len(rowErrors),
		Total:    importLog.TotalRows,
		Status:   status,
		Errors:   rowErrors,
	})
}

// ListImportLogs returns all import logs
// @Summary List import logs
// @Description Get all import history
// @Tags data
// @Produce json
// @Success 200 {array} models.ImportLog
// @Router /api/imports [get]
func (c *Controller) ListImportLogs(ctx *gin.Context) {
	logs, err := c.repo.ListImportLogs()
	if err != nil {
		internalError(ctx, "failed to fetch import logs")
		return
	}
	ctx.JSON(http.StatusOK, logs)
}

// GetImportLog returns a specific import log
// @Summary Get import log
// @Description Get a specific import log with failed data
// @Tags data
// @Produce json
// @Param id path int true "Import log ID"
// @Success 200 {object} models.ImportLog
// @Failure 404 {object} APIError
// @Router /api/imports/{id} [get]
func (c *Controller) GetImportLog(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		badRequest(ctx, "invalid id")
		return
	}

	log, err := c.repo.GetImportLogByID(id)
	if err != nil {
		notFound(ctx, "import log not found")
		return
	}

	ctx.JSON(http.StatusOK, log)
}

// RetryImport retries importing failed rows with corrected data
// @Summary Retry import
// @Description Retry importing failed rows with corrected data
// @Tags data
// @Accept json
// @Produce json
// @Param id path int true "Import log ID"
// @Param assets body []models.Asset true "Corrected assets"
// @Success 200 {object} ImportResponse
// @Failure 400 {object} APIError
// @Failure 404 {object} APIError
// @Router /api/imports/{id}/retry [post]
func (c *Controller) RetryImport(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		badRequest(ctx, "invalid id")
		return
	}

	importLog, err := c.repo.GetImportLogByID(id)
	if err != nil {
		notFound(ctx, "import log not found")
		return
	}

	var assets []models.Asset
	if err := ctx.ShouldBindJSON(&assets); err != nil {
		badRequestWithDetails(ctx, "invalid input", err.Error())
		return
	}

	supportedSymbols, currentPrices := getSupportedSymbolsWithPrices()
	var validAssets []models.Asset
	var rowErrors []RowError

	for i, asset := range assets {
		if err := validateAssetFields(&asset); err != nil {
			rowData, _ := json.Marshal(asset)
			rowErrors = append(rowErrors, RowError{
				Row:     i + 1,
				Data:    rowData,
				Message: err.Error(),
			})
			continue
		}

		if _, ok := supportedSymbols[strings.ToUpper(asset.Symbol)]; !ok {
			rowData, _ := json.Marshal(asset)
			rowErrors = append(rowErrors, RowError{
				Row:     i + 1,
				Data:    rowData,
				Field:   "symbol",
				Message: fmt.Sprintf("symbol %q is not supported by price providers", asset.Symbol),
			})
		} else {
			validAssets = append(validAssets, asset)
		}
	}

	imported := 0
	importNote := fmt.Sprintf("%s imported", importLog.Format)
	for _, asset := range validAssets {
		asset.Symbol = strings.ToUpper(asset.Symbol)
		if asset.Timestamp.IsZero() {
			asset.Timestamp = time.Now()
		}
		if asset.Notes == "" {
			asset.Notes = importNote
		} else {
			asset.Notes = asset.Notes + " (" + importNote + ")"
		}
		if err := c.repo.CreateAsset(&asset); err == nil {
			imported++
			// Create price entry at asset timestamp for USD value display
			if price, ok := currentPrices[asset.Symbol]; ok {
				c.repo.CreatePrice(&models.Price{
					Symbol:    asset.Symbol,
					Currency:  "USD",
					Price:     price,
					Timestamp: asset.Timestamp,
				})
			}
			if c.assetCreatedPub != nil {
				assetJSON, _ := json.Marshal(asset)
				c.assetCreatedPub.Publish(assetJSON)
			}
		}
	}

	// Update import log
	importLog.ImportedRows += imported
	importLog.FailedRows = len(rowErrors)
	if len(rowErrors) == 0 {
		importLog.Status = "completed"
		importLog.FailedData = "[]"
	} else {
		importLog.Status = "partial"
		failedDataJSON, _ := json.Marshal(rowErrors)
		importLog.FailedData = string(failedDataJSON)
	}
	c.repo.UpdateImportLog(importLog)

	ctx.JSON(http.StatusOK, ImportResponse{
		ID:       importLog.ID,
		Imported: imported,
		Failed:   len(rowErrors),
		Total:    len(assets),
		Status:   importLog.Status,
		Errors:   rowErrors,
	})
}

// DeleteImportLog deletes an import log
// @Summary Delete import log
// @Description Delete an import log
// @Tags data
// @Param id path int true "Import log ID"
// @Success 204
// @Failure 404 {object} APIError
// @Router /api/imports/{id} [delete]
func (c *Controller) DeleteImportLog(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		badRequest(ctx, "invalid id")
		return
	}

	if err := c.repo.DeleteImportLog(id); err != nil {
		notFound(ctx, "import log not found")
		return
	}

	ctx.Status(http.StatusNoContent)
}

func assetsToCSV(assets []models.Asset) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	w.Write([]string{"symbol", "name", "amount", "transaction_type", "timestamp", "notes"})

	for _, a := range assets {
		w.Write([]string{
			a.Symbol,
			a.Name,
			strconv.FormatFloat(a.Amount, 'f', -1, 64),
			a.TransactionType,
			a.Timestamp.Format(time.RFC3339),
			a.Notes,
		})
	}

	w.Flush()
	return buf.Bytes()
}

func exchangesToCSV(exchanges []models.Exchange) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	w.Write([]string{"from_symbol", "to_symbol", "from_amount", "to_amount", "fee", "fee_currency", "timestamp", "notes"})

	for _, e := range exchanges {
		w.Write([]string{
			e.FromSymbol,
			e.ToSymbol,
			strconv.FormatFloat(e.FromAmount, 'f', -1, 64),
			strconv.FormatFloat(e.ToAmount, 'f', -1, 64),
			strconv.FormatFloat(e.Fee, 'f', -1, 64),
			e.FeeCurrency,
			e.Timestamp.Format(time.RFC3339),
			e.Notes,
		})
	}

	w.Flush()
	return buf.Bytes()
}

func parseAssetsFromCSV(data []byte) ([]models.Asset, []RowError) {
	var assets []models.Asset
	var errors []RowError

	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = ';'
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, []RowError{{Row: 0, Message: "invalid CSV format: " + err.Error()}}
	}

	if len(records) < 2 {
		return nil, []RowError{{Row: 0, Message: "CSV file must have a header and at least one data row"}}
	}

	header := records[0]
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[strings.ToLower(strings.TrimSpace(col))] = i
	}

	for i, row := range records[1:] {
		rowNum := i + 2

		asset := models.Asset{}
		rowData := make(map[string]string)

		if idx, ok := colIndex["symbol"]; ok && idx < len(row) {
			asset.Symbol = strings.ToUpper(strings.TrimSpace(row[idx]))
			rowData["symbol"] = asset.Symbol
		}
		if idx, ok := colIndex["name"]; ok && idx < len(row) {
			asset.Name = strings.TrimSpace(row[idx])
			rowData["name"] = asset.Name
		}
		if idx, ok := colIndex["amount"]; ok && idx < len(row) {
			val := strings.TrimSpace(row[idx])
			rowData["amount"] = val
			if amt, err := strconv.ParseFloat(val, 64); err == nil {
				asset.Amount = amt
			}
		}
		if idx, ok := colIndex["transaction_type"]; ok && idx < len(row) {
			asset.TransactionType = strings.ToLower(strings.TrimSpace(row[idx]))
			rowData["transaction_type"] = asset.TransactionType
		}
		if idx, ok := colIndex["timestamp"]; ok && idx < len(row) {
			val := strings.TrimSpace(row[idx])
			rowData["timestamp"] = val
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				asset.Timestamp = t
			}
		}
		if idx, ok := colIndex["notes"]; ok && idx < len(row) {
			asset.Notes = strings.TrimSpace(row[idx])
			rowData["notes"] = asset.Notes
		}

		if err := validateAssetFields(&asset); err != nil {
			rowDataJSON, _ := json.Marshal(rowData)
			errors = append(errors, RowError{
				Row:     rowNum,
				Data:    rowDataJSON,
				Message: err.Error(),
			})
			continue
		}

		assets = append(assets, asset)
	}

	return assets, errors
}

func parseAssetsFromJSON(data []byte) ([]models.Asset, []RowError) {
	var rawAssets []json.RawMessage
	if err := json.Unmarshal(data, &rawAssets); err != nil {
		return nil, []RowError{{Row: 0, Message: "invalid JSON format: " + err.Error()}}
	}

	var assets []models.Asset
	var errors []RowError

	for i, raw := range rawAssets {
		rowNum := i + 1

		var asset models.Asset
		if err := json.Unmarshal(raw, &asset); err != nil {
			errors = append(errors, RowError{
				Row:     rowNum,
				Data:    raw,
				Message: "invalid asset format: " + err.Error(),
			})
			continue
		}

		asset.Symbol = strings.ToUpper(strings.TrimSpace(asset.Symbol))

		if err := validateAssetFields(&asset); err != nil {
			errors = append(errors, RowError{
				Row:     rowNum,
				Data:    raw,
				Message: err.Error(),
			})
			continue
		}

		assets = append(assets, asset)
	}

	return assets, errors
}

func validateAssetFields(asset *models.Asset) error {
	if asset.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if asset.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	txType := strings.ToLower(asset.TransactionType)
	if txType != "deposit" && txType != "withdrawal" && txType != "withdraw" {
		return fmt.Errorf("transaction_type must be deposit or withdrawal")
	}
	if txType == "withdraw" {
		asset.TransactionType = "withdrawal"
	}
	return nil
}

func getSupportedSymbolsWithPrices() (map[string]struct{}, map[string]float64) {
	symbols := make(map[string]struct{})
	priceMap := make(map[string]float64)

	fetcher := prices.NewPriceService()
	allPrices, err := fetcher.FetchAll()
	if err != nil {
		return symbols, priceMap
	}

	for _, p := range allPrices {
		symbol := strings.ToUpper(p.Asset.Symbol)
		symbols[symbol] = struct{}{}
		priceMap[symbol] = p.Value
	}

	return symbols, priceMap
}
