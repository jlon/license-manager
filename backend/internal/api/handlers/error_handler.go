package handlers

import (
	"net/http"

	"license-manager/internal/models"
	"license-manager/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// handleI18nError writes the shared API error response used by handlers.
func handleI18nError(c *gin.Context, err error, lang string) {
	if i18nErr, ok := err.(*i18n.I18nError); ok {
		c.JSON(i18nErr.HttpCode, models.ErrorResponse{
			Code:      i18nErr.Code,
			Message:   i18nErr.Message,
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Code:      "900004",
		Message:   i18n.GetI18nErrorMessage("900004", lang),
		Timestamp: getCurrentTimestamp(),
	})
}
