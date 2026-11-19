package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

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
	var body struct {
		Data struct {
			Name           string   `json:"name"`
			CompletionNote string   `json:"completionNote"`
			WorkTags       []string `json:"workTags"`
			YarnNames      []string `json:"yarnNames"`
			ToolNames      []string `json:"toolNames"`
			ShowMaterials  bool     `json:"showMaterials"`
			ShowNote       bool     `json:"showNote"`
		} `json:"data"`
	}
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

type memoryStore struct {
	objects map[string][]byte
	puts    int
}

func (store *memoryStore) Put(_ context.Context, key string, body []byte, _ string) error {
	if store.objects == nil {
		store.objects = map[string][]byte{}
	}
	store.objects[key] = append([]byte(nil), body...)
	store.puts++
	return nil
}
func (store *memoryStore) Get(_ context.Context, key string) ([]byte, error) {
	body, ok := store.objects[key]
	if !ok {
		return nil, os.ErrNotExist
	}
	return body, nil
}
func (store *memoryStore) Delete(_ context.Context, key string) error {
	delete(store.objects, key)
	return nil
}

func TestDecodeDataURL(t *testing.T) {
	mimeType, body, err := decodeDataURL("data:application/pdf;base64,JVBERi0xLjQK")
	if err != nil {
		t.Fatalf("expected data url to decode: %v", err)
	}
	if mimeType != "application/pdf" || string(body) != "%PDF-1.4\n" {
		t.Fatalf("unexpected payload: %q %q", mimeType, body)
	}
	if _, _, err := decodeDataURL("JVBERi0xLjQK"); err == nil {
		t.Fatal("expected raw base64 without data url prefix to fail")
	}
	if _, _, err := decodeDataURL("data:application/pdf,JVBERi0xLjQK"); err == nil {
		t.Fatal("expected non base64 data url to fail")
	}
}

func TestPatternObjectKeyAndMimeRules(t *testing.T) {
	if key := patternObjectKey(7, "application/pdf"); key != "pattern-7.pdf" {
		t.Fatalf("unexpected pdf key: %s", key)
	}
	if key := patternObjectKey(7, "image/webp"); key != "pattern-7.webp" {
		t.Fatalf("unexpected webp key: %s", key)
	}
	if !patternMimeAllowed("image", "image/png") || !patternMimeAllowed("pdf", "application/pdf") {
		t.Fatal("expected supported pattern types to be accepted")
	}
	if patternMimeAllowed("pdf", "image/png") || patternMimeAllowed("image", "application/zip") {
		t.Fatal("expected mismatched pattern types to be rejected")
	}
}

func TestReadPatternObjectServesStoredObject(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create mock database: %v", err)
	}
	defer db.Close()
	store := &memoryStore{objects: map[string][]byte{"pattern-3.pdf": []byte("%PDF-3")}}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id,COALESCE(object_key,''),mime_type,COALESCE(file_data,'') FROM pattern_files WHERE id=?`)).WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "object_key", "mime_type", "file_data"}).AddRow(11, "pattern-3.pdf", "application/pdf", ""))
	body, mimeType, err := readPatternObject(context.Background(), db, store, 11, 3)
	if err != nil {
		t.Fatalf("expected stored pattern to load: %v", err)
	}
	if string(body) != "%PDF-3" || mimeType != "application/pdf" || store.puts != 0 {
		t.Fatalf("unexpected read result: %q %q puts=%d", body, mimeType, store.puts)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReadPatternObjectMigratesLegacyRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create mock database: %v", err)
	}
	defer db.Close()
	store := &memoryStore{}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id,COALESCE(object_key,''),mime_type,COALESCE(file_data,'') FROM pattern_files WHERE id=?`)).WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "object_key", "mime_type", "file_data"}).AddRow(2, "", "application/octet-stream", "data:image/png;base64,aGVsbG8="))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE pattern_files SET object_key=?,mime_type=?,size_bytes=?,file_data=NULL WHERE id=?`)).
		WithArgs("pattern-5.png", "image/png", 5, int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	body, mimeType, err := readPatternObject(context.Background(), db, store, 2, 5)
	if err != nil {
		t.Fatalf("expected legacy pattern to migrate: %v", err)
	}
	if string(body) != "hello" || mimeType != "image/png" {
		t.Fatalf("unexpected migrated payload: %q %q", body, mimeType)
	}
	if string(store.objects["pattern-5.png"]) != "hello" {
		t.Fatalf("expected object storage write, got %#v", store.objects)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReadPatternObjectRejectsOtherUsersFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create mock database: %v", err)
	}
	defer db.Close()
	store := &memoryStore{objects: map[string][]byte{"pattern-9.pdf": []byte("%PDF-9")}}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id,COALESCE(object_key,''),mime_type,COALESCE(file_data,'') FROM pattern_files WHERE id=?`)).WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "object_key", "mime_type", "file_data"}).AddRow(4, "pattern-9.pdf", "application/pdf", ""))
	if _, _, err := readPatternObject(context.Background(), db, store, 8, 9); !errors.Is(err, errPatternNotFound) {
		t.Fatalf("expected ownership error, got %v", err)
	}
}

