package controllers

import (
	"net/http"
	"strconv"

	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	txRepo *repo.TransactionRepository
}

// Option is the functional options pattern for TransactionController
type TransactionControllerOption func(*TransactionController) error

// NewTransaction creates a new transaction controller with options
func NewTransaction(opts ...TransactionControllerOption) (*TransactionController, error) {
	c := &TransactionController{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithTransactionRepo sets the transaction repository
func WithTransactionRepo(txRepo *repo.TransactionRepository) TransactionControllerOption {
	return func(c *TransactionController) error {
		if txRepo == nil {
			return ErrNilRepository
		}
		c.txRepo = txRepo
		return nil
	}
}

func (ctrl *TransactionController) List(c *gin.Context) {
	transactions, err := ctrl.txRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.HTML(http.StatusOK, "transactions/index.html", gin.H{
		"title":        "Transactions",
		"transactions": transactions,
	})
}

func (ctrl *TransactionController) New(c *gin.Context) {
	c.HTML(http.StatusOK, "transactions/new.html", gin.H{
		"title": "New Transaction",
	})
}

func (ctrl *TransactionController) Create(c *gin.Context) {
	var tx repo.Transaction
	if err := c.ShouldBind(&tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := ctrl.txRepo.Create(&tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/transactions")
}

func (ctrl *TransactionController) Show(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	tx, err := ctrl.txRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.HTML(http.StatusOK, "transactions/show.html", gin.H{
		"title":       "Transaction Details",
		"transaction": tx,
	})
}

func (ctrl *TransactionController) Edit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	tx, err := ctrl.txRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.HTML(http.StatusOK, "transactions/edit.html", gin.H{
		"title":       "Edit Transaction",
		"transaction": tx,
	})
}

func (ctrl *TransactionController) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	var tx repo.Transaction
	if err := c.ShouldBind(&tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	tx.ID = id
	if err := ctrl.txRepo.Update(&tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/transactions/"+strconv.FormatInt(id, 10))
}

func (ctrl *TransactionController) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	if err := ctrl.txRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/transactions")
}
