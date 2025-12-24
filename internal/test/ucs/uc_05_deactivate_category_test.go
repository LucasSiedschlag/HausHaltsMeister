package ucs

import (
	"encoding/json"
	"fmt"
	std_http "net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC05_DeactivateCategory(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Setup Clean Architecture Stack
	repo := postgres.NewCategoryRepository(db.Pool)
	svc := category.NewService(repo)
	handler := http.NewCategoryHandler(svc)

	// Setup Echo
	e := echo.New()
	http.RegisterCategoryRoutes(e, handler)
	client := harness.NewHTTPClient(e)

	// Step 1: Create a Category
	payload := map[string]interface{}{
		"name":      "To Be Deactivated",
		"direction": "OUT",
	}
	createRec := client.Request(t, "POST", "/categories", payload)
	require.Equal(t, std_http.StatusCreated, createRec.Code)

	var createdCat map[string]interface{}
	err := json.Unmarshal(createRec.Body.Bytes(), &createdCat)
	require.NoError(t, err)
	catID := int(createdCat["id"].(float64))

	// Step 2: Deactivate it
	deactivatePath := fmt.Sprintf("/categories/%d/deactivate", catID)
	deactivateRec := client.Request(t, "PATCH", deactivatePath, nil)
	require.Equal(t, std_http.StatusOK, deactivateRec.Code)

	// Step 3: Verify it is NOT in active list
	listActiveRec := client.Request(t, "GET", "/categories?active=true", nil)
	require.Equal(t, std_http.StatusOK, listActiveRec.Code)

	var activeList []map[string]interface{}
	err = json.Unmarshal(listActiveRec.Body.Bytes(), &activeList)
	require.NoError(t, err)

	for _, c := range activeList {
		if int(c["id"].(float64)) == catID {
			t.Fatalf("Category %d should NOT be in active list", catID)
		}
	}

	// Step 4: Verify it IS in full list (history integrity)
	listAllRec := client.Request(t, "GET", "/categories?active=false", nil) // active=false means "active only = false" -> all
	require.Equal(t, std_http.StatusOK, listAllRec.Code)

	var allList []map[string]interface{}
	err = json.Unmarshal(listAllRec.Body.Bytes(), &allList)
	require.NoError(t, err)

	found := false
	for _, c := range allList {
		if int(c["id"].(float64)) == catID {
			found = true
			assert.False(t, c["is_active"].(bool), "Category should be inactive in full list")
			break
		}
	}
	assert.True(t, found, "Category should be present in full list")
}