func TestValidateReminderDelivery(t *testing.T) {
	channel, remindAt, err := validateReminderDelivery("", nil)
	if err != nil || channel != reminderChannelInApp || remindAt != nil {
		t.Fatalf("expected in-app default, got %q %v %v", channel, remindAt, err)
	}
	if _, _, err := validateReminderDelivery("email", nil); err == nil {
		t.Fatal("expected email reminders without a time to fail")
	}
	if _, _, err := validateReminderDelivery("sms", nil); err == nil {
		t.Fatal("expected unknown channel to fail")
	}
	due := time.Now().Add(30 * time.Minute)
	channel, remindAt, err = validateReminderDelivery("email", &due)
	if err != nil || channel != reminderChannelEmail || remindAt == nil || remindAt.Location() != time.UTC {
		t.Fatalf("expected a normalized email reminder, got %q %v %v", channel, remindAt, err)
	}
	tooFar := time.Now().Add(3 * 365 * 24 * time.Hour)
	if _, _, err := validateReminderDelivery("email", &tooFar); err == nil {
		t.Fatal("expected out-of-window remindAt to fail")
	}
}

func TestReminderEmailMessage(t *testing.T) {
	subject, body := reminderEmailMessage("林小满", "春日小熊", 24, 18, 60, "在这里换浅色线", "https://line.example.com/", 7)
	if !strings.Contains(subject, "春日小熊") || !strings.Contains(subject, "第 24 行") {
		t.Fatalf("unexpected subject: %q", subject)
	}
	for _, want := range []string{"林小满", "在这里换浅色线", "第 18 / 60 行", "https://line.example.com/?project=7"} {
		if !strings.Contains(body, want) {
			t.Fatalf("email body missing %q: %s", want, body)
		}
	}
	if _, body = reminderEmailMessage("", "春日小熊", 1, 0, 0, "起针", "", 1); !strings.Contains(body, "钩织手") || !strings.Contains(body, "当前进度：第 0 行") {
		t.Fatalf("expected fallback name and plain progress, got %s", body)
	}
}

func TestDispatchDueRemindersSendsEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create mock database: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(`SELECT r.id,r.project_id,u.id,p.name`).
		WithArgs(reminderChannelEmail, reminderMaxAttempts, reminderDispatchBatch).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "user_id", "name", "target_row", "current_row", "total_rows", "message", "attempts", "email", "user_name"}).
			AddRow(11, 7, 3, "春日小熊", 24, 18, 60, "换浅色线", 0, "lin@example.com", "林小满"))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE project_reminders SET attempts=attempts+1,email_sent_at=NOW(),last_error=NULL WHERE id=?`)).
		WithArgs(int64(11)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'reminder_emailed',?)`)).
		WithArgs(int64(7), "第 24 行提醒已发送邮件").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO audit_events (user_id,action,resource_type,resource_id,details) VALUES (?,?,?,?,?)`)).
		WithArgs(int64(3), "reminder.email_sent", "project_reminder", int64(11), []byte(`{"projectId":7}`)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	var to, subject, body string
	sent, err := dispatchDueReminders(context.Background(), db, config{appBaseURL: "http://localhost:5173"}, func(_ context.Context, recipient, mailSubject, mailBody string) error {
		to, subject, body = recipient, mailSubject, mailBody
		return nil
	})
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if sent != 1 || to != "lin@example.com" || !strings.Contains(subject, "春日小熊") || !strings.Contains(body, "换浅色线") {
		t.Fatalf("unexpected dispatch result: sent=%d to=%q subject=%q", sent, to, subject)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDispatchDueRemindersRecordsFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create mock database: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(`SELECT r.id,r.project_id,u.id,p.name`).
		WithArgs(reminderChannelEmail, reminderMaxAttempts, reminderDispatchBatch).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "user_id", "name", "target_row", "current_row", "total_rows", "message", "attempts", "email", "user_name"}).
			AddRow(12, 7, 3, "春日小熊", 30, 18, 60, "收针", 1, "lin@example.com", "林小满"))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE project_reminders SET attempts=attempts+1,last_error=? WHERE id=?`)).
		WithArgs("smtp 550 mailbox unavailable", int64(12)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO audit_events (user_id,action,resource_type,resource_id,details) VALUES (?,?,?,?,?)`)).
		WithArgs(int64(3), "reminder.email_failed", "project_reminder", int64(12), []byte(`{"projectId":7,"reason":"smtp 550 mailbox unavailable"}`)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	sent, err := dispatchDueReminders(context.Background(), db, config{}, func(context.Context, string, string, string) error {
		return errors.New("smtp 550 mailbox unavailable")
	})
	if err != nil {
		t.Fatalf("dispatch should not fail the batch: %v", err)
	}
	if sent != 0 {
		t.Fatalf("expected no email to be marked sent, got %d", sent)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSplitMailFrom(t *testing.T) {
	cases := []struct{ in, display, address string }{
		{"Line <no-reply@line.local>", "Line", "no-reply@line.local"},
		{"no-reply@line.local", "", "no-reply@line.local"},
		{"  \"线禾\" <no-reply@line.local>  ", "线禾", "no-reply@line.local"},
		{"Line <>", "", "Line <>"},
	}
	for _, tc := range cases {
		display, address := splitMailFrom(tc.in)
		if display != tc.display || address != tc.address {
			t.Fatalf("splitMailFrom(%q) = %q, %q; want %q, %q", tc.in, display, address, tc.display, tc.address)
		}
	}
}
