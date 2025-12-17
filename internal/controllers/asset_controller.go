package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AssetController struct{}

// Option is the functional options pattern for AssetController
type AssetControllerOption func(*AssetController) error

// NewAsset creates a new asset controller with options
func NewAsset(opts ...AssetControllerOption) (*AssetController, error) {
	c := &AssetController{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (ctrl *AssetController) List(c *gin.Context) {
	c.HTML(http.StatusOK, "assets/index.html", gin.H{
		"title": "Assets",
	})
}

func (ctrl *AssetController) New(c *gin.Context) {
	c.HTML(http.StatusOK, "assets/new.html", gin.H{
		"title": "New Asset",
	})
}

func (ctrl *AssetController) Create(c *gin.Context) {
	// TODO: Implement asset creation
	c.Redirect(http.StatusSeeOther, "/assets")
}

func (ctrl *AssetController) Show(c *gin.Context) {
	id := c.Param("id")
	c.HTML(http.StatusOK, "assets/show.html", gin.H{
		"title": "Asset Details",
		"id":    id,
	})
}

func (ctrl *AssetController) Edit(c *gin.Context) {
	id := c.Param("id")
	c.HTML(http.StatusOK, "assets/edit.html", gin.H{
		"title": "Edit Asset",
		"id":    id,
	})
}

func (ctrl *AssetController) Update(c *gin.Context) {
	id := c.Param("id")
	// TODO: Implement asset update
	c.Redirect(http.StatusSeeOther, "/assets/"+id)
}

func (ctrl *AssetController) Delete(c *gin.Context) {
	// TODO: Implement asset deletion
	c.Redirect(http.StatusSeeOther, "/assets")
}
