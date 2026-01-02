package handler

import (
	"net/http"

	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
)

type DataHandler struct {
	renderer *Renderer
	repo     *repo.Repository
}

func NewDataHandler(renderer *Renderer, repository *repo.Repository) *DataHandler {
	return &DataHandler{
		renderer: renderer,
		repo:     repository,
	}
}

type DataPageData struct {
	Title      string
	PageTitle  string
	ActivePage string
	Imports    []ImportLogView
}

type ImportLogView struct {
	ID           int64
	Filename     string
	Format       string
	EntityType   string
	TotalRows    int
	ImportedRows int
	FailedRows   int
	Status       string
	CreatedAt    string
	HasErrors    bool
}

func (h *DataHandler) Index(c *gin.Context) {
	logs, _ := h.repo.ListImportLogs()

	var imports []ImportLogView
	for _, log := range logs {
		imports = append(imports, ImportLogView{
			ID:           log.ID,
			Filename:     log.Filename,
			Format:       log.Format,
			EntityType:   log.EntityType,
			TotalRows:    log.TotalRows,
			ImportedRows: log.ImportedRows,
			FailedRows:   log.FailedRows,
			Status:       log.Status,
			CreatedAt:    log.CreatedAt.Format("2006-01-02 15:04"),
			HasErrors:    log.FailedRows > 0,
		})
	}

	data := DataPageData{
		Title:      "Data",
		PageTitle:  "Data Management",
		ActivePage: "data",
		Imports:    imports,
	}
	h.renderer.HTML(c, http.StatusOK, "data", data)
}

func (h *DataHandler) ImportHistory(c *gin.Context) {
	logs, _ := h.repo.ListImportLogs()

	var imports []ImportLogView
	for _, log := range logs {
		imports = append(imports, ImportLogView{
			ID:           log.ID,
			Filename:     log.Filename,
			Format:       log.Format,
			EntityType:   log.EntityType,
			TotalRows:    log.TotalRows,
			ImportedRows: log.ImportedRows,
			FailedRows:   log.FailedRows,
			Status:       log.Status,
			CreatedAt:    log.CreatedAt.Format("2006-01-02 15:04"),
			HasErrors:    log.FailedRows > 0,
		})
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "data_import_history.html", gin.H{
		"Imports": imports,
	})
}
