package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHealthWithoutDatabase(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	res := httptest.NewRecorder()
	healthHandler(nil).ServeHTTP(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", res.Code)
	}
}

func TestPublicShareHandlerReturnsOnlyEnabledFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create mock database: %v", err)
	}
	defer db.Close()
	query := regexp.QuoteMeta(`SELECT name,COALESCE(completion_photo_data,cover_data,''),completion_note,actual_minutes,work_tags,yarn_names,tool_names,share_show_materials,share_show_note FROM projects WHERE share_token=? AND status='completed'`)
	mock.ExpectQuery(query).WithArgs("active-token").WillReturnRows(sqlmock.NewRows([]string{"name", "photo", "note", "minutes", "tags", "yarns", "tools", "show_materials", "show_note"}).AddRow("春日小熊", "data:image/png;base64,abc", "留给自己的春天", 180, `["玩偶","送礼"]`, `["棉线"]`, `["3.0 mm 钩针"]`, false, false))
	req := httptest.NewRequest(http.MethodGet, "/api/share/active-token", nil)
	req.SetPathValue("token", "active-token")
	res := httptest.NewRecorder()
	publicShareHandler(db).ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct { Data struct { Name string `json:"name"`; CompletionNote string `json:"completionNote"`; WorkTags []string `json:"workTags"`; YarnNames []string `json:"yarnNames"`; ToolNames []string `json:"toolNames"`; ShowMaterials bool `json:"showMaterials"`; ShowNote bool `json:"showNote"` } `json:"data"` }
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Name != "春日小熊" || len(body.Data.WorkTags) != 2 {
		t.Fatalf("unexpected public data: %#v", body.Data)
	}
	if body.Data.CompletionNote != "" || len(body.Data.YarnNames) != 0 || len(body.Data.ToolNames) != 0 || body.Data.ShowMaterials || body.Data.ShowNote {
		t.Fatalf("disabled fields leaked into public response: %#v", body.Data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPublicShareHandlerReturnsNotFoundAfterLinkClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create mock database: %v", err)
	}
	defer db.Close()
	query := regexp.QuoteMeta(`SELECT name,COALESCE(completion_photo_data,cover_data,''),completion_note,actual_minutes,work_tags,yarn_names,tool_names,share_show_materials,share_show_note FROM projects WHERE share_token=? AND status='completed'`)
	req := httptest.NewRequest(http.MethodGet, "/api/share/closed-token", nil)
	req.SetPathValue("token", "closed-token")
	res := httptest.NewRecorder()
	mock.ExpectQuery(query).WithArgs("closed-token").WillReturnRows(sqlmock.NewRows([]string{"name", "photo", "note", "minutes", "tags", "yarns", "tools", "show_materials", "show_note"}))
	publicShareHandler(db).ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after closed or missing share link, got %d", res.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCreateUser(t *testing.T) {
	valid := createUserRequest{Name: "林小满", Email: "lin@example.com", Password: "password123", Role: "user"}
	if err := validateCreateUser(valid); err != nil {
		t.Fatalf("expected valid input, got %v", err)
	}
	invalid := valid
	invalid.Role = "owner"
	if err := validateCreateUser(invalid); err == nil {
		t.Fatal("expected invalid role to fail")
	}
}

func TestCalculateCircle(t *testing.T) {
	result, err := calculateCircle(circleCalculatorRequest{StartStitches: 6, Rounds: 3, IncreasePerRound: 6})
	if err != nil {
		t.Fatalf("expected calculation to succeed: %v", err)
	}
	if result["finalStitches"] != 18 {
		t.Fatalf("expected 18 stitches, got %#v", result["finalStitches"])
	}
}

func TestCalculateSize(t *testing.T) {
	result, err := calculateSize(sizeCalculatorRequest{StitchesPer10CM: 20, RowsPer10CM: 18, WidthCM: 30, HeightCM: 25})
	if err != nil {
		t.Fatalf("expected calculation to succeed: %v", err)
	}
	if result["castOnStitches"] != 60 || result["rows"] != 45 {
		t.Fatalf("unexpected size result: %#v", result)
	}
}

func TestValidateCreateYarn(t *testing.T) {
	valid := createYarnRequest{Name: "云朵棉", Brand: "线禾", Color: "奶油白", Material: "棉", Thickness: "DK", Plies: 2}
	if err := validateCreateYarn(valid); err != nil {
		t.Fatalf("expected valid yarn, got %v", err)
	}
	invalid := valid
	invalid.Plies = 0
	if err := validateCreateYarn(invalid); err == nil {
		t.Fatal("expected invalid plies to fail")
	}
}

func TestValidateCreateTool(t *testing.T) {
	valid := createToolRequest{Name: "樱花钩针", Type: "钩针", Thickness: "4.0 mm"}
	if err := validateCreateTool(valid); err != nil {
		t.Fatalf("expected valid tool, got %v", err)
	}
	invalid := valid
	invalid.Name = ""
	if err := validateCreateTool(invalid); err == nil {
		t.Fatal("expected empty tool name to fail")
	}
}

func TestValidateShoppingItem(t *testing.T) {
	valid := shoppingItemRequest{YarnName: "云朵棉奶油白", TargetGrams: 100, EstimatedPrice: 28.5, Priority: "high", Status: "pending"}
	if err := validateShoppingItem(valid); err != nil {
		t.Fatalf("expected valid shopping item, got %v", err)
	}
	invalid := valid
	invalid.TargetGrams = 0
	if err := validateShoppingItem(invalid); err == nil {
		t.Fatal("expected zero grams to fail")
	}
}

func TestProjectScheduleBounds(t *testing.T) {
	if !validWeeklyFocusMinutes(240) {
		t.Fatal("expected a normal weekly focus target to be accepted")
	}
	if validWeeklyFocusMinutes(10081) || validWeeklyFocusMinutes(-1) {
		t.Fatal("expected out-of-range weekly focus targets to be rejected")
	}
}
