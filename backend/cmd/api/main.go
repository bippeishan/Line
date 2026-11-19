package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jung-kurt/gofpdf"
	"golang.org/x/crypto/bcrypt"
)

type config struct {
	port                   string
	databaseURL            string
	bootstrapAdminPassword string
	openAIAPIKey           string
	openAIModel            string
	openAIBaseURL          string
	appBaseURL             string
	smtpHost               string
	smtpPort               string
	smtpUsername           string
	smtpPassword           string
	smtpFrom               string
	mediaRoot              string
	mediaDriver            string
	s3Bucket               string
	s3Region               string
	s3Endpoint             string
	s3AccessKey            string
	s3SecretKey            string
	s3ForcePathStyle       bool
}

type objectStore interface {
	Put(context.Context, string, []byte, string) error
	Get(context.Context, string) ([]byte, error)
	Delete(context.Context, string) error
}
type localObjectStore struct{ root string }
type s3ObjectStore struct {
	client *s3.Client
	bucket string
}

type paletteTemplate struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Style       string    `json:"style"`
	Occasion    string    `json:"occasion"`
	Description string    `json:"description"`
	Colors      []string  `json:"colors"`
	IsBuiltin   bool      `json:"isBuiltin"`
	CreatedAt   time.Time `json:"createdAt"`
}
type createPaletteTemplateRequest struct {
	Name        string   `json:"name"`
	Style       string   `json:"style"`
	Occasion    string   `json:"occasion"`
	Description string   `json:"description"`
	Colors      []string `json:"colors"`
}
type pattern struct {
	ID              int64  `json:"id"`
	Title           string `json:"title"`
	Category        string `json:"category"`
	Difficulty      string `json:"difficulty"`
	Yarn            string `json:"yarn"`
	HookSize        string `json:"hookSize"`
	DurationMinutes int    `json:"durationMinutes"`
}

type user struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenRequest struct {
	Token string `json:"token"`
}
type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type circleCalculatorRequest struct {
	StartStitches    int `json:"startStitches"`
	Rounds           int `json:"rounds"`
	IncreasePerRound int `json:"increasePerRound"`
}

type sizeCalculatorRequest struct {
	StitchesPer10CM float64 `json:"stitchesPer10cm"`
	RowsPer10CM     float64 `json:"rowsPer10cm"`
	WidthCM         float64 `json:"widthCm"`
	HeightCM        float64 `json:"heightCm"`
}

type yarnCalculatorRequest struct {
	MetersPer100G float64 `json:"metersPer100g"`
	PatternMeters float64 `json:"patternMeters"`
	ReserveRate   float64 `json:"reserveRate"`
}
type contentAskRequest struct {
	Question      string `json:"question"`
	PatternFileID *int64 `json:"patternFileId"`
}
type saveCalculationRequest struct {
	Mode      string          `json:"mode"`
	ProjectID *int64          `json:"projectId"`
	Input     json.RawMessage `json:"input"`
}
type calculationRecord struct {
	ID        int64           `json:"id"`
	ProjectID *int64          `json:"projectId"`
	Mode      string          `json:"mode"`
	Input     json.RawMessage `json:"input"`
	Result    json.RawMessage `json:"result"`
	Formula   string          `json:"formula"`
	CreatedAt time.Time       `json:"createdAt"`
}

type yarn struct {
	ID             int64     `json:"id"`
	PhotoData      string    `json:"photoData"`
	Name           string    `json:"name"`
	Brand          string    `json:"brand"`
	BatchCode      string    `json:"batchCode"`
	Color          string    `json:"color"`
	ColorHex       string    `json:"colorHex"`
	Material       string    `json:"material"`
	Thickness      string    `json:"thickness"`
	Plies          int       `json:"plies"`
	RemainingGrams float64   `json:"remainingGrams"`
	UnitPrice      float64   `json:"unitPrice"`
	CreatedAt      time.Time `json:"createdAt"`
}

type createYarnRequest struct {
	PhotoData      string  `json:"photoData"`
	Name           string  `json:"name"`
	Brand          string  `json:"brand"`
	BatchCode      string  `json:"batchCode"`
	Color          string  `json:"color"`
	ColorHex       string  `json:"colorHex"`
	Material       string  `json:"material"`
	Thickness      string  `json:"thickness"`
	Plies          int     `json:"plies"`
	RemainingGrams float64 `json:"remainingGrams"`
	UnitPrice      float64 `json:"unitPrice"`
}

type shoppingItem struct {
	ID             int64     `json:"id"`
	YarnID         *int64    `json:"yarnId"`
	YarnName       string    `json:"yarnName"`
	TargetGrams    float64   `json:"targetGrams"`
	EstimatedPrice float64   `json:"estimatedPrice"`
	Priority       string    `json:"priority"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}
type shoppingItemRequest struct {
	YarnID         *int64  `json:"yarnId"`
	YarnName       string  `json:"yarnName"`
	TargetGrams    float64 `json:"targetGrams"`
	EstimatedPrice float64 `json:"estimatedPrice"`
	Priority       string  `json:"priority"`
	Status         string  `json:"status"`
}
type inventoryShortage struct {
	YarnID         int64    `json:"yarnId"`
	YarnName       string   `json:"yarnName"`
	RemainingGrams float64  `json:"remainingGrams"`
	SuggestedGrams float64  `json:"suggestedGrams"`
	ProjectNames   []string `json:"projectNames"`
}
type projectReadinessItem struct {
	Key     string  `json:"key"`
	Label   string  `json:"label"`
	Ready   bool    `json:"ready"`
	Detail  string  `json:"detail"`
	YarnID  int64   `json:"yarnId,omitempty"`
	Missing float64 `json:"missingGrams,omitempty"`
}

type tool struct {
	ID        int64     `json:"id"`
	PhotoData string    `json:"photoData"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Thickness string    `json:"thickness"`
	CreatedAt time.Time `json:"createdAt"`
}

type createToolRequest struct {
	PhotoData string `json:"photoData"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Thickness string `json:"thickness"`
}

type project struct {
	ID                  int64      `json:"id"`
	CoverData           string     `json:"coverData"`
	CompletionPhotoData string     `json:"completionPhotoData"`
	Name                string     `json:"name"`
	PatternName         string     `json:"patternName"`
	PatternFileID       *int64     `json:"patternFileId"`
	TargetSize          string     `json:"targetSize"`
	HookSize            string     `json:"hookSize"`
	YarnNames           []string   `json:"yarnNames"`
	ToolNames           []string   `json:"toolNames"`
	Palette             []string   `json:"palette"`
	WorkTags            []string   `json:"workTags"`
	ShareToken          string     `json:"shareToken"`
	ShareShowMaterials  bool       `json:"shareShowMaterials"`
	ShareShowNote       bool       `json:"shareShowNote"`
	Status              string     `json:"status"`
	Version             int        `json:"version"`
	CurrentRow          int        `json:"currentRow"`
	TotalRows           int        `json:"totalRows"`
	BreakpointNote      string     `json:"breakpointNote"`
	CompletedAt         *time.Time `json:"completedAt"`
	CompletionNote      string     `json:"completionNote"`
	ActualMinutes       int        `json:"actualMinutes"`
	TargetDate          *time.Time `json:"targetDate"`
	WeeklyFocusMinutes  int        `json:"weeklyFocusMinutes"`
	CreatedAt           time.Time  `json:"createdAt"`
	MaterialCost        float64    `json:"materialCost"`
}
type createProjectRequest struct {
	CoverData          string     `json:"coverData"`
	Name               string     `json:"name"`
	PatternName        string     `json:"patternName"`
	PatternFileID      *int64     `json:"patternFileId"`
	TargetSize         string     `json:"targetSize"`
	HookSize           string     `json:"hookSize"`
	YarnNames          []string   `json:"yarnNames"`
	ToolNames          []string   `json:"toolNames"`
	WorkTags           []string   `json:"workTags"`
	Palette            []string   `json:"palette"`
	TotalRows          int        `json:"totalRows"`
	BreakpointNote     string     `json:"breakpointNote"`
	TargetDate         *time.Time `json:"targetDate"`
	WeeklyFocusMinutes int        `json:"weeklyFocusMinutes"`
}
type updateProjectRequest struct {
	CoverData           string     `json:"coverData"`
	CompletionPhotoData string     `json:"completionPhotoData"`
	Name                string     `json:"name"`
	PatternName         string     `json:"patternName"`
	PatternFileID       *int64     `json:"patternFileId"`
	TargetSize          string     `json:"targetSize"`
	HookSize            string     `json:"hookSize"`
	YarnNames           []string   `json:"yarnNames"`
	ToolNames           []string   `json:"toolNames"`
	Palette             []string   `json:"palette"`
	Status              string     `json:"status"`
	CurrentRow          int        `json:"currentRow"`
	TotalRows           int        `json:"totalRows"`
	BreakpointNote      string     `json:"breakpointNote"`
	CompletionNote      string     `json:"completionNote"`
	ActualMinutes       int        `json:"actualMinutes"`
	WorkTags            []string   `json:"workTags"`
	TargetDate          *time.Time `json:"targetDate"`
	WeeklyFocusMinutes  int        `json:"weeklyFocusMinutes"`
	Version             int        `json:"version"`
}
type shareProjectRequest struct {
	WorkTags      []string `json:"workTags"`
	ShowMaterials bool     `json:"showMaterials"`
	ShowNote      bool     `json:"showNote"`
}
type patternFile struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	FileType  string    `json:"fileType"`
	MimeType  string    `json:"mimeType"`
	SizeBytes int64     `json:"sizeBytes"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
}
type createPatternFileRequest struct {
	Title    string `json:"title"`
	FileData string `json:"fileData"`
	FileType string `json:"fileType"`
}
type patternAnnotation struct {
	ID         int64     `json:"id"`
	PageNumber int       `json:"pageNumber"`
	RowNumber  int       `json:"rowNumber"`
	X          float64   `json:"x"`
	Y          float64   `json:"y"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"createdAt"`
}
type createPatternAnnotationRequest struct {
	PageNumber int     `json:"pageNumber"`
	RowNumber  int     `json:"rowNumber"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Note       string  `json:"note"`
}
type stitchTerm struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	NameZH      string `json:"nameZh"`
	NameEN      string `json:"nameEn"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Instruction string `json:"instruction"`
	Caution     string `json:"caution"`
	IsFavorite  bool   `json:"isFavorite"`
}
type projectStep struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	IsDone    bool   `json:"isDone"`
	SortOrder int    `json:"sortOrder"`
}
type createProjectStepRequest struct {
	Title string `json:"title"`
}
type updateProjectStepRequest struct {
	Title     string `json:"title"`
	IsDone    bool   `json:"isDone"`
	SortOrder int    `json:"sortOrder"`
}
type projectNote struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	PhotoData string    `json:"photoData"`
	CreatedAt time.Time `json:"createdAt"`
}
type createProjectNoteRequest struct {
	Body      string `json:"body"`
	PhotoData string `json:"photoData"`
}
type projectActivity struct {
	ID        int64     `json:"id"`
	Kind      string    `json:"kind"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}
type projectReminder struct {
	ID          int64      `json:"id"`
	TargetRow   int        `json:"targetRow"`
	Message     string     `json:"message"`
	IsDone      bool       `json:"isDone"`
	RemindAt    *time.Time `json:"remindAt"`
	Channel     string     `json:"channel"`
	EmailSentAt *time.Time `json:"emailSentAt"`
	Attempts    int        `json:"attempts"`
	LastError   string     `json:"lastError"`
}
type createProjectReminderRequest struct {
	TargetRow int        `json:"targetRow"`
	Message   string     `json:"message"`
	RemindAt  *time.Time `json:"remindAt"`
	Channel   string     `json:"channel"`
}
type updateProjectReminderRequest struct {
	TargetRow int        `json:"targetRow"`
	Message   string     `json:"message"`
	IsDone    bool       `json:"isDone"`
	RemindAt  *time.Time `json:"remindAt"`
	Channel   string     `json:"channel"`
}
type upcomingReminder struct {
	ProjectID   int64      `json:"projectId"`
	ProjectName string     `json:"projectName"`
	TargetRow   int        `json:"targetRow"`
	Message     string     `json:"message"`
	RemindAt    *time.Time `json:"remindAt"`
	Channel     string     `json:"channel"`
}
type addUsageRequest struct {
	YarnID int64   `json:"yarnId"`
	Grams  float64 `json:"grams"`
}
type projectYarnUsage struct {
	ID        int64     `json:"id"`
	YarnID    int64     `json:"yarnId"`
	YarnName  string    `json:"yarnName"`
	Grams     float64   `json:"grams"`
	Cost      float64   `json:"cost"`
	CreatedAt time.Time `json:"createdAt"`
}
type projectAccessory struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Quantity  float64   `json:"quantity"`
	Unit      string    `json:"unit"`
	UnitPrice float64   `json:"unitPrice"`
	Cost      float64   `json:"cost"`
	CreatedAt time.Time `json:"createdAt"`
}
type projectYarnPlan struct {
	YarnID         int64   `json:"yarnId"`
	YarnName       string  `json:"yarnName"`
	PlannedGrams   float64 `json:"plannedGrams"`
	UsedGrams      float64 `json:"usedGrams"`
	RemainingGrams float64 `json:"remainingGrams"`
}
type projectYarnPlanRequest struct {
	Plans []projectYarnPlan `json:"plans"`
}
type projectTimer struct {
	ID             int64      `json:"id"`
	StartedAt      time.Time  `json:"startedAt"`
	StoppedAt      *time.Time `json:"stoppedAt"`
	SecondsElapsed int        `json:"secondsElapsed"`
}
type insightFocusDay struct {
	Date    string `json:"date"`
	Minutes int    `json:"minutes"`
}
type insightProject struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Minutes int    `json:"minutes"`
}
type inactiveProject struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	DaysIdle int    `json:"daysIdle"`
}
type createProjectAccessoryRequest struct {
	Name      string  `json:"name"`
	Quantity  float64 `json:"quantity"`
	Unit      string  `json:"unit"`
	UnitPrice float64 `json:"unitPrice"`
}

func main() {
	cfg := config{port: env("PORT", "8080"), databaseURL: os.Getenv("DATABASE_URL"), bootstrapAdminPassword: env("BOOTSTRAP_ADMIN_PASSWORD", "line_admin_2026"), openAIAPIKey: os.Getenv("OPENAI_API_KEY"), openAIModel: env("OPENAI_MODEL", "gpt-4.1-mini"), openAIBaseURL: env("OPENAI_BASE_URL", "https://api.openai.com/v1"), appBaseURL: env("APP_BASE_URL", "http://localhost:5173"), smtpHost: os.Getenv("SMTP_HOST"), smtpPort: env("SMTP_PORT", "587"), smtpUsername: os.Getenv("SMTP_USERNAME"), smtpPassword: os.Getenv("SMTP_PASSWORD"), smtpFrom: env("SMTP_FROM", "Line <no-reply@example.com>"), mediaRoot: env("MEDIA_ROOT", "./data/media"), mediaDriver: env("MEDIA_DRIVER", "local"), s3Bucket: os.Getenv("S3_BUCKET"), s3Region: env("S3_REGION", "us-east-1"), s3Endpoint: os.Getenv("S3_ENDPOINT"), s3AccessKey: os.Getenv("S3_ACCESS_KEY"), s3SecretKey: os.Getenv("S3_SECRET_KEY"), s3ForcePathStyle: strings.EqualFold(env("S3_FORCE_PATH_STYLE", "true"), "true")}
	store, err := newObjectStore(context.Background(), cfg)
	if err != nil {
		log.Fatalf("initialize object storage: %v", err)
	}
	var db *sql.DB
	if cfg.databaseURL != "" {
		var err error
		db, err = sql.Open("mysql", cfg.databaseURL)
		if err != nil {
			log.Fatalf("open database: %v", err)
		}
		defer db.Close()
		if err := ensureBootstrapAdmin(db, cfg.bootstrapAdminPassword); err != nil {
			log.Printf("bootstrap administrator: %v", err)
		}
		backfillPatternFiles(context.Background(), db, store)
		startReminderDispatcher(context.Background(), db, cfg)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", healthHandler(db))
	mux.HandleFunc("GET /api/patterns", patternsHandler(db))
	mux.HandleFunc("GET /api/stitch-terms", stitchTermsHandler(db))
	mux.HandleFunc("POST /api/stitch-terms/{id}/favorite", favoriteStitchTermHandler(db))
	mux.HandleFunc("DELETE /api/stitch-terms/{id}/favorite", unfavoriteStitchTermHandler(db))
	mux.HandleFunc("POST /api/auth/login", loginHandler(db))
	mux.HandleFunc("POST /api/auth/register", registerHandler(db, cfg))
	mux.HandleFunc("POST /api/auth/verify", verifyEmailHandler(db))
	mux.HandleFunc("GET /api/auth/verify", verifyEmailHandler(db))
	mux.HandleFunc("GET /api/auth/verify-email", verifyEmailHandler(db))
	mux.HandleFunc("POST /api/auth/password-reset/request", requestPasswordResetHandler(db, cfg))
	mux.HandleFunc("POST /api/auth/request-reset", requestPasswordResetHandler(db, cfg))
	mux.HandleFunc("POST /api/auth/password-reset/confirm", confirmPasswordResetHandler(db))
	mux.HandleFunc("POST /api/auth/reset-password", confirmPasswordResetHandler(db))
	mux.HandleFunc("POST /api/auth/logout", logoutHandler(db))
	mux.HandleFunc("GET /api/auth/me", currentUserHandler(db))
	mux.HandleFunc("GET /api/auth/permissions", currentPermissionsHandler(db))
	mux.HandleFunc("GET /api/audit", permissionHandler(db, "audit.read", auditEventsHandler(db)))
	mux.HandleFunc("POST /api/media/images", mediaUploadHandler(db, cfg, store))
	mux.HandleFunc("GET /api/media/{id}", mediaGetHandler(db, store))
	mux.HandleFunc("DELETE /api/media/{id}", mediaDeleteHandler(db, cfg, store))
	mux.HandleFunc("POST /api/calculator/{mode}", calculatorHandler(db))
	mux.HandleFunc("POST /api/calculations", saveCalculationHandler(db))
	mux.HandleFunc("POST /api/content/ask", contentAskHandler(db, cfg, store))
	mux.HandleFunc("DELETE /api/calculations/{id}", deleteCalculationHandler(db))
	mux.HandleFunc("GET /api/yarns", permissionHandler(db, "inventory.read", yarnsHandler(db)))
	mux.HandleFunc("POST /api/yarns", permissionHandler(db, "inventory.write", createYarnHandler(db)))
	mux.HandleFunc("DELETE /api/yarns/{id}", permissionHandler(db, "inventory.write", deleteYarnHandler(db)))
	mux.HandleFunc("GET /api/shopping-list", permissionHandler(db, "inventory.read", shoppingListHandler(db)))
	mux.HandleFunc("POST /api/shopping-list", permissionHandler(db, "inventory.write", createShoppingItemHandler(db)))
	mux.HandleFunc("PATCH /api/shopping-list/{id}", permissionHandler(db, "inventory.write", updateShoppingItemHandler(db)))
	mux.HandleFunc("DELETE /api/shopping-list/{id}", permissionHandler(db, "inventory.write", deleteShoppingItemHandler(db)))
	mux.HandleFunc("GET /api/inventory/shortages", inventoryShortagesHandler(db))
	mux.HandleFunc("GET /api/tools", permissionHandler(db, "inventory.read", toolsHandler(db)))
	mux.HandleFunc("POST /api/tools", permissionHandler(db, "inventory.write", createToolHandler(db)))
	mux.HandleFunc("DELETE /api/tools/{id}", permissionHandler(db, "inventory.write", deleteToolHandler(db)))
	mux.HandleFunc("GET /api/projects", permissionHandler(db, "project.read", projectsHandler(db)))
	mux.HandleFunc("POST /api/projects/{id}/share", permissionHandler(db, "project.share", shareProjectHandler(db)))
	mux.HandleFunc("DELETE /api/projects/{id}/share", permissionHandler(db, "project.share", unshareProjectHandler(db)))
	mux.HandleFunc("GET /api/share/{token}", publicShareHandler(db))
	mux.HandleFunc("POST /api/projects", permissionHandler(db, "project.write", createProjectHandler(db)))
	mux.HandleFunc("PATCH /api/projects/{id}", permissionHandler(db, "project.write", updateProjectHandler(db)))
	mux.HandleFunc("DELETE /api/projects/{id}", permissionHandler(db, "project.write", deleteProjectHandler(db)))
	mux.HandleFunc("POST /api/projects/{id}/usages", permissionHandler(db, "project.write", addProjectUsageHandler(db)))
	mux.HandleFunc("GET /api/projects/{id}/materials", permissionHandler(db, "project.read", projectMaterialsHandler(db)))
	mux.HandleFunc("GET /api/projects/{id}/export.pdf", permissionHandler(db, "project.read", projectPDFExportHandler(db)))
	mux.HandleFunc("GET /api/projects/{id}/yarn-plans", permissionHandler(db, "project.read", projectYarnPlansHandler(db)))
	mux.HandleFunc("PUT /api/projects/{id}/yarn-plans", permissionHandler(db, "project.write", saveProjectYarnPlansHandler(db)))
	mux.HandleFunc("GET /api/projects/{id}/readiness", permissionHandler(db, "project.read", projectReadinessHandler(db)))
	mux.HandleFunc("GET /api/projects/{id}/forecast", permissionHandler(db, "project.read", projectForecastHandler(db)))
	mux.HandleFunc("GET /api/projects/{id}/timers", permissionHandler(db, "project.read", projectTimersHandler(db)))
	mux.HandleFunc("POST /api/projects/{id}/timers", permissionHandler(db, "project.write", startProjectTimerHandler(db)))
	mux.HandleFunc("POST /api/projects/{id}/timers/{timerID}/stop", permissionHandler(db, "project.write", stopProjectTimerHandler(db)))
	mux.HandleFunc("GET /api/insights/monthly", monthlyInsightsHandler(db))
	mux.HandleFunc("POST /api/projects/{id}/accessories", createProjectAccessoryHandler(db))
	mux.HandleFunc("DELETE /api/projects/{id}/accessories/{accessoryID}", deleteProjectAccessoryHandler(db))
	mux.HandleFunc("POST /api/palette-suggest", permissionHandler(db, "project.read", paletteSuggestHandler(db)))
	mux.HandleFunc("GET /api/palette-templates", permissionHandler(db, "project.read", paletteTemplatesHandler(db)))
	mux.HandleFunc("POST /api/palette-templates", permissionHandler(db, "project.write", createPaletteTemplateHandler(db)))
	mux.HandleFunc("DELETE /api/palette-templates/{id}", permissionHandler(db, "project.write", deletePaletteTemplateHandler(db)))
	mux.HandleFunc("GET /api/pattern-files", patternFilesHandler(db))
	mux.HandleFunc("POST /api/pattern-files", createPatternFileHandler(db, store))
	mux.HandleFunc("GET /api/pattern-files/{id}/file", patternFileContentHandler(db, store))
	mux.HandleFunc("DELETE /api/pattern-files/{id}", deletePatternFileHandler(db, store))
	mux.HandleFunc("GET /api/pattern-files/{id}/annotations", patternAnnotationsHandler(db))
	mux.HandleFunc("POST /api/pattern-files/{id}/annotations", createPatternAnnotationHandler(db))
	mux.HandleFunc("DELETE /api/pattern-files/{id}/annotations/{annotationID}", deletePatternAnnotationHandler(db))
	mux.HandleFunc("GET /api/projects/{id}/steps", projectStepsHandler(db))
	mux.HandleFunc("POST /api/projects/{id}/steps", createProjectStepHandler(db))
	mux.HandleFunc("PATCH /api/projects/{id}/steps/{stepID}", updateProjectStepHandler(db))
	mux.HandleFunc("DELETE /api/projects/{id}/steps/{stepID}", deleteProjectStepHandler(db))
	mux.HandleFunc("GET /api/projects/{id}/notes", projectNotesHandler(db))
	mux.HandleFunc("POST /api/projects/{id}/notes", createProjectNoteHandler(db))
	mux.HandleFunc("GET /api/projects/{id}/activities", projectActivitiesHandler(db))
	mux.HandleFunc("GET /api/projects/{id}/reminders", projectRemindersHandler(db))
	mux.HandleFunc("GET /api/projects/{id}/calculations", projectCalculationsHandler(db))
	mux.HandleFunc("POST /api/projects/{id}/reminders", createProjectReminderHandler(db))
	mux.HandleFunc("PATCH /api/projects/{id}/reminders/{reminderID}", updateProjectReminderHandler(db))
	mux.HandleFunc("DELETE /api/projects/{id}/reminders/{reminderID}", deleteProjectReminderHandler(db))
	mux.HandleFunc("GET /api/reminders/upcoming", upcomingReminderHandler(db))
	mux.HandleFunc("POST /api/projects/{id}/reminders/{reminderID}/send", permissionHandler(db, "project.write", sendReminderEmailHandler(db, cfg)))
	mux.HandleFunc("GET /api/users", usersHandler(db))
	mux.HandleFunc("POST /api/users", createUserHandler(db))
	mux.HandleFunc("DELETE /api/users/{id}", deleteUserHandler(db))
	mux.HandleFunc("GET /api/users/{id}/permissions", userPermissionsHandler(db))
	mux.HandleFunc("PUT /api/users/{id}/permissions", saveUserPermissionsHandler(db))

	server := &http.Server{Addr: ":" + cfg.port, Handler: cors(mux), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Line API listening on http://localhost:%s", cfg.port)
	log.Fatal(server.ListenAndServe())
}

func addProjectUsageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || pid < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid project id"})
			return
		}
		var in addUsageRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.YarnID < 1 || in.Grams <= 0 {
			writeJSON(w, 400, map[string]string{"error": "valid yarn and grams are required"})
			return
		}
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not start inventory update"})
			return
		}
		defer tx.Rollback()
		var exists int
		if err := tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM projects WHERE id=? AND user_id=?`, pid, account.ID).Scan(&exists); err != nil || exists == 0 {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var remaining, price float64
		if err := tx.QueryRowContext(r.Context(), `SELECT remaining_grams,unit_price FROM yarns WHERE id=? AND user_id=? FOR UPDATE`, in.YarnID, account.ID).Scan(&remaining, &price); err != nil {
			writeJSON(w, 404, map[string]string{"error": "yarn not found"})
			return
		}
		if remaining < in.Grams {
			writeJSON(w, 409, map[string]string{"error": "not enough yarn remaining"})
			return
		}
		if _, err := tx.ExecContext(r.Context(), `UPDATE yarns SET remaining_grams=remaining_grams-? WHERE id=?`, in.Grams, in.YarnID); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not update yarn inventory"})
			return
		}
		if _, err := tx.ExecContext(r.Context(), `INSERT INTO project_yarn_usages (project_id,yarn_id,grams_used,unit_cost) VALUES (?,?,?,?)`, pid, in.YarnID, in.Grams, price); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not record yarn usage"})
			return
		}
		if err := tx.Commit(); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not save yarn usage"})
			return
		}
		writeJSON(w, 201, map[string]any{"data": map[string]float64{"remainingGrams": remaining - in.Grams, "cost": in.Grams * price / 100}})
	}
}

func projectPDFExportHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var item project
		err = db.QueryRowContext(r.Context(), `SELECT id,cover_data,completion_photo_data,name,pattern_name,pattern_file_id,target_size,hook_size,yarn_names,tool_names,palette,work_tags,share_token,share_show_materials,share_show_note,status,version,current_row,total_rows,breakpoint_note,completed_at,completion_note,actual_minutes,target_date,weekly_focus_minutes,created_at FROM projects WHERE id=? AND user_id=?`, projectID, account.ID).Scan(&item.ID, &item.CoverData, &item.CompletionPhotoData, &item.Name, &item.PatternName, &item.PatternFileID, &item.TargetSize, &item.HookSize, &item.YarnNames, &item.ToolNames, &item.Palette, &item.WorkTags, &item.ShareToken, &item.ShareShowMaterials, &item.ShareShowNote, &item.Status, &item.Version, &item.CurrentRow, &item.TotalRows, &item.BreakpointNote, &item.CompletedAt, &item.CompletionNote, &item.ActualMinutes, &item.TargetDate, &item.WeeklyFocusMinutes, &item.CreatedAt)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load project"})
			return
		}
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.SetTitle("Line Project Archive", false)
		pdf.SetAuthor("Line", false)
		pdf.AddPage()
		pdf.SetFont("Helvetica", "B", 20)
		pdf.Cell(0, 12, "LINE PROJECT ARCHIVE")
		pdf.Ln(16)
		pdf.SetFont("Helvetica", "B", 16)
		pdf.MultiCell(0, 8, item.Name, "", "L", false)
		pdf.SetFont("Helvetica", "", 11)
		pdf.Ln(3)
		pdf.MultiCell(0, 6, "Pattern: "+item.PatternName+"\nStatus: "+item.Status+"\nProgress: "+strconv.Itoa(item.CurrentRow)+" / "+strconv.Itoa(item.TotalRows)+" rows\nTarget size: "+item.TargetSize+"\nHook: "+item.HookSize+"\nFocused time: "+strconv.Itoa(item.ActualMinutes)+" minutes", "", "L", false)
		if strings.TrimSpace(item.BreakpointNote) != "" {
			pdf.Ln(4)
			pdf.SetFont("Helvetica", "B", 12)
			pdf.Cell(0, 7, "Breakpoint note")
			pdf.Ln(8)
			pdf.SetFont("Helvetica", "", 11)
			pdf.MultiCell(0, 6, item.BreakpointNote, "", "L", false)
		}
		if strings.TrimSpace(item.CompletionNote) != "" {
			pdf.Ln(4)
			pdf.SetFont("Helvetica", "B", 12)
			pdf.Cell(0, 7, "Completion note")
			pdf.Ln(8)
			pdf.SetFont("Helvetica", "", 11)
			pdf.MultiCell(0, 6, item.CompletionNote, "", "L", false)
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=\"project-"+strconv.FormatInt(projectID, 10)+".pdf\"")
		if err := pdf.Output(w); err != nil {
			log.Printf("export project pdf: %v", err)
		}
	}
}

func projectMaterialsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		usageRows, err := db.QueryContext(r.Context(), `SELECT u.id,u.yarn_id,y.name,u.grams_used,u.grams_used*u.unit_cost/100,u.created_at FROM project_yarn_usages u JOIN yarns y ON y.id=u.yarn_id WHERE u.project_id=? ORDER BY u.created_at DESC,u.id DESC`, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load yarn usages"})
			return
		}
		defer usageRows.Close()
		usages := make([]projectYarnUsage, 0)
		for usageRows.Next() {
			var item projectYarnUsage
			if err := usageRows.Scan(&item.ID, &item.YarnID, &item.YarnName, &item.Grams, &item.Cost, &item.CreatedAt); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read yarn usage"})
				return
			}
			usages = append(usages, item)
		}
		accessoryRows, err := db.QueryContext(r.Context(), `SELECT id,name,quantity,unit,unit_price,quantity*unit_price,created_at FROM project_accessories WHERE project_id=? ORDER BY created_at DESC,id DESC`, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load accessories"})
			return
		}
		defer accessoryRows.Close()
		accessories := make([]projectAccessory, 0)
		for accessoryRows.Next() {
			var item projectAccessory
			if err := accessoryRows.Scan(&item.ID, &item.Name, &item.Quantity, &item.Unit, &item.UnitPrice, &item.Cost, &item.CreatedAt); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read accessory"})
				return
			}
			accessories = append(accessories, item)
		}
		writeJSON(w, 200, map[string]any{"data": map[string]any{"yarnUsages": usages, "accessories": accessories}})
	}
}

func projectYarnPlansHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT p.yarn_id,y.name,p.planned_grams,COALESCE((SELECT SUM(u.grams_used) FROM project_yarn_usages u WHERE u.project_id=p.project_id AND u.yarn_id=p.yarn_id),0),y.remaining_grams FROM project_yarn_plans p JOIN yarns y ON y.id=p.yarn_id WHERE p.project_id=? ORDER BY y.name`, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load yarn plans"})
			return
		}
		defer rows.Close()
		plans := make([]projectYarnPlan, 0)
		for rows.Next() {
			var plan projectYarnPlan
			if err := rows.Scan(&plan.YarnID, &plan.YarnName, &plan.PlannedGrams, &plan.UsedGrams, &plan.RemainingGrams); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read yarn plan"})
				return
			}
			plans = append(plans, plan)
		}
		writeJSON(w, 200, map[string]any{"data": plans})
	}
}
func saveProjectYarnPlansHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var input projectYarnPlanRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil || len(input.Plans) > 30 {
			writeJSON(w, 400, map[string]string{"error": "invalid yarn plans"})
			return
		}
		seen := map[int64]bool{}
		for _, plan := range input.Plans {
			if plan.YarnID < 1 || plan.PlannedGrams <= 0 || plan.PlannedGrams > 100000 || seen[plan.YarnID] {
				writeJSON(w, 400, map[string]string{"error": "invalid yarn plans"})
				return
			}
			seen[plan.YarnID] = true
			var count int
			if db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM yarns WHERE id=? AND user_id=?`, plan.YarnID, account.ID).Scan(&count) != nil || count == 0 {
				writeJSON(w, 400, map[string]string{"error": "yarn not found"})
				return
			}
		}
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not save yarn plans"})
			return
		}
		defer tx.Rollback()
		if _, err = tx.ExecContext(r.Context(), `DELETE FROM project_yarn_plans WHERE project_id=?`, projectID); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not save yarn plans"})
			return
		}
		for _, plan := range input.Plans {
			if _, err = tx.ExecContext(r.Context(), `INSERT INTO project_yarn_plans (project_id,yarn_id,planned_grams) VALUES (?,?,?)`, projectID, plan.YarnID, plan.PlannedGrams); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not save yarn plans"})
				return
			}
		}
		if err = tx.Commit(); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not save yarn plans"})
			return
		}
		w.WriteHeader(204)
	}
}

func projectReadinessHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid project id"})
			return
		}
		var patternName, hookSize string
		var patternFileID *int64
		var toolNames []byte
		if err := db.QueryRowContext(r.Context(), `SELECT pattern_name,pattern_file_id,hook_size,tool_names FROM projects WHERE id=? AND user_id=?`, projectID, account.ID).Scan(&patternName, &patternFileID, &hookSize, &toolNames); err == sql.ErrNoRows {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		} else if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not check project readiness"})
			return
		}
		var tools []string
		_ = json.Unmarshal(toolNames, &tools)
		items := []projectReadinessItem{{Key: "pattern", Label: "图解", Ready: strings.TrimSpace(patternName) != "" || patternFileID != nil, Detail: "已关联图解"}, {Key: "hook", Label: "针号", Ready: strings.TrimSpace(hookSize) != "", Detail: "已填写针号"}, {Key: "tools", Label: "工具", Ready: len(tools) > 0, Detail: "已关联工具"}}
		if !items[0].Ready {
			items[0].Detail = "请关联图解或填写图解名称"
		}
		if !items[1].Ready {
			items[1].Detail = "请填写使用针号"
		}
		if !items[2].Ready {
			items[2].Detail = "请至少关联一件工具"
		}
		rows, err := db.QueryContext(r.Context(), `SELECT p.yarn_id,y.name,p.planned_grams,COALESCE((SELECT SUM(u.grams_used) FROM project_yarn_usages u WHERE u.project_id=p.project_id AND u.yarn_id=p.yarn_id),0),y.remaining_grams FROM project_yarn_plans p JOIN yarns y ON y.id=p.yarn_id WHERE p.project_id=?`, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not check material readiness"})
			return
		}
		defer rows.Close()
		planCount := 0
		for rows.Next() {
			var yarnID int64
			var name string
			var planned, used, remaining float64
			if rows.Scan(&yarnID, &name, &planned, &used, &remaining) != nil {
				writeJSON(w, 500, map[string]string{"error": "could not check material readiness"})
				return
			}
			planCount++
			missing := math.Max(0, planned-used-remaining)
			items = append(items, projectReadinessItem{Key: "yarn-" + strconv.FormatInt(yarnID, 10), Label: name, Ready: missing == 0, Detail: func() string {
				if missing > 0 {
					return fmt.Sprintf("还缺 %.0fg，可加入采购清单", missing)
				}
				return fmt.Sprintf("预计 %.0fg，库存可覆盖", planned)
			}(), YarnID: yarnID, Missing: missing})
		}
		if planCount == 0 {
			items = append(items, projectReadinessItem{Key: "material-plan", Label: "材料计划", Ready: false, Detail: "请先填写预计用量"})
		}
		ready := true
		for _, item := range items {
			ready = ready && item.Ready
		}
		writeJSON(w, 200, map[string]any{"data": map[string]any{"ready": ready, "items": items}})
	}
}

func projectForecastHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid project id"})
			return
		}
		var currentRows, totalRows, actualMinutes, weeklyTarget int
		var targetDate *time.Time
		if err := db.QueryRowContext(r.Context(), `SELECT current_row,total_rows,actual_minutes,target_date,weekly_focus_minutes FROM projects WHERE id=? AND user_id=?`, projectID, account.ID).Scan(&currentRows, &totalRows, &actualMinutes, &targetDate, &weeklyTarget); err == sql.ErrNoRows {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		} else if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not calculate forecast"})
			return
		}
		var recentMinutes int
		if err := db.QueryRowContext(r.Context(), `SELECT COALESCE(CEIL(SUM(t.seconds_elapsed)/60),0) FROM project_timers t WHERE t.project_id=? AND t.started_at>=CURDATE()-INTERVAL 6 DAY`, projectID).Scan(&recentMinutes); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not calculate forecast"})
			return
		}
		remainingRows := max(0, totalRows-currentRows)
		result := map[string]any{"remainingRows": remainingRows, "recentMinutes": recentMinutes, "weeklyTargetMinutes": weeklyTarget, "status": "needs_data", "detail": "记录一些进度和专注时长后即可预测完成时间"}
		if totalRows > 0 && currentRows > 0 && actualMinutes > 0 {
			minutesPerRow := float64(actualMinutes) / float64(currentRows)
			remainingMinutes := int(math.Ceil(float64(remainingRows) * minutesPerRow))
			result["remainingMinutes"] = remainingMinutes
			result["minutesPerRow"] = math.Round(minutesPerRow*10) / 10
			if targetDate != nil && remainingRows > 0 {
				days := max(1, int(math.Ceil(targetDate.Sub(time.Now()).Hours()/24)))
				requiredWeekly := int(math.Ceil(float64(remainingMinutes) * 7 / float64(days)))
				result["targetDate"] = targetDate
				result["requiredWeeklyMinutes"] = requiredWeekly
				forecastWeekly := recentMinutes
				if forecastWeekly == 0 {
					forecastWeekly = weeklyTarget
				}
				if forecastWeekly > 0 {
					forecastDate := time.Now().Add(time.Duration(math.Ceil(float64(remainingMinutes)*7/float64(forecastWeekly))) * 24 * time.Hour)
					result["forecastDate"] = forecastDate
					if targetDate.Before(time.Now()) || forecastDate.After(*targetDate) {
						result["status"] = "at_risk"
						result["detail"] = fmt.Sprintf("本周建议专注 %d 分钟，才能赶上目标日期", requiredWeekly)
					} else {
						result["status"] = "on_track"
						result["detail"] = fmt.Sprintf("按当前节奏预计 %s 完成", forecastDate.Format("1月2日"))
					}
				}
			} else if remainingRows == 0 {
				result["status"] = "complete"
				result["detail"] = "所有行数已完成，可以整理成品记录"
			}
		}
		writeJSON(w, 200, map[string]any{"data": result})
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func projectTimersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id,started_at,stopped_at,seconds_elapsed FROM project_timers WHERE project_id=? ORDER BY started_at DESC,id DESC LIMIT 30`, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load timers"})
			return
		}
		defer rows.Close()
		items := make([]projectTimer, 0)
		for rows.Next() {
			var item projectTimer
			if err := rows.Scan(&item.ID, &item.StartedAt, &item.StoppedAt, &item.SecondsElapsed); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read timer"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}
func startProjectTimerHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid project id"})
			return
		}
		var status string
		if err := db.QueryRowContext(r.Context(), `SELECT status FROM projects WHERE id=? AND user_id=?`, projectID, account.ID).Scan(&status); err == sql.ErrNoRows {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		} else if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not start timer"})
			return
		}
		if status != "in_progress" {
			writeJSON(w, 409, map[string]string{"error": "only in-progress projects can be timed"})
			return
		}
		var active int
		if err := db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM project_timers t JOIN projects p ON p.id=t.project_id WHERE p.user_id=? AND t.stopped_at IS NULL`, account.ID).Scan(&active); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not start timer"})
			return
		}
		if active > 0 {
			writeJSON(w, 409, map[string]string{"error": "another focus timer is already running"})
			return
		}
		result, err := db.ExecContext(r.Context(), `INSERT INTO project_timers (project_id,started_at) VALUES (?,NOW())`, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not start timer"})
			return
		}
		id, _ := result.LastInsertId()
		_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'timer_started','开始专注计时')`, projectID)
		writeJSON(w, 201, map[string]any{"data": map[string]int64{"id": id}})
	}
}
func stopProjectTimerHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		timerID, timerErr := strconv.ParseInt(r.PathValue("timerID"), 10, 64)
		if err != nil || timerErr != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not stop timer"})
			return
		}
		defer tx.Rollback()
		result, err := tx.ExecContext(r.Context(), `UPDATE project_timers SET stopped_at=NOW(),seconds_elapsed=TIMESTAMPDIFF(SECOND,started_at,NOW()) WHERE id=? AND project_id=? AND stopped_at IS NULL`, timerID, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not stop timer"})
			return
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			writeJSON(w, 409, map[string]string{"error": "timer is not running"})
			return
		}
		var seconds int
		if err = tx.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(seconds_elapsed),0) FROM project_timers WHERE project_id=?`, projectID).Scan(&seconds); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not stop timer"})
			return
		}
		if _, err = tx.ExecContext(r.Context(), `UPDATE projects SET actual_minutes=? WHERE id=?`, int(math.Ceil(float64(seconds)/60)), projectID); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not update project duration"})
			return
		}
		_, _ = tx.ExecContext(r.Context(), `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'timer_stopped',?)`, projectID, fmt.Sprintf("结束专注计时，累计 %d 分钟", int(math.Ceil(float64(seconds)/60))))
		if err = tx.Commit(); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not stop timer"})
			return
		}
		writeJSON(w, 200, map[string]any{"data": map[string]int{"actualMinutes": int(math.Ceil(float64(seconds) / 60))}})
	}
}

func monthlyInsightsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		weekly := make([]insightFocusDay, 0, 7)
		rows, err := db.QueryContext(r.Context(), `SELECT DATE(t.started_at),CEIL(SUM(t.seconds_elapsed)/60) FROM project_timers t JOIN projects p ON p.id=t.project_id WHERE p.user_id=? AND t.started_at>=CURDATE()-INTERVAL 6 DAY GROUP BY DATE(t.started_at) ORDER BY DATE(t.started_at)`, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load focus insights"})
			return
		}
		for rows.Next() {
			var day insightFocusDay
			if err := rows.Scan(&day.Date, &day.Minutes); err != nil {
				rows.Close()
				writeJSON(w, 500, map[string]string{"error": "could not read focus insights"})
				return
			}
			weekly = append(weekly, day)
		}
		rows.Close()
		var completed, minutes int
		var cost float64
		err = db.QueryRowContext(r.Context(), `SELECT COUNT(*),COALESCE(SUM(actual_minutes),0),COALESCE(SUM(COALESCE((SELECT SUM(grams_used*unit_cost/100) FROM project_yarn_usages u WHERE u.project_id=projects.id),0)+COALESCE((SELECT SUM(quantity*unit_price) FROM project_accessories a WHERE a.project_id=projects.id),0)),0) FROM projects WHERE user_id=? AND status='completed' AND completed_at>=DATE_FORMAT(CURDATE(),'%Y-%m-01')`, account.ID).Scan(&completed, &minutes, &cost)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load monthly insights"})
			return
		}
		distribution := make([]insightProject, 0)
		rows, err = db.QueryContext(r.Context(), `SELECT id,name,actual_minutes FROM projects WHERE user_id=? AND actual_minutes>0 ORDER BY actual_minutes DESC,id DESC LIMIT 4`, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load project insights"})
			return
		}
		for rows.Next() {
			var item insightProject
			if err := rows.Scan(&item.ID, &item.Name, &item.Minutes); err != nil {
				rows.Close()
				writeJSON(w, 500, map[string]string{"error": "could not read project insights"})
				return
			}
			distribution = append(distribution, item)
		}
		rows.Close()
		inactive := make([]inactiveProject, 0)
		rows, err = db.QueryContext(r.Context(), `SELECT id,name,status,DATEDIFF(CURDATE(),updated_at) FROM projects WHERE user_id=? AND status IN ('planning','in_progress') AND updated_at<CURDATE()-INTERVAL 14 DAY ORDER BY updated_at ASC LIMIT 3`, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load inactive projects"})
			return
		}
		for rows.Next() {
			var item inactiveProject
			if err := rows.Scan(&item.ID, &item.Name, &item.Status, &item.DaysIdle); err != nil {
				rows.Close()
				writeJSON(w, 500, map[string]string{"error": "could not read inactive projects"})
				return
			}
			inactive = append(inactive, item)
		}
		rows.Close()
		writeJSON(w, 200, map[string]any{"data": map[string]any{"weeklyFocus": weekly, "monthlyCompleted": completed, "monthlyMinutes": minutes, "monthlyCost": cost, "projectDistribution": distribution, "inactiveProjects": inactive}})
	}
}

func createProjectAccessoryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var input createProjectAccessoryRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil || len(strings.TrimSpace(input.Name)) < 1 || len(strings.TrimSpace(input.Name)) > 120 || input.Quantity <= 0 || input.Quantity > 100000 || input.UnitPrice < 0 || input.UnitPrice > 1000000 || len(strings.TrimSpace(input.Unit)) > 30 {
			writeJSON(w, 400, map[string]string{"error": "invalid accessory"})
			return
		}
		result, err := db.ExecContext(r.Context(), `INSERT INTO project_accessories (project_id,name,quantity,unit,unit_price) VALUES (?,?,?,?,?)`, projectID, strings.TrimSpace(input.Name), input.Quantity, strings.TrimSpace(input.Unit), input.UnitPrice)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not add accessory"})
			return
		}
		id, _ := result.LastInsertId()
		_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'accessory_added',?)`, projectID, "添加辅料："+strings.TrimSpace(input.Name))
		writeJSON(w, 201, map[string]any{"data": map[string]int64{"id": id}})
	}
}

func deleteProjectAccessoryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		accessoryID, accessoryErr := strconv.ParseInt(r.PathValue("accessoryID"), 10, 64)
		if err != nil || accessoryErr != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		result, err := db.ExecContext(r.Context(), `DELETE FROM project_accessories WHERE id=? AND project_id=?`, accessoryID, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not delete accessory"})
			return
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			writeJSON(w, 404, map[string]string{"error": "accessory not found"})
			return
		}
		w.WriteHeader(204)
	}
}

func toolsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id, COALESCE(photo_data, ''), name, type, thickness, created_at FROM tools WHERE user_id = ? ORDER BY created_at DESC, id DESC`, account.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load tools"})
			return
		}
		defer rows.Close()
		items := make([]tool, 0)
		for rows.Next() {
			var item tool
			if err := rows.Scan(&item.ID, &item.PhotoData, &item.Name, &item.Type, &item.Thickness, &item.CreatedAt); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read tool"})
				return
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load tools"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	}
}

func createToolHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		var input createToolRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if err := validateCreateTool(input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		result, err := db.ExecContext(r.Context(), `INSERT INTO tools (user_id, photo_data, name, type, thickness) VALUES (?, ?, ?, ?, ?)`, account.ID, input.PhotoData, strings.TrimSpace(input.Name), strings.TrimSpace(input.Type), strings.TrimSpace(input.Thickness))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not add tool"})
			return
		}
		id, _ := result.LastInsertId()
		recordAudit(r.Context(), db, account.ID, "tool.create", "tool", id, map[string]any{"name": strings.TrimSpace(input.Name), "type": strings.TrimSpace(input.Type)})
		writeJSON(w, http.StatusCreated, tool{ID: id, PhotoData: input.PhotoData, Name: strings.TrimSpace(input.Name), Type: strings.TrimSpace(input.Type), Thickness: strings.TrimSpace(input.Thickness)})
	}
}

func deleteToolHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tool id"})
			return
		}
		result, err := db.ExecContext(r.Context(), `DELETE FROM tools WHERE id = ? AND user_id = ?`, id, account.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not delete tool"})
			return
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "tool not found"})
			return
		}
		recordAudit(r.Context(), db, account.ID, "tool.delete", "tool", id, map[string]any{})
		w.WriteHeader(http.StatusNoContent)
	}
}

func validateCreateTool(input createToolRequest) error {
	if len(strings.TrimSpace(input.Name)) < 1 || len(strings.TrimSpace(input.Name)) > 120 {
		return fmt.Errorf("tool name must contain 1 to 120 characters")
	}
	if len(input.Type) > 80 || len(input.Thickness) > 80 {
		return fmt.Errorf("tool details are too long")
	}
	if !validImageReference(input.PhotoData) {
		return fmt.Errorf("photo must be an uploaded image")
	}
	return nil
}

func projectsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT p.id,COALESCE(p.cover_data,''),COALESCE(p.completion_photo_data,''),p.name,p.pattern_name,p.pattern_file_id,p.target_size,p.hook_size,p.yarn_names,p.tool_names,p.palette,p.work_tags,COALESCE(p.share_token,''),p.share_show_materials,p.share_show_note,p.status,p.version,p.current_row,p.total_rows,p.breakpoint_note,p.completed_at,p.completion_note,p.actual_minutes,p.target_date,p.weekly_focus_minutes,p.created_at,COALESCE((SELECT SUM(grams_used*unit_cost/100) FROM project_yarn_usages u WHERE u.project_id=p.id),0)+COALESCE((SELECT SUM(quantity*unit_price) FROM project_accessories a WHERE a.project_id=p.id),0) FROM projects p WHERE p.user_id=? ORDER BY p.updated_at DESC,p.id DESC`, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load projects"})
			return
		}
		defer rows.Close()
		items := make([]project, 0)
		for rows.Next() {
			var p project
			var ys, ts, ps, tags []byte
			if err := rows.Scan(&p.ID, &p.CoverData, &p.CompletionPhotoData, &p.Name, &p.PatternName, &p.PatternFileID, &p.TargetSize, &p.HookSize, &ys, &ts, &ps, &tags, &p.ShareToken, &p.ShareShowMaterials, &p.ShareShowNote, &p.Status, &p.Version, &p.CurrentRow, &p.TotalRows, &p.BreakpointNote, &p.CompletedAt, &p.CompletionNote, &p.ActualMinutes, &p.TargetDate, &p.WeeklyFocusMinutes, &p.CreatedAt, &p.MaterialCost); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read project"})
				return
			}
			_ = json.Unmarshal(ys, &p.YarnNames)
			_ = json.Unmarshal(ts, &p.ToolNames)
			_ = json.Unmarshal(ps, &p.Palette)
			_ = json.Unmarshal(tags, &p.WorkTags)
			items = append(items, p)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}

func projectForUser(ctx context.Context, db *sql.DB, id, userID int64) (project, error) {
	var item project
	var yarns, tools, palette, tags []byte
	err := db.QueryRowContext(ctx, `SELECT id,COALESCE(cover_data,''),COALESCE(completion_photo_data,''),name,pattern_name,pattern_file_id,target_size,hook_size,yarn_names,tool_names,palette,work_tags,COALESCE(share_token,''),share_show_materials,share_show_note,status,version,current_row,total_rows,breakpoint_note,completed_at,completion_note,actual_minutes,target_date,weekly_focus_minutes,created_at FROM projects WHERE id=? AND user_id=?`, id, userID).Scan(&item.ID, &item.CoverData, &item.CompletionPhotoData, &item.Name, &item.PatternName, &item.PatternFileID, &item.TargetSize, &item.HookSize, &yarns, &tools, &palette, &tags, &item.ShareToken, &item.ShareShowMaterials, &item.ShareShowNote, &item.Status, &item.Version, &item.CurrentRow, &item.TotalRows, &item.BreakpointNote, &item.CompletedAt, &item.CompletionNote, &item.ActualMinutes, &item.TargetDate, &item.WeeklyFocusMinutes, &item.CreatedAt)
	if err != nil {
		return project{}, err
	}
	_ = json.Unmarshal(yarns, &item.YarnNames)
	_ = json.Unmarshal(tools, &item.ToolNames)
	_ = json.Unmarshal(palette, &item.Palette)
	_ = json.Unmarshal(tags, &item.WorkTags)
	return item, nil
}
func createProjectHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		var in createProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid request body"})
			return
		}
		if err := validateProject(in.Name, in.TotalRows, in.CoverData); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if err := validateOwnedMedia(r.Context(), db, account.ID, in.CoverData); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if !validWeeklyFocusMinutes(in.WeeklyFocusMinutes) {
			writeJSON(w, 400, map[string]string{"error": "invalid weekly focus target"})
			return
		}
		ys, _ := json.Marshal(in.YarnNames)
		ts, _ := json.Marshal(in.ToolNames)
		ps, _ := json.Marshal(in.Palette)
		tags, _ := json.Marshal(in.WorkTags)
		if err := validatePatternFileOwner(r.Context(), db, account.ID, in.PatternFileID); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		result, err := db.ExecContext(r.Context(), `INSERT INTO projects (user_id,cover_data,name,pattern_name,pattern_file_id,target_size,hook_size,yarn_names,tool_names,palette,work_tags,status,total_rows,breakpoint_note,target_date,weekly_focus_minutes) VALUES (?,?,?,?,?,?,?,?,?,?,?, 'planning', ?,?,?,?)`, account.ID, in.CoverData, strings.TrimSpace(in.Name), strings.TrimSpace(in.PatternName), in.PatternFileID, strings.TrimSpace(in.TargetSize), strings.TrimSpace(in.HookSize), ys, ts, ps, tags, in.TotalRows, strings.TrimSpace(in.BreakpointNote), in.TargetDate, in.WeeklyFocusMinutes)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not create project"})
			return
		}
		id, _ := result.LastInsertId()
		_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id, kind, detail) VALUES (?, 'project_created', '创建项目')`, id)
		recordAudit(r.Context(), db, account.ID, "project.create", "project", id, map[string]any{"name": strings.TrimSpace(in.Name)})
		writeJSON(w, 201, map[string]any{"data": map[string]int64{"id": id}})
	}
}

func projectOwned(ctx context.Context, db *sql.DB, userID, projectID int64) bool {
	var count int
	return db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects WHERE id=? AND user_id=?`, projectID, userID).Scan(&count) == nil && count == 1
}

func validatePatternFileOwner(ctx context.Context, db *sql.DB, userID int64, fileID *int64) error {
	if fileID == nil {
		return nil
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pattern_files WHERE id=? AND user_id=?`, *fileID, userID).Scan(&count); err != nil || count != 1 {
		return fmt.Errorf("pattern file not found")
	}
	return nil
}

func patternFilesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id,title,file_type,mime_type,size_bytes,created_at FROM pattern_files WHERE user_id=? ORDER BY created_at DESC,id DESC`, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load pattern files"})
			return
		}
		defer rows.Close()
		items := make([]patternFile, 0)
		for rows.Next() {
			var item patternFile
			if err := rows.Scan(&item.ID, &item.Title, &item.FileType, &item.MimeType, &item.SizeBytes, &item.CreatedAt); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read pattern file"})
				return
			}
			item.URL = patternFileURL(item.ID)
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}

func createPatternFileHandler(db *sql.DB, store objectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		var in createPatternFileRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxPatternBytes*4/3+(1<<20))).Decode(&in); err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid request body"})
			return
		}
		if len(strings.TrimSpace(in.Title)) < 1 || len(strings.TrimSpace(in.Title)) > 160 || (in.FileType != "pdf" && in.FileType != "image") {
			writeJSON(w, 400, map[string]string{"error": "invalid pattern file"})
			return
		}
		mimeType, body, err := decodeDataURL(in.FileData)
		if err != nil || !patternMimeAllowed(in.FileType, mimeType) {
			writeJSON(w, 400, map[string]string{"error": "only PDF, JPEG, PNG, GIF, or WebP pattern files are supported"})
			return
		}
		if int64(len(body)) > maxPatternBytes {
			writeJSON(w, 413, map[string]string{"error": "pattern file must be smaller than 20 MB"})
			return
		}
		var used int64
		if err := db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(size_bytes),0) FROM pattern_files WHERE user_id=?`, account.ID).Scan(&used); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not check pattern storage quota"})
			return
		}
		if used+int64(len(body)) > patternQuotaBytes {
			writeJSON(w, 413, map[string]string{"error": "pattern storage quota exceeded (200 MB per account)"})
			return
		}
		result, err := db.ExecContext(r.Context(), `INSERT INTO pattern_files (user_id,title,file_type,mime_type,size_bytes) VALUES (?,?,?,?,?)`, account.ID, strings.TrimSpace(in.Title), in.FileType, mimeType, len(body))
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not save pattern file"})
			return
		}
		id, _ := result.LastInsertId()
		objectKey := patternObjectKey(id, mimeType)
		if err = store.Put(r.Context(), objectKey, body, mimeType); err != nil {
			_, _ = db.ExecContext(r.Context(), `DELETE FROM pattern_files WHERE id=?`, id)
			writeJSON(w, 500, map[string]string{"error": "could not store pattern file"})
			return
		}
		if _, err = db.ExecContext(r.Context(), `UPDATE pattern_files SET object_key=? WHERE id=? AND user_id=?`, objectKey, id, account.ID); err != nil {
			_ = store.Delete(r.Context(), objectKey)
			_, _ = db.ExecContext(r.Context(), `DELETE FROM pattern_files WHERE id=?`, id)
			writeJSON(w, 500, map[string]string{"error": "could not finalize pattern file"})
			return
		}
		recordAudit(r.Context(), db, account.ID, "pattern.create", "pattern_file", id, map[string]any{"mimeType": mimeType, "sizeBytes": len(body)})
		writeJSON(w, 201, map[string]any{"data": map[string]any{"id": id, "url": patternFileURL(id), "mimeType": mimeType, "sizeBytes": len(body)}})
	}
}

// patternFileContentHandler streams the stored object. Legacy rows that still keep the
// bytes in pattern_files.file_data are migrated to object storage on first read.
func patternFileContentHandler(db *sql.DB, store objectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			http.NotFound(w, r)
			return
		}
		body, mimeType, err := readPatternObject(r.Context(), db, store, account.ID, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Cache-Control", "private, max-age=600")
		http.ServeContent(w, r, patternObjectKey(id, mimeType), time.Now(), bytes.NewReader(body))
	}
}

func deletePatternFileHandler(db *sql.DB, store objectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid pattern file id"})
			return
		}
		var objectKey string
		if err := db.QueryRowContext(r.Context(), `SELECT COALESCE(object_key,'') FROM pattern_files WHERE id=? AND user_id=?`, id, account.ID).Scan(&objectKey); err != nil {
			writeJSON(w, 404, map[string]string{"error": "pattern file not found"})
			return
		}
		if _, err = db.ExecContext(r.Context(), `DELETE FROM pattern_files WHERE id=? AND user_id=?`, id, account.ID); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not delete pattern file"})
			return
		}
		if objectKey != "" {
			_ = store.Delete(r.Context(), objectKey)
		}
		recordAudit(r.Context(), db, account.ID, "pattern.delete", "pattern_file", id, map[string]any{})
		w.WriteHeader(204)
	}
}

func contentAskHandler(db *sql.DB, cfg config, store objectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		var input contentAskRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil || len(strings.TrimSpace(input.Question)) == 0 || len(strings.TrimSpace(input.Question)) > 4000 {
			writeJSON(w, 400, map[string]string{"error": "invalid question"})
			return
		}
		if strings.TrimSpace(cfg.openAIAPIKey) == "" {
			writeJSON(w, 503, map[string]string{"error": "AI 服务尚未配置，请设置 OPENAI_API_KEY"})
			return
		}
		content := []map[string]string{{"type": "input_text", "text": strings.TrimSpace(input.Question)}}
		if input.PatternFileID != nil {
			if !patternFileOwned(r.Context(), db, account.ID, *input.PatternFileID) {
				writeJSON(w, 404, map[string]string{"error": "pattern file not found"})
				return
			}
			var fileType string
			if err := db.QueryRowContext(r.Context(), `SELECT file_type FROM pattern_files WHERE id=? AND user_id=?`, *input.PatternFileID, account.ID).Scan(&fileType); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not load pattern file"})
				return
			}
			body, mimeType, err := readPatternObject(r.Context(), db, store, account.ID, *input.PatternFileID)
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not load pattern file"})
				return
			}
			if fileType == "image" {
				content = append(content, map[string]string{"type": "input_image", "image_url": "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(body)})
			} else {
				content = append(content, map[string]string{"type": "input_text", "text": "用户已关联一份 PDF 图解。当前版本不能直接提取 PDF 页面；请基于问题说明需要用户提供的页码、行号或截图。"})
			}
		}
		payload := map[string]any{
			"model":             cfg.openAIModel,
			"instructions":      "你是钩织图解助手。仅根据用户问题和已授权图解回答；不确定时明确说明。回答使用简体中文，给出可执行的逐步建议，不编造图解中的针数。",
			"input":             []map[string]any{{"role": "user", "content": content}},
			"max_output_tokens": 900,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not encode AI request"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cfg.openAIBaseURL, "/")+"/responses", bytes.NewReader(body))
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not create AI request"})
			return
		}
		req.Header.Set("Authorization", "Bearer "+cfg.openAIAPIKey)
		req.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			writeJSON(w, 502, map[string]string{"error": "AI 服务请求失败"})
			return
		}
		defer response.Body.Close()
		responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			writeJSON(w, 502, map[string]string{"error": "AI 服务返回错误"})
			return
		}
		var result struct {
			OutputText string `json:"output_text"`
		}
		if err := json.Unmarshal(responseBody, &result); err != nil || strings.TrimSpace(result.OutputText) == "" {
			writeJSON(w, 502, map[string]string{"error": "AI 服务未返回可用回答"})
			return
		}
		writeJSON(w, 200, map[string]any{"data": map[string]string{"answer": result.OutputText}})
	}
}

func patternFileOwned(ctx context.Context, db *sql.DB, userID, fileID int64) bool {
	var count int
	return db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pattern_files WHERE id=? AND user_id=?`, fileID, userID).Scan(&count) == nil && count == 1
}

func patternAnnotationsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		fileID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !patternFileOwned(r.Context(), db, account.ID, fileID) {
			writeJSON(w, 404, map[string]string{"error": "pattern file not found"})
			return
		}
		rows, err := db.QueryContext(r.Context(), "SELECT id,page_number,`row_number`,x,y,note,created_at FROM pattern_annotations WHERE pattern_file_id=? ORDER BY page_number,`row_number`,id", fileID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load annotations"})
			return
		}
		defer rows.Close()
		items := make([]patternAnnotation, 0)
		for rows.Next() {
			var item patternAnnotation
			if err := rows.Scan(&item.ID, &item.PageNumber, &item.RowNumber, &item.X, &item.Y, &item.Note, &item.CreatedAt); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read annotation"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}

func createPatternAnnotationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		fileID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !patternFileOwned(r.Context(), db, account.ID, fileID) {
			writeJSON(w, 404, map[string]string{"error": "pattern file not found"})
			return
		}
		var input createPatternAnnotationRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil || input.PageNumber < 1 || input.PageNumber > 10000 || input.RowNumber < 0 || input.RowNumber > 100000 || input.X < 0 || input.X > 1 || input.Y < 0 || input.Y > 1 || len(strings.TrimSpace(input.Note)) < 1 || len(strings.TrimSpace(input.Note)) > 500 {
			writeJSON(w, 400, map[string]string{"error": "invalid annotation"})
			return
		}
		result, err := db.ExecContext(r.Context(), "INSERT INTO pattern_annotations (pattern_file_id,page_number,`row_number`,x,y,note) VALUES (?,?,?,?,?,?)", fileID, input.PageNumber, input.RowNumber, input.X, input.Y, strings.TrimSpace(input.Note))
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not save annotation"})
			return
		}
		id, _ := result.LastInsertId()
		writeJSON(w, 201, map[string]any{"data": map[string]int64{"id": id}})
	}
}

func deletePatternAnnotationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		fileID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		annotationID, annotationErr := strconv.ParseInt(r.PathValue("annotationID"), 10, 64)
		if err != nil || annotationErr != nil || !patternFileOwned(r.Context(), db, account.ID, fileID) {
			writeJSON(w, 404, map[string]string{"error": "pattern file not found"})
			return
		}
		result, err := db.ExecContext(r.Context(), `DELETE FROM pattern_annotations WHERE id=? AND pattern_file_id=?`, annotationID, fileID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not delete annotation"})
			return
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			writeJSON(w, 404, map[string]string{"error": "annotation not found"})
			return
		}
		w.WriteHeader(204)
	}
}

func projectStepsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, pid) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id,title,is_done,sort_order FROM project_steps WHERE project_id=? ORDER BY sort_order,id`, pid)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load steps"})
			return
		}
		defer rows.Close()
		items := make([]projectStep, 0)
		for rows.Next() {
			var item projectStep
			if err := rows.Scan(&item.ID, &item.Title, &item.IsDone, &item.SortOrder); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read step"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}
func createProjectStepHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, pid) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var in createProjectStepRequest
		if json.NewDecoder(r.Body).Decode(&in) != nil || len(strings.TrimSpace(in.Title)) < 1 || len(strings.TrimSpace(in.Title)) > 160 {
			writeJSON(w, 400, map[string]string{"error": "a step title is required"})
			return
		}
		res, err := db.ExecContext(r.Context(), `INSERT INTO project_steps (project_id,title,sort_order) VALUES (?,?,(SELECT COALESCE(MAX(sort_order),-1)+1 FROM project_steps s WHERE s.project_id=?))`, pid, strings.TrimSpace(in.Title), pid)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not add step"})
			return
		}
		id, _ := res.LastInsertId()
		_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'step_added',?)`, pid, strings.TrimSpace(in.Title))
		writeJSON(w, 201, map[string]any{"data": map[string]int64{"id": id}})
	}
}
func updateProjectStepHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		sid, serr := strconv.ParseInt(r.PathValue("stepID"), 10, 64)
		if err != nil || serr != nil || !projectOwned(r.Context(), db, account.ID, pid) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var in updateProjectStepRequest
		if json.NewDecoder(r.Body).Decode(&in) != nil || len(strings.TrimSpace(in.Title)) < 1 || len(strings.TrimSpace(in.Title)) > 160 {
			writeJSON(w, 400, map[string]string{"error": "invalid step"})
			return
		}
		res, err := db.ExecContext(r.Context(), `UPDATE project_steps SET title=?,is_done=?,sort_order=? WHERE id=? AND project_id=?`, strings.TrimSpace(in.Title), in.IsDone, in.SortOrder, sid, pid)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not update step"})
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			writeJSON(w, 404, map[string]string{"error": "step not found"})
			return
		}
		_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'step_updated',?)`, pid, strings.TrimSpace(in.Title))
		w.WriteHeader(204)
	}
}
func deleteProjectStepHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		sid, serr := strconv.ParseInt(r.PathValue("stepID"), 10, 64)
		if err != nil || serr != nil || !projectOwned(r.Context(), db, account.ID, pid) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		db.ExecContext(r.Context(), `DELETE FROM project_steps WHERE id=? AND project_id=?`, sid, pid)
		w.WriteHeader(204)
	}
}
func projectNotesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, pid) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id,body,COALESCE(photo_data,''),created_at FROM project_notes WHERE project_id=? ORDER BY created_at DESC,id DESC`, pid)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load notes"})
			return
		}
		defer rows.Close()
		items := make([]projectNote, 0)
		for rows.Next() {
			var item projectNote
			if err := rows.Scan(&item.ID, &item.Body, &item.PhotoData, &item.CreatedAt); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read note"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}
func createProjectNoteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, pid) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var in createProjectNoteRequest
		if json.NewDecoder(r.Body).Decode(&in) != nil || len(strings.TrimSpace(in.Body)) < 1 || len(strings.TrimSpace(in.Body)) > 10000 || !validImageReference(in.PhotoData) {
			writeJSON(w, 400, map[string]string{"error": "invalid note"})
			return
		}
		_, err = db.ExecContext(r.Context(), `INSERT INTO project_notes (project_id,body,photo_data) VALUES (?,?,?)`, pid, strings.TrimSpace(in.Body), in.PhotoData)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not save note"})
			return
		}
		_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'note_added','添加断点笔记')`, pid)
		w.WriteHeader(201)
	}
}
func projectActivitiesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, pid) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id,kind,detail,created_at FROM project_activities WHERE project_id=? ORDER BY created_at DESC,id DESC LIMIT 20`, pid)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load activities"})
			return
		}
		defer rows.Close()
		items := make([]projectActivity, 0)
		for rows.Next() {
			var item projectActivity
			if err := rows.Scan(&item.ID, &item.Kind, &item.Detail, &item.CreatedAt); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read activity"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}

func projectRemindersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id,target_row,message,is_done,remind_at,channel,email_sent_at,attempts,COALESCE(last_error,'') FROM project_reminders WHERE project_id=? ORDER BY is_done,target_row,id`, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load reminders"})
			return
		}
		defer rows.Close()
		items := make([]projectReminder, 0)
		for rows.Next() {
			var item projectReminder
			if err := rows.Scan(&item.ID, &item.TargetRow, &item.Message, &item.IsDone, &item.RemindAt, &item.Channel, &item.EmailSentAt, &item.Attempts, &item.LastError); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read reminder"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}

func createProjectReminderHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var input createProjectReminderRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil || input.TargetRow < 1 || input.TargetRow > 100000 || len(strings.TrimSpace(input.Message)) < 1 || len(strings.TrimSpace(input.Message)) > 255 {
			writeJSON(w, 400, map[string]string{"error": "a valid target row and reminder are required"})
			return
		}
		channel, remindAt, err := validateReminderDelivery(input.Channel, input.RemindAt)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		result, err := db.ExecContext(r.Context(), `INSERT INTO project_reminders (project_id,target_row,message,remind_at,channel) VALUES (?,?,?,?,?)`, projectID, input.TargetRow, strings.TrimSpace(input.Message), remindAt, channel)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not create reminder"})
			return
		}
		id, _ := result.LastInsertId()
		_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'reminder_added',?)`, projectID, fmt.Sprintf("第 %d 行提醒：%s", input.TargetRow, strings.TrimSpace(input.Message)))
		writeJSON(w, 201, map[string]any{"data": map[string]int64{"id": id}})
	}
}

func updateProjectReminderHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		reminderID, reminderErr := strconv.ParseInt(r.PathValue("reminderID"), 10, 64)
		if err != nil || reminderErr != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var input updateProjectReminderRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil || input.TargetRow < 1 || input.TargetRow > 100000 || len(strings.TrimSpace(input.Message)) < 1 || len(strings.TrimSpace(input.Message)) > 255 {
			writeJSON(w, 400, map[string]string{"error": "invalid reminder"})
			return
		}
		channel, remindAt, err := validateReminderDelivery(input.Channel, input.RemindAt)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		// 修改提醒会重新武装投递状态，避免旧的发送记录掩盖新的时间。
		result, err := db.ExecContext(r.Context(), `UPDATE project_reminders SET target_row=?,message=?,is_done=?,channel=?,remind_at=?,email_sent_at=NULL,attempts=0,last_error=NULL WHERE id=? AND project_id=?`, input.TargetRow, strings.TrimSpace(input.Message), input.IsDone, channel, remindAt, reminderID, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not update reminder"})
			return
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			writeJSON(w, 404, map[string]string{"error": "reminder not found"})
			return
		}
		if input.IsDone {
			_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'reminder_done',?)`, projectID, fmt.Sprintf("完成第 %d 行提醒", input.TargetRow))
		}
		w.WriteHeader(204)
	}
}

func deleteProjectReminderHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		reminderID, reminderErr := strconv.ParseInt(r.PathValue("reminderID"), 10, 64)
		if err != nil || reminderErr != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		_, err = db.ExecContext(r.Context(), `DELETE FROM project_reminders WHERE id=? AND project_id=?`, reminderID, projectID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not delete reminder"})
			return
		}
		w.WriteHeader(204)
	}
}

func upcomingReminderHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		var item upcomingReminder
		err := db.QueryRowContext(r.Context(), `SELECT p.id,p.name,r.target_row,r.message,r.remind_at,r.channel FROM project_reminders r JOIN projects p ON p.id=r.project_id WHERE p.user_id=? AND p.status IN ('planning','in_progress') AND r.is_done=FALSE ORDER BY CASE WHEN r.target_row>=p.current_row THEN 0 ELSE 1 END,r.target_row,p.updated_at DESC LIMIT 1`, account.ID).Scan(&item.ProjectID, &item.ProjectName, &item.TargetRow, &item.Message, &item.RemindAt, &item.Channel)
		if err == sql.ErrNoRows {
			writeJSON(w, 200, map[string]any{"data": nil})
			return
		}
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load upcoming reminder"})
			return
		}
		writeJSON(w, 200, map[string]any{"data": item})
	}
}

const (
	reminderChannelInApp       = "in_app"
	reminderChannelEmail       = "email"
	reminderMaxAttempts        = 5
	reminderDispatchInterval   = time.Minute
	reminderDispatchBatch      = 50
	reminderMaxFutureWindow    = 2 * 365 * 24 * time.Hour
	reminderMaxPastWindow      = 365 * 24 * time.Hour
	reminderErrorMaxCharacters = 240
)

// validateReminderDelivery 归一化提醒渠道与发送时间：未指定时间时视为仅站内提醒。
func validateReminderDelivery(channel string, remindAt *time.Time) (string, *time.Time, error) {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		channel = reminderChannelInApp
	}
	if channel != reminderChannelInApp && channel != reminderChannelEmail {
		return "", nil, fmt.Errorf("channel must be in_app or email")
	}
	if remindAt == nil {
		if channel == reminderChannelEmail {
			return "", nil, fmt.Errorf("email reminders need a remindAt time")
		}
		return channel, nil, nil
	}
	now := time.Now()
	if remindAt.Before(now.Add(-reminderMaxPastWindow)) || remindAt.After(now.Add(reminderMaxFutureWindow)) {
		return "", nil, fmt.Errorf("remindAt must be within one year ago and two years ahead")
	}
	utc := remindAt.UTC()
	return channel, &utc, nil
}

// reminderEmailMessage 生成纯文本提醒邮件，便于单测固定文案与关键信息。
func reminderEmailMessage(recipientName, projectName string, targetRow, currentRow, totalRows int, note, appBaseURL string, projectID int64) (string, string) {
	name := strings.TrimSpace(recipientName)
	if name == "" {
		name = "钩织手"
	}
	subject := fmt.Sprintf("【线禾】%s · 第 %d 行提醒", projectName, targetRow)
	progress := fmt.Sprintf("当前进度：第 %d 行", currentRow)
	if totalRows > 0 {
		progress = fmt.Sprintf("当前进度：第 %d / %d 行", currentRow, totalRows)
	}
	lines := []string{
		name + "，你好：",
		"",
		fmt.Sprintf("《%s》在第 %d 行设了提醒：%s", projectName, targetRow, note),
		progress,
		"",
		fmt.Sprintf("继续钩织：%s/?project=%d", strings.TrimRight(appBaseURL, "/"), projectID),
		"",
		"如果已经完成，可以在项目页把这条提醒标记为已完成，之后不会再收到邮件。",
		"— 线禾",
	}
	return subject, strings.Join(lines, "\r\n")
}

type dueReminder struct {
	ID          int64
	ProjectID   int64
	UserID      int64
	ProjectName string
	TargetRow   int
	CurrentRow  int
	TotalRows   int
	Message     string
	Attempts    int
	Email       string
	Name        string
}

// dispatchDueReminders 发送到期的邮件提醒，并记录投递结果。
func dispatchDueReminders(ctx context.Context, db *sql.DB, cfg config, send func(context.Context, string, string, string) error) (int, error) {
	rows, err := db.QueryContext(ctx, `SELECT r.id,r.project_id,u.id,p.name,r.target_row,p.current_row,p.total_rows,r.message,r.attempts,u.email,u.name
FROM project_reminders r
JOIN projects p ON p.id=r.project_id
JOIN users u ON u.id=p.user_id
WHERE r.channel=? AND r.email_sent_at IS NULL AND r.is_done=FALSE AND r.remind_at IS NOT NULL AND r.remind_at<=NOW()
  AND r.attempts<? AND p.status IN ('planning','in_progress') AND u.email_verified_at IS NOT NULL
ORDER BY r.remind_at,r.id LIMIT ?`, reminderChannelEmail, reminderMaxAttempts, reminderDispatchBatch)
	if err != nil {
		return 0, err
	}
	items := make([]dueReminder, 0)
	for rows.Next() {
		var item dueReminder
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.UserID, &item.ProjectName, &item.TargetRow, &item.CurrentRow, &item.TotalRows, &item.Message, &item.Attempts, &item.Email, &item.Name); err != nil {
			rows.Close()
			return 0, err
		}
		items = append(items, item)
	}
	rows.Close()

	sent := 0
	for _, item := range items {
		subject, body := reminderEmailMessage(item.Name, item.ProjectName, item.TargetRow, item.CurrentRow, item.TotalRows, item.Message, cfg.appBaseURL, item.ProjectID)
		if err := send(ctx, item.Email, subject, body); err != nil {
			reason := err.Error()
			if len(reason) > reminderErrorMaxCharacters {
				reason = reason[:reminderErrorMaxCharacters]
			}
			_, _ = db.ExecContext(ctx, `UPDATE project_reminders SET attempts=attempts+1,last_error=? WHERE id=?`, reason, item.ID)
			recordAudit(ctx, db, item.UserID, "reminder.email_failed", "project_reminder", item.ID, map[string]any{"projectId": item.ProjectID, "reason": reason})
			continue
		}
		_, _ = db.ExecContext(ctx, `UPDATE project_reminders SET attempts=attempts+1,email_sent_at=NOW(),last_error=NULL WHERE id=?`, item.ID)
		_, _ = db.ExecContext(ctx, `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'reminder_emailed',?)`, item.ProjectID, fmt.Sprintf("第 %d 行提醒已发送邮件", item.TargetRow))
		recordAudit(ctx, db, item.UserID, "reminder.email_sent", "project_reminder", item.ID, map[string]any{"projectId": item.ProjectID})
		sent++
	}
	return sent, nil
}

// startReminderDispatcher 每分钟检查一次到期提醒；未配置 SMTP 时不启动，避免把提醒误标为已发送。
func startReminderDispatcher(ctx context.Context, db *sql.DB, cfg config) {
	if strings.TrimSpace(cfg.smtpHost) == "" {
		log.Printf("SMTP 未配置：邮件提醒调度器未启动")
		return
	}
	send := func(ctx context.Context, to, subject, body string) error {
		return sendAuthEmail(cfg, to, subject, body)
	}
	dispatch := func() {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		sent, err := dispatchDueReminders(ctx, db, cfg, send)
		if err != nil {
			log.Printf("reminder dispatch: %v", err)
			return
		}
		if sent > 0 {
			log.Printf("reminder dispatch: sent %d email(s)", sent)
		}
	}
	go func() {
		dispatch()
		ticker := time.NewTicker(reminderDispatchInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dispatch()
			}
		}
	}()
}

// sendReminderEmailHandler 立即发送单条提醒，用于用户手动确认邮件通道是否可用。
func sendReminderEmailHandler(db *sql.DB, cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		if strings.TrimSpace(cfg.smtpHost) == "" {
			writeJSON(w, 503, map[string]string{"error": "邮件服务尚未配置，请设置 SMTP_HOST"})
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		reminderID, reminderErr := strconv.ParseInt(r.PathValue("reminderID"), 10, 64)
		if err != nil || reminderErr != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var item dueReminder
		err = db.QueryRowContext(r.Context(), `SELECT r.id,r.project_id,u.id,p.name,r.target_row,p.current_row,p.total_rows,r.message,r.attempts,u.email,u.name
FROM project_reminders r JOIN projects p ON p.id=r.project_id JOIN users u ON u.id=p.user_id
WHERE r.id=? AND r.project_id=? AND p.user_id=?`, reminderID, projectID, account.ID).
			Scan(&item.ID, &item.ProjectID, &item.UserID, &item.ProjectName, &item.TargetRow, &item.CurrentRow, &item.TotalRows, &item.Message, &item.Attempts, &item.Email, &item.Name)
		if err == sql.ErrNoRows {
			writeJSON(w, 404, map[string]string{"error": "reminder not found"})
			return
		}
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load reminder"})
			return
		}
		subject, body := reminderEmailMessage(item.Name, item.ProjectName, item.TargetRow, item.CurrentRow, item.TotalRows, item.Message, cfg.appBaseURL, item.ProjectID)
		if err := sendAuthEmail(cfg, item.Email, subject, body); err != nil {
			reason := err.Error()
			if len(reason) > reminderErrorMaxCharacters {
				reason = reason[:reminderErrorMaxCharacters]
			}
			_, _ = db.ExecContext(r.Context(), `UPDATE project_reminders SET attempts=attempts+1,last_error=? WHERE id=?`, reason, item.ID)
			writeJSON(w, 502, map[string]string{"error": "邮件发送失败：" + reason})
			return
		}
		if _, err := db.ExecContext(r.Context(), `UPDATE project_reminders SET attempts=attempts+1,email_sent_at=NOW(),last_error=NULL,channel=? WHERE id=?`, reminderChannelEmail, item.ID); err != nil {
			writeJSON(w, 500, map[string]string{"error": "邮件已发送，但提醒状态未能保存"})
			return
		}
		_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'reminder_emailed',?)`, item.ProjectID, fmt.Sprintf("第 %d 行提醒已发送邮件", item.TargetRow))
		recordAudit(r.Context(), db, account.ID, "reminder.email_sent", "project_reminder", item.ID, map[string]any{"projectId": item.ProjectID, "manual": true})
		writeJSON(w, 200, map[string]any{"data": map[string]string{"email": item.Email}})
	}
}

func updateProjectHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid project id"})
			return
		}
		var in updateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid request body"})
			return
		}
		if in.Status != "planning" && in.Status != "in_progress" && in.Status != "completed" && in.Status != "archived" {
			writeJSON(w, 400, map[string]string{"error": "invalid project status"})
			return
		}
		if in.CurrentRow < 0 || in.TotalRows < 0 || in.CurrentRow > in.TotalRows && in.TotalRows > 0 {
			writeJSON(w, 400, map[string]string{"error": "invalid project progress"})
			return
		}
		if err := validateProject(in.Name, in.TotalRows, in.CoverData); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if err := validateOwnedMedia(r.Context(), db, account.ID, in.CoverData); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if err := validatePatternFileOwner(r.Context(), db, account.ID, in.PatternFileID); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if in.ActualMinutes < 0 || in.ActualMinutes > 1000000 || len(strings.TrimSpace(in.CompletionNote)) > 10000 || !validImageReference(in.CompletionPhotoData) {
			writeJSON(w, 400, map[string]string{"error": "invalid completion record"})
			return
		}
		if err := validateOwnedMedia(r.Context(), db, account.ID, in.CompletionPhotoData); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if !validWeeklyFocusMinutes(in.WeeklyFocusMinutes) {
			writeJSON(w, 400, map[string]string{"error": "invalid weekly focus target"})
			return
		}
		var previousStatus string
		var currentVersion int
		if err := db.QueryRowContext(r.Context(), `SELECT status,version FROM projects WHERE id=? AND user_id=?`, id, account.ID).Scan(&previousStatus, &currentVersion); err == sql.ErrNoRows {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		} else if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load project"})
			return
		}
		if in.Version > 0 && in.Version != currentVersion {
			current, _ := projectForUser(r.Context(), db, id, account.ID)
			writeJSON(w, http.StatusConflict, map[string]any{"error": "project was changed elsewhere", "code": "conflict", "currentVersion": currentVersion, "currentProject": current})
			return
		}
		ys, _ := json.Marshal(in.YarnNames)
		ts, _ := json.Marshal(in.ToolNames)
		ps, _ := json.Marshal(in.Palette)
		tags, _ := json.Marshal(in.WorkTags)
		updateSQL := `UPDATE projects SET cover_data=?,completion_photo_data=?,name=?,pattern_name=?,pattern_file_id=?,target_size=?,hook_size=?,yarn_names=?,tool_names=?,palette=?,work_tags=?,status=?,version=version+1,current_row=?,total_rows=?,breakpoint_note=?,completed_at=CASE WHEN ?='completed' AND ?<>'completed' THEN NOW() WHEN ?<>'completed' THEN NULL ELSE completed_at END,completion_note=?,actual_minutes=?,target_date=?,weekly_focus_minutes=? WHERE id=? AND user_id=?`
		args := []any{in.CoverData, in.CompletionPhotoData, strings.TrimSpace(in.Name), strings.TrimSpace(in.PatternName), in.PatternFileID, strings.TrimSpace(in.TargetSize), strings.TrimSpace(in.HookSize), ys, ts, ps, tags, in.Status, in.CurrentRow, in.TotalRows, strings.TrimSpace(in.BreakpointNote), in.Status, previousStatus, in.Status, strings.TrimSpace(in.CompletionNote), in.ActualMinutes, in.TargetDate, in.WeeklyFocusMinutes, id, account.ID}
		if in.Version > 0 {
			updateSQL += ` AND version=?`
			args = append(args, in.Version)
		}
		res, err := db.ExecContext(r.Context(), updateSQL, args...)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not update project"})
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			current, _ := projectForUser(r.Context(), db, id, account.ID)
			writeJSON(w, http.StatusConflict, map[string]any{"error": "project was changed elsewhere", "code": "conflict", "currentVersion": currentVersion, "currentProject": current})
			return
		}
		activityKind, activityDetail := "project_updated", "更新项目资料与进度"
		if previousStatus != "completed" && in.Status == "completed" {
			activityKind, activityDetail = "project_completed", "完成项目并生成作品档案"
		}
		_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id, kind, detail) VALUES (?, ?, ?)`, id, activityKind, activityDetail)
		recordAudit(r.Context(), db, account.ID, "project.update", "project", id, map[string]any{"version": currentVersion + 1, "status": in.Status, "currentRow": in.CurrentRow})
		var shareToken string
		if db.QueryRowContext(r.Context(), `SELECT COALESCE(share_token,'') FROM projects WHERE id=? AND user_id=?`, id, account.ID).Scan(&shareToken) == nil && shareToken != "" {
			setMediaPublic(r.Context(), db, in.CoverData, true)
			setMediaPublic(r.Context(), db, in.CompletionPhotoData, true)
		}
		w.WriteHeader(204)
	}
}
func deleteProjectHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid project id"})
			return
		}
		var cover, completion string
		_ = db.QueryRowContext(r.Context(), `SELECT COALESCE(cover_data,''),COALESCE(completion_photo_data,'') FROM projects WHERE id=? AND user_id=?`, id, account.ID).Scan(&cover, &completion)
		res, err := db.ExecContext(r.Context(), `DELETE FROM projects WHERE id=? AND user_id=?`, id, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not delete project"})
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		releaseMedia(r.Context(), db, cover)
		releaseMedia(r.Context(), db, completion)
		recordAudit(r.Context(), db, account.ID, "project.delete", "project", id, map[string]any{"name": "deleted"})
		w.WriteHeader(204)
	}
}

func shareProjectHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, id) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		var in shareProjectRequest
		_ = json.NewDecoder(r.Body).Decode(&in)
		raw := make([]byte, 16)
		if _, err := rand.Read(raw); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not create share link"})
			return
		}
		token := hex.EncodeToString(raw)
		tags, _ := json.Marshal(in.WorkTags)
		_, err = db.ExecContext(r.Context(), `UPDATE projects SET share_token=?,work_tags=?,share_show_materials=?,share_show_note=? WHERE id=? AND user_id=?`, token, tags, in.ShowMaterials, in.ShowNote, id, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not share project"})
			return
		}
		var cover, completion string
		if db.QueryRowContext(r.Context(), `SELECT COALESCE(cover_data,''),COALESCE(completion_photo_data,'') FROM projects WHERE id=? AND user_id=?`, id, account.ID).Scan(&cover, &completion) == nil {
			setMediaPublic(r.Context(), db, cover, true)
			setMediaPublic(r.Context(), db, completion, true)
		}
		recordAudit(r.Context(), db, account.ID, "project.share", "project", id, map[string]any{"showMaterials": in.ShowMaterials, "showNote": in.ShowNote})
		writeJSON(w, 200, map[string]any{"data": map[string]string{"token": token}})
	}
}
func unshareProjectHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid project id"})
			return
		}
		_, err = db.ExecContext(r.Context(), `UPDATE projects SET share_token=NULL WHERE id=? AND user_id=?`, id, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not close share link"})
			return
		}
		var cover, completion string
		if db.QueryRowContext(r.Context(), `SELECT COALESCE(cover_data,''),COALESCE(completion_photo_data,'') FROM projects WHERE id=? AND user_id=?`, id, account.ID).Scan(&cover, &completion) == nil {
			setMediaPublic(r.Context(), db, cover, false)
			setMediaPublic(r.Context(), db, completion, false)
		}
		recordAudit(r.Context(), db, account.ID, "project.unshare", "project", id, map[string]any{})
		w.WriteHeader(204)
	}
}
func publicShareHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.PathValue("token")
		var name, photo, note string
		var minutes int
		var tags, yarns, tools []byte
		var showMaterials, showNote bool
		err := db.QueryRowContext(r.Context(), `SELECT name,COALESCE(completion_photo_data,cover_data,''),completion_note,actual_minutes,work_tags,yarn_names,tool_names,share_show_materials,share_show_note FROM projects WHERE share_token=? AND status='completed'`, token).Scan(&name, &photo, &note, &minutes, &tags, &yarns, &tools, &showMaterials, &showNote)
		if err == sql.ErrNoRows {
			writeJSON(w, 404, map[string]string{"error": "shared work not found"})
			return
		}
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load shared work"})
			return
		}
		var workTags, yarnNames, toolNames []string
		_ = json.Unmarshal(tags, &workTags)
		if showMaterials {
			_ = json.Unmarshal(yarns, &yarnNames)
			_ = json.Unmarshal(tools, &toolNames)
		}
		if !showNote {
			note = ""
		}
		writeJSON(w, 200, map[string]any{"data": map[string]any{"name": name, "photoData": photo, "completionNote": note, "actualMinutes": minutes, "workTags": workTags, "yarnNames": yarnNames, "toolNames": toolNames, "showMaterials": showMaterials, "showNote": showNote}})
	}
}
func validateProject(name string, total int, cover string) error {
	if len(strings.TrimSpace(name)) < 1 || len(strings.TrimSpace(name)) > 120 {
		return fmt.Errorf("project name must contain 1 to 120 characters")
	}
	if total < 0 || total > 100000 {
		return fmt.Errorf("total rows must be between 0 and 100000")
	}
	if !validImageReference(cover) {
		return fmt.Errorf("cover image must be an uploaded image")
	}
	return nil
}

func validWeeklyFocusMinutes(minutes int) bool { return minutes >= 0 && minutes <= 10080 }

func validImageReference(value string) bool {
	if value == "" {
		return true
	}
	return (strings.HasPrefix(value, "/api/media/") && len(value) < 128) || (strings.HasPrefix(value, "data:image/") && len(value) <= 2_800_000)
}

func mediaIDFromReference(value string) (int64, bool) {
	if !strings.HasPrefix(value, "/api/media/") {
		return 0, false
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(value, "/api/media/"), 10, 64)
	return id, err == nil && id > 0
}

func setMediaPublic(ctx context.Context, db *sql.DB, value string, public bool) {
	if id, ok := mediaIDFromReference(value); ok {
		_, _ = db.ExecContext(ctx, `UPDATE media_assets SET is_public=? WHERE id=?`, public, id)
	}
}

func validateOwnedMedia(ctx context.Context, db *sql.DB, owner int64, value string) error {
	id, ok := mediaIDFromReference(value)
	if !ok {
		return nil
	}
	var found int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM media_assets WHERE id=? AND user_id=?`, id, owner).Scan(&found); err != nil || found == 0 {
		return fmt.Errorf("image does not belong to this account")
	}
	_, _ = db.ExecContext(ctx, `UPDATE media_assets SET deleted_at=NULL WHERE id=? AND user_id=?`, id, owner)
	return nil
}

func releaseMedia(ctx context.Context, db *sql.DB, value string) {
	if id, ok := mediaIDFromReference(value); ok {
		_, _ = db.ExecContext(ctx, `UPDATE media_assets SET deleted_at=COALESCE(deleted_at,NOW()),is_public=FALSE WHERE id=?`, id)
	}
}

func yarnsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id, COALESCE(photo_data, ''), name, brand, batch_code, color, color_hex, material, thickness, plies, remaining_grams, unit_price, created_at FROM yarns WHERE user_id = ? ORDER BY created_at DESC, id DESC`, account.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load yarns"})
			return
		}
		defer rows.Close()
		items := make([]yarn, 0)
		for rows.Next() {
			var item yarn
			if err := rows.Scan(&item.ID, &item.PhotoData, &item.Name, &item.Brand, &item.BatchCode, &item.Color, &item.ColorHex, &item.Material, &item.Thickness, &item.Plies, &item.RemainingGrams, &item.UnitPrice, &item.CreatedAt); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read yarn"})
				return
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load yarns"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	}
}

func createYarnHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		var input createYarnRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if err := validateCreateYarn(input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := validateOwnedMedia(r.Context(), db, account.ID, input.PhotoData); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		result, err := db.ExecContext(r.Context(), `INSERT INTO yarns (user_id, photo_data, name, brand, batch_code, color, color_hex, material, thickness, plies, remaining_grams, unit_price) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, account.ID, input.PhotoData, strings.TrimSpace(input.Name), strings.TrimSpace(input.Brand), strings.TrimSpace(input.BatchCode), strings.TrimSpace(input.Color), strings.TrimSpace(input.ColorHex), strings.TrimSpace(input.Material), strings.TrimSpace(input.Thickness), input.Plies, input.RemainingGrams, input.UnitPrice)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not add yarn"})
			return
		}
		id, _ := result.LastInsertId()
		recordAudit(r.Context(), db, account.ID, "yarn.create", "yarn", id, map[string]any{"name": strings.TrimSpace(input.Name), "remainingGrams": input.RemainingGrams})
		writeJSON(w, http.StatusCreated, yarn{ID: id, PhotoData: input.PhotoData, Name: strings.TrimSpace(input.Name), Brand: strings.TrimSpace(input.Brand), BatchCode: strings.TrimSpace(input.BatchCode), Color: strings.TrimSpace(input.Color), ColorHex: strings.TrimSpace(input.ColorHex), Material: strings.TrimSpace(input.Material), Thickness: strings.TrimSpace(input.Thickness), Plies: input.Plies, RemainingGrams: input.RemainingGrams, UnitPrice: input.UnitPrice})
	}
}

func deleteYarnHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid yarn id"})
			return
		}
		var photo string
		_ = db.QueryRowContext(r.Context(), `SELECT COALESCE(photo_data,'') FROM yarns WHERE id=? AND user_id=?`, id, account.ID).Scan(&photo)
		result, err := db.ExecContext(r.Context(), `DELETE FROM yarns WHERE id = ? AND user_id = ?`, id, account.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not delete yarn"})
			return
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "yarn not found"})
			return
		}
		releaseMedia(r.Context(), db, photo)
		recordAudit(r.Context(), db, account.ID, "yarn.delete", "yarn", id, map[string]any{})
		w.WriteHeader(http.StatusNoContent)
	}
}

func validateShoppingItem(input shoppingItemRequest) error {
	if len(strings.TrimSpace(input.YarnName)) < 1 || len(strings.TrimSpace(input.YarnName)) > 120 || input.TargetGrams <= 0 || input.TargetGrams > 100000 || input.EstimatedPrice < 0 || input.EstimatedPrice > 1000000 {
		return fmt.Errorf("invalid shopping item")
	}
	if input.Priority != "low" && input.Priority != "normal" && input.Priority != "high" {
		return fmt.Errorf("invalid shopping priority")
	}
	if input.Status != "pending" && input.Status != "purchased" {
		return fmt.Errorf("invalid shopping status")
	}
	return nil
}

func shoppingListHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id,yarn_id,yarn_name,target_grams,estimated_price,priority,status,created_at FROM shopping_items WHERE user_id=? ORDER BY status='pending' DESC,CASE priority WHEN 'high' THEN 0 WHEN 'normal' THEN 1 ELSE 2 END,created_at DESC,id DESC`, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load shopping list"})
			return
		}
		defer rows.Close()
		items := make([]shoppingItem, 0)
		for rows.Next() {
			var item shoppingItem
			if err := rows.Scan(&item.ID, &item.YarnID, &item.YarnName, &item.TargetGrams, &item.EstimatedPrice, &item.Priority, &item.Status, &item.CreatedAt); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read shopping item"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}
func createShoppingItemHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		var in shoppingItemRequest
		if json.NewDecoder(r.Body).Decode(&in) != nil || validateShoppingItem(in) != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid shopping item"})
			return
		}
		if in.YarnID != nil {
			var count int
			if db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM yarns WHERE id=? AND user_id=?`, *in.YarnID, account.ID).Scan(&count) != nil || count == 0 {
				writeJSON(w, 400, map[string]string{"error": "yarn not found"})
				return
			}
		}
		result, err := db.ExecContext(r.Context(), `INSERT INTO shopping_items (user_id,yarn_id,yarn_name,target_grams,estimated_price,priority,status) VALUES (?,?,?,?,?,?,?)`, account.ID, in.YarnID, strings.TrimSpace(in.YarnName), in.TargetGrams, in.EstimatedPrice, in.Priority, in.Status)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not add shopping item"})
			return
		}
		id, _ := result.LastInsertId()
		recordAudit(r.Context(), db, account.ID, "shopping_item.create", "shopping_item", id, map[string]any{"name": strings.TrimSpace(in.YarnName), "status": in.Status})
		writeJSON(w, 201, map[string]any{"data": map[string]int64{"id": id}})
	}
}
func updateShoppingItemHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		var in shoppingItemRequest
		if err != nil || json.NewDecoder(r.Body).Decode(&in) != nil || validateShoppingItem(in) != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid shopping item"})
			return
		}
		result, err := db.ExecContext(r.Context(), `UPDATE shopping_items SET yarn_id=?,yarn_name=?,target_grams=?,estimated_price=?,priority=?,status=? WHERE id=? AND user_id=?`, in.YarnID, strings.TrimSpace(in.YarnName), in.TargetGrams, in.EstimatedPrice, in.Priority, in.Status, id, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not update shopping item"})
			return
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			writeJSON(w, 404, map[string]string{"error": "shopping item not found"})
			return
		}
		recordAudit(r.Context(), db, account.ID, "shopping_item.update", "shopping_item", id, map[string]any{"name": strings.TrimSpace(in.YarnName), "status": in.Status})
		w.WriteHeader(204)
	}
}
func deleteShoppingItemHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid shopping item id"})
			return
		}
		result, err := db.ExecContext(r.Context(), `DELETE FROM shopping_items WHERE id=? AND user_id=?`, id, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not delete shopping item"})
			return
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			writeJSON(w, 404, map[string]string{"error": "shopping item not found"})
			return
		}
		recordAudit(r.Context(), db, account.ID, "shopping_item.delete", "shopping_item", id, map[string]any{})
		w.WriteHeader(204)
	}
}
func inventoryShortagesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		yarnRows, err := db.QueryContext(r.Context(), `SELECT id,name,remaining_grams FROM yarns WHERE user_id=? ORDER BY remaining_grams ASC`, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load inventory shortages"})
			return
		}
		defer yarnRows.Close()
		projectRows, err := db.QueryContext(r.Context(), `SELECT yp.yarn_id,p.name,yp.planned_grams,COALESCE((SELECT SUM(u.grams_used) FROM project_yarn_usages u WHERE u.project_id=yp.project_id AND u.yarn_id=yp.yarn_id),0) FROM project_yarn_plans yp JOIN projects p ON p.id=yp.project_id WHERE p.user_id=? AND p.status IN ('planning','in_progress')`, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load project requirements"})
			return
		}
		defer projectRows.Close()
		projectYarns := map[int64][]string{}
		plannedNeeds := map[int64]float64{}
		for projectRows.Next() {
			var yarnID int64
			var name string
			var planned, used float64
			if projectRows.Scan(&yarnID, &name, &planned, &used) == nil {
				projectYarns[yarnID] = append(projectYarns[yarnID], name)
				plannedNeeds[yarnID] += math.Max(0, planned-used)
			}
		}
		items := make([]inventoryShortage, 0)
		for yarnRows.Next() {
			var item inventoryShortage
			if yarnRows.Scan(&item.YarnID, &item.YarnName, &item.RemainingGrams) != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read inventory shortage"})
				return
			}
			plannedGap := math.Max(0, plannedNeeds[item.YarnID]-item.RemainingGrams)
			if item.RemainingGrams < 20 || plannedGap > 0 {
				if plannedGap > 0 {
					item.SuggestedGrams = plannedGap
				} else {
					item.SuggestedGrams = 100 - item.RemainingGrams
				}
				item.ProjectNames = projectYarns[item.YarnID]
				items = append(items, item)
			}
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}

func validateCreateYarn(input createYarnRequest) error {
	if len(strings.TrimSpace(input.Name)) < 1 || len(strings.TrimSpace(input.Name)) > 120 {
		return fmt.Errorf("yarn name must contain 1 to 120 characters")
	}
	if len(input.Brand) > 120 || len(input.Color) > 80 || len(input.Material) > 120 || len(input.Thickness) > 80 {
		return fmt.Errorf("yarn details are too long")
	}
	if input.Plies < 1 || input.Plies > 50 {
		return fmt.Errorf("plies must be between 1 and 50")
	}
	if input.RemainingGrams < 0 || input.UnitPrice < 0 {
		return fmt.Errorf("remaining grams and unit price cannot be negative")
	}
	if input.ColorHex != "" && (len(input.ColorHex) != 7 || input.ColorHex[0] != '#') {
		return fmt.Errorf("color hex must use #RRGGBB")
	}
	if !validImageReference(input.PhotoData) {
		return fmt.Errorf("photo must be an uploaded image")
	}
	return nil
}

func requireSignedIn(db *sql.DB, w http.ResponseWriter, r *http.Request) (user, bool) {
	if db == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database is not configured"})
		return user{}, false
	}
	account, err := sessionUser(db, r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "login is required"})
		return user{}, false
	}
	return account, true
}

func calculatorHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database is not configured"})
			return
		}
		if _, err := sessionUser(db, r); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "login is required"})
			return
		}
		switch r.PathValue("mode") {
		case "circle":
			var input circleCalculatorRequest
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
				return
			}
			result, err := calculateCircle(input)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"data": result})
		case "size":
			var input sizeCalculatorRequest
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
				return
			}
			result, err := calculateSize(input)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"data": result})
		case "yarn":
			var input yarnCalculatorRequest
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
				return
			}
			result, err := calculateYarn(input)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"data": result})
		default:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "calculator mode not found"})
		}
	}
}

func calculateFromRaw(mode string, raw json.RawMessage) (map[string]any, error) {
	switch mode {
	case "circle":
		var input circleCalculatorRequest
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, fmt.Errorf("invalid circle input")
		}
		return calculateCircle(input)
	case "size":
		var input sizeCalculatorRequest
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, fmt.Errorf("invalid size input")
		}
		return calculateSize(input)
	case "yarn":
		var input yarnCalculatorRequest
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, fmt.Errorf("invalid yarn input")
		}
		return calculateYarn(input)
	default:
		return nil, fmt.Errorf("calculator mode not found")
	}
}

func saveCalculationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		var input saveCalculationRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || len(input.Input) == 0 {
			writeJSON(w, 400, map[string]string{"error": "invalid calculation request"})
			return
		}
		result, err := calculateFromRaw(input.Mode, input.Input)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if input.ProjectID != nil && !projectOwned(r.Context(), db, account.ID, *input.ProjectID) {
			writeJSON(w, 400, map[string]string{"error": "project not found"})
			return
		}
		resultData, _ := json.Marshal(result)
		formula, _ := result["formula"].(string)
		response, err := db.ExecContext(r.Context(), `INSERT INTO calculation_records (user_id,project_id,mode,input_data,result_data,formula) VALUES (?,?,?,?,?,?)`, account.ID, input.ProjectID, input.Mode, input.Input, resultData, formula)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not save calculation"})
			return
		}
		id, _ := response.LastInsertId()
		if input.ProjectID != nil {
			_, _ = db.ExecContext(r.Context(), `INSERT INTO project_activities (project_id,kind,detail) VALUES (?,'calculation_saved',?)`, *input.ProjectID, "保存计算记录："+input.Mode)
		}
		writeJSON(w, 201, map[string]any{"data": map[string]int64{"id": id}})
	}
}

func deleteCalculationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid calculation id"})
			return
		}
		result, err := db.ExecContext(r.Context(), `DELETE FROM calculation_records WHERE id=? AND user_id=?`, id, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not delete calculation"})
			return
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			writeJSON(w, 404, map[string]string{"error": "calculation not found"})
			return
		}
		w.WriteHeader(204)
	}
}

func projectCalculationsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || !projectOwned(r.Context(), db, account.ID, projectID) {
			writeJSON(w, 404, map[string]string{"error": "project not found"})
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id,project_id,mode,input_data,result_data,formula,created_at FROM calculation_records WHERE project_id=? AND user_id=? ORDER BY created_at DESC,id DESC`, projectID, account.ID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load calculations"})
			return
		}
		defer rows.Close()
		items := make([]calculationRecord, 0)
		for rows.Next() {
			var item calculationRecord
			if err := rows.Scan(&item.ID, &item.ProjectID, &item.Mode, &item.Input, &item.Result, &item.Formula, &item.CreatedAt); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read calculation"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}

func calculateCircle(input circleCalculatorRequest) (map[string]any, error) {
	if input.StartStitches < 3 || input.StartStitches > 200 || input.Rounds < 1 || input.Rounds > 100 || input.IncreasePerRound < 0 || input.IncreasePerRound > 200 {
		return nil, fmt.Errorf("please enter valid circle parameters")
	}
	rounds := make([]map[string]int, input.Rounds)
	for i := range rounds {
		rounds[i] = map[string]int{"round": i + 1, "stitches": input.StartStitches + i*input.IncreasePerRound}
	}
	return map[string]any{"rounds": rounds, "finalStitches": rounds[len(rounds)-1]["stitches"], "formula": fmt.Sprintf("首圈 %d 针；每圈增加 %d 针", input.StartStitches, input.IncreasePerRound)}, nil
}

func calculateSize(input sizeCalculatorRequest) (map[string]any, error) {
	if input.StitchesPer10CM <= 0 || input.RowsPer10CM <= 0 || input.WidthCM <= 0 || input.HeightCM <= 0 || input.WidthCM > 1000 || input.HeightCM > 1000 {
		return nil, fmt.Errorf("gauge and target size must be greater than zero")
	}
	stitches := int(math.Ceil(input.WidthCM * input.StitchesPer10CM / 10))
	rows := int(math.Ceil(input.HeightCM * input.RowsPer10CM / 10))
	return map[string]any{"castOnStitches": stitches, "rows": rows, "stitchGaugePerCM": input.StitchesPer10CM / 10, "rowGaugePerCM": input.RowsPer10CM / 10, "formula": "目标尺寸 × 每厘米密度，结果向上取整"}, nil
}

func calculateYarn(input yarnCalculatorRequest) (map[string]any, error) {
	if input.MetersPer100G <= 0 || input.PatternMeters <= 0 || input.ReserveRate < 0 || input.ReserveRate > 100 {
		return nil, fmt.Errorf("please enter valid yarn length and reserve")
	}
	base := input.PatternMeters / input.MetersPer100G * 100
	grams := math.Ceil(base * (1 + input.ReserveRate/100))
	return map[string]any{"baseGrams": math.Round(base*10) / 10, "recommendedGrams": grams, "recommendedSkeins": int(math.Ceil(grams / 100)), "formula": fmt.Sprintf("图解用线 ÷ 每 100 克米数 × 100，再预留 %.0f%%", input.ReserveRate)}, nil
}

func normalizeEmail(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newOpaqueToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

// splitMailFrom 拆分“显示名 <地址>”形式，SMTP 信封发件人必须是裸地址，
// 而 From 头可以带显示名。之前直接把带显示名的配置当信封发件人，
// 会被 SMTP 服务器以 501 invalid FROM parameter 拒绝。
func splitMailFrom(from string) (string, string) {
	from = strings.TrimSpace(from)
	if start := strings.Index(from, "<"); start >= 0 {
		if end := strings.Index(from[start:], ">"); end > 0 {
			display := strings.Trim(strings.TrimSpace(from[:start]), "\"")
			address := strings.TrimSpace(from[start+1 : start+end])
			if address != "" {
				return display, address
			}
		}
	}
	return "", from
}

func sendAuthEmail(cfg config, recipient, subject, body string) error {
	if strings.TrimSpace(cfg.smtpHost) == "" {
		log.Printf("SMTP not configured; auth email for %s (%s): %s", recipient, subject, body)
		return nil
	}
	host := cfg.smtpHost
	from := cfg.smtpFrom
	if strings.TrimSpace(from) == "" {
		from = "no-reply@localhost"
	}
	displayName, envelopeFrom := splitMailFrom(from)
	fromHeader := envelopeFrom
	if displayName != "" {
		fromHeader = displayName + " <" + envelopeFrom + ">"
	}
	message := []byte("From: " + fromHeader + "\r\n" + "To: " + recipient + "\r\n" + "Subject: " + subject + "\r\n" + "MIME-Version: 1.0\r\n" + "Content-Type: text/plain; charset=UTF-8\r\n\r\n" + body)
	var auth smtp.Auth
	if cfg.smtpUsername != "" {
		auth = smtp.PlainAuth("", cfg.smtpUsername, cfg.smtpPassword, host)
	}
	port := cfg.smtpPort
	if port == "" {
		port = "587"
	}
	return smtp.SendMail(host+":"+port, auth, envelopeFrom, []string{recipient}, message)
}

const maxMediaBytes int64 = 8 << 20
const mediaQuotaBytes int64 = 100 << 20
const maxPatternBytes int64 = 20 << 20
const patternQuotaBytes int64 = 200 << 20

var errPatternNotFound = errors.New("pattern file not found")

func patternFileURL(id int64) string { return fmt.Sprintf("/api/pattern-files/%d/file", id) }

func patternObjectKey(id int64, mimeType string) string {
	ext := map[string]string{
		"application/pdf": "pdf",
		"image/jpeg":      "jpg",
		"image/png":       "png",
		"image/gif":       "gif",
		"image/webp":      "webp",
	}[mimeType]
	if ext == "" {
		ext = "bin"
	}
	return fmt.Sprintf("pattern-%d.%s", id, ext)
}

func patternMimeAllowed(fileType, mimeType string) bool {
	if fileType == "pdf" {
		return mimeType == "application/pdf"
	}
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return fileType == "image"
	}
	return false
}

// decodeDataURL splits a base64 data URL into its media type and raw bytes.
func decodeDataURL(data string) (string, []byte, error) {
	if !strings.HasPrefix(data, "data:") {
		return "", nil, errors.New("data url expected")
	}
	marker := strings.Index(data, ";base64,")
	if marker < 5 {
		return "", nil, errors.New("base64 payload expected")
	}
	mimeType := strings.ToLower(strings.TrimSpace(data[5:marker]))
	body, err := base64.StdEncoding.DecodeString(data[marker+len(";base64,"):])
	if err != nil {
		return "", nil, err
	}
	if len(body) == 0 {
		return "", nil, errors.New("empty payload")
	}
	return mimeType, body, nil
}

// readPatternObject returns the pattern file bytes for its owner, migrating rows that
// still keep base64 content in pattern_files.file_data into object storage on first read.
func readPatternObject(ctx context.Context, db *sql.DB, store objectStore, userID, fileID int64) ([]byte, string, error) {
	var owner int64
	var objectKey, mimeType, fileData string
	err := db.QueryRowContext(ctx, `SELECT user_id,COALESCE(object_key,''),mime_type,COALESCE(file_data,'') FROM pattern_files WHERE id=?`, fileID).Scan(&owner, &objectKey, &mimeType, &fileData)
	if err != nil {
		return nil, "", errPatternNotFound
	}
	if owner != userID {
		return nil, "", errPatternNotFound
	}
	if objectKey != "" {
		body, err := store.Get(ctx, objectKey)
		return body, mimeType, err
	}
	legacyMime, body, err := decodeDataURL(fileData)
	if err != nil {
		return nil, "", err
	}
	key := patternObjectKey(fileID, legacyMime)
	if err = store.Put(ctx, key, body, legacyMime); err != nil {
		return nil, "", err
	}
	if _, err = db.ExecContext(ctx, `UPDATE pattern_files SET object_key=?,mime_type=?,size_bytes=?,file_data=NULL WHERE id=?`, key, legacyMime, len(body), fileID); err != nil {
		return nil, "", err
	}
	return body, legacyMime, nil
}

// backfillPatternFiles moves a bounded batch of legacy base64 rows into object storage.
func backfillPatternFiles(ctx context.Context, db *sql.DB, store objectStore) {
	rows, err := db.QueryContext(ctx, `SELECT id,user_id FROM pattern_files WHERE COALESCE(object_key,'')='' AND file_data IS NOT NULL ORDER BY id LIMIT 100`)
	if err != nil {
		return
	}
	type pending struct{ id, userID int64 }
	var items []pending
	for rows.Next() {
		var item pending
		if rows.Scan(&item.id, &item.userID) == nil {
			items = append(items, item)
		}
	}
	rows.Close()
	migrated := 0
	for _, item := range items {
		if _, _, err := readPatternObject(ctx, db, store, item.userID, item.id); err == nil {
			migrated++
		}
	}
	if migrated > 0 {
		log.Printf("migrated %d pattern file(s) to object storage", migrated)
	}
}

func mediaPath(root string, id int64) string { return filepath.Join(root, fmt.Sprintf("%d.jpg", id)) }

func newObjectStore(ctx context.Context, cfg config) (objectStore, error) {
	if cfg.mediaDriver == "" || cfg.mediaDriver == "local" {
		return localObjectStore{root: cfg.mediaRoot}, nil
	}
	if cfg.mediaDriver != "s3" {
		return nil, fmt.Errorf("unsupported MEDIA_DRIVER: %s", cfg.mediaDriver)
	}
	if cfg.s3Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required when MEDIA_DRIVER=s3")
	}
	loadOptions := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(cfg.s3Region)}
	if cfg.s3AccessKey != "" || cfg.s3SecretKey != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.s3AccessKey, cfg.s3SecretKey, "")))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.UsePathStyle = cfg.s3ForcePathStyle
		if cfg.s3Endpoint != "" {
			options.EndpointResolver = s3.EndpointResolverFromURL(cfg.s3Endpoint)
		}
	})
	return s3ObjectStore{client: client, bucket: cfg.s3Bucket}, nil
}

func (store localObjectStore) Put(_ context.Context, key string, body []byte, _ string) error {
	if err := os.MkdirAll(store.root, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(store.root, filepath.Base(key)), body, 0o644)
}
func (store localObjectStore) Get(_ context.Context, key string) ([]byte, error) {
	return os.ReadFile(filepath.Join(store.root, filepath.Base(key)))
}
func (store localObjectStore) Delete(_ context.Context, key string) error {
	err := os.Remove(filepath.Join(store.root, filepath.Base(key)))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
func (store s3ObjectStore) Put(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := store.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(store.bucket), Key: aws.String(key), Body: bytes.NewReader(body), ContentType: aws.String(contentType)})
	return err
}
func (store s3ObjectStore) Get(ctx context.Context, key string) ([]byte, error) {
	output, err := store.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(store.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()
	return io.ReadAll(output.Body)
}
func (store s3ObjectStore) Delete(ctx context.Context, key string) error {
	_, err := store.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(store.bucket), Key: aws.String(key)})
	return err
}

func mediaUploadHandler(db *sql.DB, cfg config, store objectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		cleanupDeletedMedia(r.Context(), db, store)
		if err := r.ParseMultipartForm(maxMediaBytes + 1<<20); err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid multipart upload"})
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "image file is required"})
			return
		}
		defer file.Close()
		if header.Size > maxMediaBytes {
			writeJSON(w, 413, map[string]string{"error": "image must be smaller than 8 MB"})
			return
		}
		limited := io.LimitReader(file, maxMediaBytes+1)
		raw, err := io.ReadAll(limited)
		if err != nil || int64(len(raw)) > maxMediaBytes {
			writeJSON(w, 413, map[string]string{"error": "image must be smaller than 8 MB"})
			return
		}
		img, _, err := image.Decode(bytes.NewReader(raw))
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "only JPEG, PNG, or GIF images are supported"})
			return
		}
		img = normalizeImage(img, 2400)
		var encoded bytes.Buffer
		if err := jpeg.Encode(&encoded, img, &jpeg.Options{Quality: 82}); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not compress image"})
			return
		}
		if int64(encoded.Len()) > maxMediaBytes {
			writeJSON(w, 413, map[string]string{"error": "compressed image is still too large"})
			return
		}
		var used int64
		if err := db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(size_bytes),0) FROM media_assets WHERE user_id=? AND deleted_at IS NULL`, account.ID).Scan(&used); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not check storage quota"})
			return
		}
		if used+int64(encoded.Len()) > mediaQuotaBytes {
			writeJSON(w, 413, map[string]string{"error": "storage quota exceeded (100 MB per account)"})
			return
		}
		result, err := db.ExecContext(r.Context(), `INSERT INTO media_assets (user_id,object_key,mime_type,size_bytes) VALUES (?,?,?,?)`, account.ID, "pending", "image/jpeg", encoded.Len())
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not create media object"})
			return
		}
		id, _ := result.LastInsertId()
		objectKey := fmt.Sprintf("%d.jpg", id)
		if _, err = db.ExecContext(r.Context(), `UPDATE media_assets SET object_key=? WHERE id=? AND user_id=?`, objectKey, id, account.ID); err != nil {
			_, _ = db.ExecContext(r.Context(), `DELETE FROM media_assets WHERE id=?`, id)
			writeJSON(w, 500, map[string]string{"error": "could not finalize media object"})
			return
		}
		if err = store.Put(r.Context(), objectKey, encoded.Bytes(), "image/jpeg"); err != nil {
			_, _ = db.ExecContext(r.Context(), `DELETE FROM media_assets WHERE id=?`, id)
			writeJSON(w, 500, map[string]string{"error": "could not store media object"})
			return
		}
		recordAudit(r.Context(), db, account.ID, "media.create", "media", id, map[string]any{"mimeType": "image/jpeg", "sizeBytes": encoded.Len()})
		writeJSON(w, 201, map[string]any{"data": map[string]any{"id": id, "url": fmt.Sprintf("/api/media/%d", id), "mimeType": "image/jpeg", "sizeBytes": encoded.Len()}})
	}
}

func cleanupDeletedMedia(ctx context.Context, db *sql.DB, store objectStore) {
	rows, err := db.QueryContext(ctx, `SELECT id,object_key FROM media_assets WHERE deleted_at IS NOT NULL AND deleted_at < DATE_SUB(NOW(),INTERVAL 30 DAY) LIMIT 100`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var key string
		if rows.Scan(&id, &key) == nil {
			_ = store.Delete(ctx, key)
			_, _ = db.ExecContext(ctx, `DELETE FROM media_assets WHERE id=? AND deleted_at IS NOT NULL`, id)
		}
	}
}

func normalizeImage(src image.Image, maxEdge int) image.Image {
	b := src.Bounds()
	width, height := b.Dx(), b.Dy()
	if width <= maxEdge && height <= maxEdge {
		return src
	}
	scale := float64(maxEdge) / float64(width)
	if height > width {
		scale = float64(maxEdge) / float64(height)
	}
	nw, nh := int(float64(width)*scale), int(float64(height)*scale)
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			sx := b.Min.X + x*width/nw
			sy := b.Min.Y + y*height/nh
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func mediaGetHandler(db *sql.DB, store objectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			http.NotFound(w, r)
			return
		}
		var owner int64
		var mimeType, objectKey string
		var public bool
		if err := db.QueryRowContext(r.Context(), `SELECT user_id,mime_type,object_key,is_public FROM media_assets WHERE id=? AND deleted_at IS NULL`, id).Scan(&owner, &mimeType, &objectKey, &public); err != nil {
			http.NotFound(w, r)
			return
		}
		if !public {
			account, err := sessionUser(db, r)
			if err != nil || account.ID != owner {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}
		data, err := store.Get(r.Context(), objectKey)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
		w.WriteHeader(200)
		_, _ = w.Write(data)
	}
}

func mediaDeleteHandler(db *sql.DB, _ config, store objectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid media id"})
			return
		}
		var objectKey string
		if err := db.QueryRowContext(r.Context(), `SELECT object_key FROM media_assets WHERE id=? AND user_id=? AND deleted_at IS NULL`, id, account.ID).Scan(&objectKey); err != nil {
			http.NotFound(w, r)
			return
		}
		if _, err = db.ExecContext(r.Context(), `UPDATE media_assets SET deleted_at=NOW() WHERE id=? AND user_id=?`, id, account.ID); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not delete media object"})
			return
		}
		recordAudit(r.Context(), db, account.ID, "media.delete", "media", id, map[string]any{})
		_ = store.Delete(r.Context(), objectKey)
		w.WriteHeader(http.StatusNoContent)
	}
}

func registerHandler(db *sql.DB, cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database is not configured"})
			return
		}
		var input registerRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		input.Name = strings.TrimSpace(input.Name)
		input.Email = normalizeEmail(input.Email)
		if len(input.Name) < 2 || len(input.Name) > 80 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name must contain 2 to 80 characters"})
			return
		}
		if !strings.Contains(input.Email, "@") || len(input.Email) > 255 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "a valid email is required"})
			return
		}
		if len(input.Password) < 8 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must contain at least 8 characters"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not secure password"})
			return
		}
		token, err := newOpaqueToken()
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not create verification token"})
			return
		}
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not create account"})
			return
		}
		result, err := tx.ExecContext(r.Context(), `INSERT INTO users (name,email,password_hash) VALUES (?,?,?)`, input.Name, input.Email, string(hash))
		if err != nil {
			tx.Rollback()
			if isDuplicateEntry(err) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "email is already in use"})
			} else {
				writeJSON(w, 500, map[string]string{"error": "could not create account"})
			}
			return
		}
		userID, _ := result.LastInsertId()
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO email_verification_tokens (user_id,token_hash,expires_at) VALUES (?,?,?)`, userID, tokenHash(token), time.Now().Add(24*time.Hour)); err != nil {
			tx.Rollback()
			writeJSON(w, 500, map[string]string{"error": "could not create verification token"})
			return
		}
		if err = tx.Commit(); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not create account"})
			return
		}
		link := strings.TrimRight(cfg.appBaseURL, "/") + "/verify-email?token=" + token
		if env("APP_ENV", "development") != "production" {
			log.Printf("email verification link for %s: %s", input.Email, link)
		}
		if err := sendAuthEmail(cfg, input.Email, "验证你的线禾邮箱", "你好，"+input.Name+"！\n\n请打开以下链接完成邮箱验证：\n"+link+"\n\n链接 24 小时内有效。\n"); err != nil {
			log.Printf("send verification email: %v", err)
		}
		writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{"email": input.Email, "verificationRequired": true}})
	}
}

func verifyEmailHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, 503, map[string]string{"error": "database is not configured"})
			return
		}
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			var in tokenRequest
			_ = json.NewDecoder(r.Body).Decode(&in)
			token = strings.TrimSpace(in.Token)
		}
		if token == "" {
			writeJSON(w, 400, map[string]string{"error": "verification token is required"})
			return
		}
		var userID int64
		err := db.QueryRowContext(r.Context(), `SELECT user_id FROM email_verification_tokens WHERE token_hash=? AND used_at IS NULL AND expires_at>NOW()`, tokenHash(token)).Scan(&userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "verification link is invalid or expired"})
			return
		}
		if _, err = db.ExecContext(r.Context(), `UPDATE users SET email_verified_at=NOW() WHERE id=?`, userID); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not verify email"})
			return
		}
		_, _ = db.ExecContext(r.Context(), `UPDATE email_verification_tokens SET used_at=NOW() WHERE token_hash=?`, tokenHash(token))
		writeJSON(w, 200, map[string]any{"data": map[string]bool{"verified": true}})
	}
}

func requestPasswordResetHandler(db *sql.DB, cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, 503, map[string]string{"error": "database is not configured"})
			return
		}
		var in struct {
			Email string `json:"email"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid request body"})
			return
		}
		email := normalizeEmail(in.Email)
		var userID int64
		var name string
		if err := db.QueryRowContext(r.Context(), `SELECT id,name FROM users WHERE email=?`, email).Scan(&userID, &name); err == nil {
			token, tokenErr := newOpaqueToken()
			if tokenErr == nil {
				if _, tokenErr = db.ExecContext(r.Context(), `DELETE FROM password_reset_tokens WHERE user_id=? OR expires_at<NOW()`, userID); tokenErr == nil {
					_, tokenErr = db.ExecContext(r.Context(), `INSERT INTO password_reset_tokens (user_id,token_hash,expires_at) VALUES (?,?,?)`, userID, tokenHash(token), time.Now().Add(1*time.Hour))
				}
				if tokenErr == nil {
					link := strings.TrimRight(cfg.appBaseURL, "/") + "/reset-password?token=" + token
					if env("APP_ENV", "development") != "production" {
						log.Printf("password reset link for %s: %s", email, link)
					}
					if err := sendAuthEmail(cfg, email, "重置你的线禾密码", "你好，"+name+"！\n\n请打开以下链接设置新密码：\n"+link+"\n\n链接 1 小时内有效。\n"); err != nil {
						log.Printf("send reset email: %v", err)
					}
				}
			}
		}
		writeJSON(w, 200, map[string]any{"data": map[string]bool{"sent": true}})
	}
}

func confirmPasswordResetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, 503, map[string]string{"error": "database is not configured"})
			return
		}
		var in resetPasswordRequest
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid request body"})
			return
		}
		if len(in.Password) < 8 {
			writeJSON(w, 400, map[string]string{"error": "password must contain at least 8 characters"})
			return
		}
		var userID int64
		if err := db.QueryRowContext(r.Context(), `SELECT user_id FROM password_reset_tokens WHERE token_hash=? AND used_at IS NULL AND expires_at>NOW()`, tokenHash(strings.TrimSpace(in.Token))).Scan(&userID); err != nil {
			writeJSON(w, 400, map[string]string{"error": "reset link is invalid or expired"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not secure password"})
			return
		}
		if _, err = db.ExecContext(r.Context(), `UPDATE users SET password_hash=?,email_verified_at=COALESCE(email_verified_at,NOW()) WHERE id=?`, string(hash), userID); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not reset password"})
			return
		}
		_, _ = db.ExecContext(r.Context(), `UPDATE password_reset_tokens SET used_at=NOW() WHERE token_hash=?`, tokenHash(strings.TrimSpace(in.Token)))
		_, _ = db.ExecContext(r.Context(), `DELETE FROM user_sessions WHERE user_id=?`, userID)
		writeJSON(w, 200, map[string]any{"data": map[string]bool{"reset": true}})
	}
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database is not configured"})
			return
		}
		var input loginRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		var account user
		var passwordHash sql.NullString
		var verifiedAt sql.NullTime
		err := db.QueryRowContext(r.Context(), `SELECT id, name, email, role, created_at, password_hash, email_verified_at FROM users WHERE email = ?`, normalizeEmail(input.Email)).Scan(&account.ID, &account.Name, &account.Email, &account.Role, &account.CreatedAt, &passwordHash, &verifiedAt)
		if err != nil || !passwordHash.Valid || bcrypt.CompareHashAndPassword([]byte(passwordHash.String), []byte(input.Password)) != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "email or password is incorrect"})
			return
		}
		if account.Role != "admin" && !verifiedAt.Valid {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "please verify your email before signing in"})
			return
		}
		token, err := createSession(db, r.Context(), account.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create login session"})
			return
		}
		setSessionCookie(w, token, 7*24*time.Hour)
		writeJSON(w, http.StatusOK, map[string]user{"data": account})
	}
}

func logoutHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db != nil {
			if cookie, err := r.Cookie("line_session"); err == nil {
				_, _ = db.ExecContext(r.Context(), `DELETE FROM user_sessions WHERE token = ?`, cookie.Value)
			}
		}
		setSessionCookie(w, "", -time.Hour)
		w.WriteHeader(http.StatusNoContent)
	}
}

func currentUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database is not configured"})
			return
		}
		account, err := sessionUser(db, r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "login is required"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]user{"data": account})
	}
}

func ensureBootstrapAdmin(db *sql.DB, password string) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("bootstrap administrator password is empty")
	}
	var existing sql.NullString
	err := db.QueryRow(`SELECT password_hash FROM users WHERE email = ?`, "admin@line.local").Scan(&existing)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if existing.Valid {
		_, err = db.Exec(`UPDATE users SET email_verified_at = COALESCE(email_verified_at, NOW()) WHERE email = ?`, "admin@line.local")
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE users SET password_hash = ?, email_verified_at = COALESCE(email_verified_at, NOW()) WHERE email = ? AND password_hash IS NULL`, string(hash), "admin@line.local")
	return err
}

func createSession(db *sql.DB, ctx context.Context, userID int64) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	_, err := db.ExecContext(ctx, `DELETE FROM user_sessions WHERE expires_at < NOW()`)
	if err != nil {
		return "", err
	}
	_, err = db.ExecContext(ctx, `INSERT INTO user_sessions (token, user_id, expires_at) VALUES (?, ?, ?)`, token, userID, time.Now().Add(7*24*time.Hour))
	return token, err
}

func setSessionCookie(w http.ResponseWriter, token string, duration time.Duration) {
	http.SetCookie(w, &http.Cookie{Name: "line_session", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: int(duration.Seconds()), Secure: env("APP_ENV", "development") == "production"})
}

func sessionUser(db *sql.DB, r *http.Request) (user, error) {
	cookie, err := r.Cookie("line_session")
	if err != nil || cookie.Value == "" {
		return user{}, fmt.Errorf("missing session")
	}
	var account user
	err = db.QueryRowContext(r.Context(), `SELECT u.id, u.name, u.email, u.role, u.created_at FROM user_sessions s JOIN users u ON u.id = s.user_id WHERE s.token = ? AND s.expires_at > NOW()`, cookie.Value).Scan(&account.ID, &account.Name, &account.Email, &account.Role, &account.CreatedAt)
	return account, err
}

func usersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database is not configured"})
			return
		}
		if !requireAdmin(db, w, r) {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		rows, err := db.QueryContext(ctx, `SELECT id, name, email, role, created_at FROM users ORDER BY created_at DESC, id DESC`)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load users"})
			return
		}
		defer rows.Close()
		items := make([]user, 0)
		for rows.Next() {
			var item user
			if err := rows.Scan(&item.ID, &item.Name, &item.Email, &item.Role, &item.CreatedAt); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read user"})
				return
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load users"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	}
}

func createUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database is not configured"})
			return
		}
		if !requireAdmin(db, w, r) {
			return
		}
		var input createUserRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if err := validateCreateUser(input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not secure password"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		result, err := db.ExecContext(ctx, `INSERT INTO users (name, email, password_hash, role, email_verified_at) VALUES (?, ?, ?, ?, NOW())`, strings.TrimSpace(input.Name), strings.ToLower(strings.TrimSpace(input.Email)), string(hash), input.Role)
		if err != nil {
			if isDuplicateEntry(err) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "email is already in use"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create user"})
			return
		}
		id, _ := result.LastInsertId()
		writeJSON(w, http.StatusCreated, user{ID: id, Name: strings.TrimSpace(input.Name), Email: strings.ToLower(strings.TrimSpace(input.Email)), Role: input.Role})
	}
}

func deleteUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database is not configured"})
			return
		}
		if !requireAdmin(db, w, r) {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		var role string
		if err := db.QueryRowContext(ctx, `SELECT role FROM users WHERE id = ?`, id).Scan(&role); err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load user"})
			return
		}
		if role == "admin" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "administrator accounts cannot be deleted"})
			return
		}
		if _, err := db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not delete user"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func validateCreateUser(input createUserRequest) error {
	if len(strings.TrimSpace(input.Name)) < 2 || len(strings.TrimSpace(input.Name)) > 80 {
		return fmt.Errorf("name must contain 2 to 80 characters")
	}
	if !strings.Contains(strings.TrimSpace(input.Email), "@") || len(input.Email) > 255 {
		return fmt.Errorf("a valid email is required")
	}
	if len(input.Password) < 8 {
		return fmt.Errorf("password must contain at least 8 characters")
	}
	if input.Role != "admin" && input.Role != "user" {
		return fmt.Errorf("role must be admin or user")
	}
	return nil
}

func isDuplicateEntry(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
}

// requireAdmin gates management APIs by the current signed-in user's persisted role.
func requireAdmin(db *sql.DB, w http.ResponseWriter, r *http.Request) bool {
	account, err := sessionUser(db, r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "login is required"})
		return false
	}
	if account.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "administrator role is required"})
		return false
	}
	return true
}

var permissionCatalog = []string{"project.read", "project.write", "project.share", "inventory.read", "inventory.write", "audit.read", "users.manage"}

func effectivePermissions(ctx context.Context, db *sql.DB, account user) (map[string]bool, error) {
	permissions := map[string]bool{}
	if account.Role == "admin" {
		for _, permission := range permissionCatalog {
			permissions[permission] = true
		}
		return permissions, nil
	}
	for _, permission := range []string{"project.read", "project.write", "project.share", "inventory.read", "inventory.write", "audit.read"} {
		permissions[permission] = true
	}
	rows, err := db.QueryContext(ctx, `SELECT permission,allowed FROM user_permissions WHERE user_id=?`, account.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var permission string
		var allowed bool
		if rows.Scan(&permission, &allowed) == nil {
			permissions[permission] = allowed
		}
	}
	return permissions, rows.Err()
}

func permissionHandler(db *sql.DB, permission string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		permissions, err := effectivePermissions(r.Context(), db, account)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not evaluate permissions"})
			return
		}
		if !permissions[permission] {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "permission denied: " + permission})
			return
		}
		next(w, r)
	}
}

func currentPermissionsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		permissions, err := effectivePermissions(r.Context(), db, account)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load permissions"})
			return
		}
		granted := make([]string, 0)
		for _, permission := range permissionCatalog {
			if permissions[permission] {
				granted = append(granted, permission)
			}
		}
		writeJSON(w, 200, map[string]any{"data": map[string]any{"permissions": granted}})
	}
}

func userPermissionsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireAdmin(db, w, r) {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid user id"})
			return
		}
		var target user
		if err = db.QueryRowContext(r.Context(), `SELECT id,name,email,role,created_at FROM users WHERE id=?`, id).Scan(&target.ID, &target.Name, &target.Email, &target.Role, &target.CreatedAt); err != nil {
			writeJSON(w, 404, map[string]string{"error": "user not found"})
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT permission,allowed FROM user_permissions WHERE user_id=?`, id)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load permissions"})
			return
		}
		defer rows.Close()
		overrides := map[string]bool{}
		for rows.Next() {
			var p string
			var a bool
			if rows.Scan(&p, &a) == nil {
				overrides[p] = a
			}
		}
		effective, err := effectivePermissions(r.Context(), db, target)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not evaluate permissions"})
			return
		}
		writeJSON(w, 200, map[string]any{"data": map[string]any{"overrides": overrides, "effective": effective, "catalog": permissionCatalog}})
	}
}

func saveUserPermissionsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		permissions, err := effectivePermissions(r.Context(), db, admin)
		if err != nil || !permissions["users.manage"] {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "permission denied: users.manage"})
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid user id"})
			return
		}
		var input struct {
			Overrides map[string]bool `json:"overrides"`
		}
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid request body"})
			return
		}
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not save permissions"})
			return
		}
		for p, a := range input.Overrides {
			valid := false
			for _, candidate := range permissionCatalog {
				if p == candidate {
					valid = true
					break
				}
			}
			if !valid {
				tx.Rollback()
				writeJSON(w, 400, map[string]string{"error": "unknown permission: " + p})
				return
			}
			if _, err = tx.ExecContext(r.Context(), `INSERT INTO user_permissions (user_id,permission,allowed) VALUES (?,?,?) ON DUPLICATE KEY UPDATE allowed=VALUES(allowed)`, id, p, a); err != nil {
				tx.Rollback()
				writeJSON(w, 500, map[string]string{"error": "could not save permissions"})
				return
			}
		}
		if err = tx.Commit(); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not save permissions"})
			return
		}
		recordAudit(r.Context(), db, admin.ID, "user.permissions.update", "user", id, map[string]any{"overrides": input.Overrides})
		w.WriteHeader(http.StatusNoContent)
	}
}

func recordAudit(ctx context.Context, db *sql.DB, accountID int64, action, resourceType string, resourceID int64, details any) {
	raw, _ := json.Marshal(details)
	_, _ = db.ExecContext(ctx, `INSERT INTO audit_events (user_id,action,resource_type,resource_id,details) VALUES (?,?,?,?,?)`, accountID, action, resourceType, resourceID, raw)
}

func auditEventsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		limit := 100
		rows, err := db.QueryContext(r.Context(), `SELECT id,action,resource_type,resource_id,details,created_at FROM audit_events WHERE user_id=? ORDER BY created_at DESC,id DESC LIMIT ?`, account.ID, limit)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load audit events"})
			return
		}
		defer rows.Close()
		items := make([]map[string]any, 0)
		for rows.Next() {
			var id, resourceID int64
			var action, resourceType string
			var details []byte
			var created time.Time
			if rows.Scan(&id, &action, &resourceType, &resourceID, &details, &created) != nil {
				continue
			}
			var parsed any
			_ = json.Unmarshal(details, &parsed)
			items = append(items, map[string]any{"id": id, "action": action, "resourceType": resourceType, "resourceId": resourceID, "details": parsed, "createdAt": created})
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}

func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		code := http.StatusOK
		if db == nil {
			status, code = "database_not_configured", http.StatusServiceUnavailable
		} else if err := db.PingContext(r.Context()); err != nil {
			status, code = "database_unavailable", http.StatusServiceUnavailable
		}
		writeJSON(w, code, map[string]string{"status": status, "service": "line-api"})
	}
}

func patternsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database is not configured"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		rows, err := db.QueryContext(ctx, `SELECT id, title, category, difficulty, yarn, hook_size, duration_minutes FROM patterns ORDER BY created_at DESC LIMIT 24`)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load patterns"})
			return
		}
		defer rows.Close()
		items := make([]pattern, 0)
		for rows.Next() {
			var item pattern
			if err := rows.Scan(&item.ID, &item.Title, &item.Category, &item.Difficulty, &item.Yarn, &item.HookSize, &item.DurationMinutes); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read pattern"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	}
}

var paletteHexPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// validatePaletteTemplate 校验配色模板：名称必填，2–6 个 #RRGGBB 颜色且不得重复。
func validatePaletteTemplate(name string, colors []string) (string, []string, error) {
	name = strings.TrimSpace(name)
	if len([]rune(name)) < 1 || len([]rune(name)) > 60 {
		return "", nil, fmt.Errorf("名称需要 1 到 60 个字符")
	}
	normalized, err := normalizePaletteColors(colors, 2, 6)
	if err != nil {
		return "", nil, err
	}
	return name, normalized, nil
}

// normalizePaletteColors 统一大小写并校验 #RRGGBB 与重复色。
func normalizePaletteColors(colors []string, minCount, maxCount int) ([]string, error) {
	if len(colors) < minCount || len(colors) > maxCount {
		return nil, fmt.Errorf("需要 %d 到 %d 个颜色", minCount, maxCount)
	}
	normalized := make([]string, 0, len(colors))
	seen := map[string]bool{}
	for _, color := range colors {
		color = strings.ToUpper(strings.TrimSpace(color))
		if !paletteHexPattern.MatchString(color) {
			return nil, fmt.Errorf("颜色需要 #RRGGBB 格式")
		}
		if seen[color] {
			return nil, fmt.Errorf("配色里有重复颜色")
		}
		seen[color] = true
		normalized = append(normalized, color)
	}
	return normalized, nil
}

func scanPaletteTemplate(scan func(...any) error) (paletteTemplate, error) {
	var item paletteTemplate
	var raw []byte
	if err := scan(&item.ID, &item.Name, &item.Style, &item.Occasion, &item.Description, &raw, &item.IsBuiltin, &item.CreatedAt); err != nil {
		return item, err
	}
	item.Colors = []string{}
	_ = json.Unmarshal(raw, &item.Colors)
	for index, color := range item.Colors {
		item.Colors[index] = strings.ToUpper(color)
	}
	return item, nil
}

type paletteSuggestRequest struct {
	Colors []string `json:"colors"`
}
type paletteSuggestion struct {
	Name   string   `json:"name"`
	Rule   string   `json:"rule"`
	Colors []string `json:"colors"`
	Score  float64  `json:"score"`
	Reason string   `json:"reason"`
}
type hslColor struct {
	H float64
	S float64
	L float64
}

func hexToHSL(hex string) hslColor {
	value, err := strconv.ParseUint(strings.TrimPrefix(strings.ToUpper(hex), "#"), 16, 32)
	if err != nil {
		return hslColor{}
	}
	red := float64((value>>16)&0xFF) / 255
	green := float64((value>>8)&0xFF) / 255
	blue := float64(value&0xFF) / 255
	max := math.Max(red, math.Max(green, blue))
	min := math.Min(red, math.Min(green, blue))
	delta := max - min
	var hue float64
	switch {
	case delta == 0:
		hue = 0
	case max == red:
		hue = math.Mod(60*((green-blue)/delta)+360, 360)
	case max == green:
		hue = 60*((blue-red)/delta) + 120
	default:
		hue = 60*((red-green)/delta) + 240
	}
	lightness := (max + min) / 2
	saturation := 0.0
	if delta != 0 {
		saturation = delta / (1 - math.Abs(2*lightness-1))
	}
	return hslColor{H: hue, S: saturation, L: lightness}
}

// hueGap 返回色相在环上的最小距离（0–180）。
func hueGap(left, right float64) float64 {
	gap := math.Abs(left - right)
	if gap > 180 {
		gap = 360 - gap
	}
	return gap
}

// classifyHue 按最大色相跨度判断配色关系，并给出该关系的加成。
func classifyHue(gap float64) (string, float64) {
	switch {
	case gap <= 15:
		return "同色系", 6
	case gap <= 65:
		return "邻近色", 8
	case gap <= 105:
		return "弱对比", 4
	case gap <= 150:
		return "三角色", 8
	default:
		return "互补色", 10
	}
}

// scorePaletteCombination 给一组颜色打分：色相关系 + 明度层次 + 饱和度结构 + 明暗极端惩罚。
// 返回分值、关系名称与一句可展示的理由。
func scorePaletteCombination(colors []string) paletteSuggestion {
	items := make([]hslColor, 0, len(colors))
	for _, color := range colors {
		items = append(items, hexToHSL(color))
	}
	maxGap := 0.0
	maxSat, minSat, avgSat, avgLight := 0.0, 1.0, 0.0, 0.0
	minLight, maxLight := 1.0, 0.0
	highSaturation := 0
	for index, item := range items {
		if item.S > maxSat {
			maxSat = item.S
		}
		if item.S < minSat {
			minSat = item.S
		}
		avgSat += item.S
		avgLight += item.L
		if item.L < minLight {
			minLight = item.L
		}
		if item.L > maxLight {
			maxLight = item.L
		}
		if item.S >= 0.55 {
			highSaturation++
		}
		for other := index + 1; other < len(items); other++ {
			if gap := hueGap(item.H, items[other].H); gap > maxGap {
				maxGap = gap
			}
		}
	}
	avgSat /= float64(len(items))
	avgLight /= float64(len(items))
	lightSpread := maxLight - minLight

	rule, bonus := classifyHue(maxGap)
	score := 68 + bonus
	notes := []string{}

	switch {
	case lightSpread < 0.10:
		score -= 28
		notes = append(notes, "明度几乎一致，容易糊在一起")
	case lightSpread < 0.22:
		score -= 10
		notes = append(notes, "明度差偏小")
	case lightSpread > 0.72:
		score -= 8
		notes = append(notes, "明暗跨度偏大")
	default:
		score += 6
		notes = append(notes, fmt.Sprintf("明度跨度 %.2f，层次清楚", lightSpread))
	}
	if avgSat < 0.10 {
		score -= 16
		notes = append(notes, "整体饱和度偏低")
	}
	if maxSat > 0.9 && avgSat > 0.75 {
		score -= 10
		notes = append(notes, "全部高饱和，容易抢眼")
	}
	if maxSat-minSat >= 0.30 && highSaturation == 1 {
		score += 8
		notes = append(notes, "一个高饱和色做点睛")
	}
	if rule == "互补色" && lightSpread < 0.20 {
		score -= 10
		notes = append(notes, "互补色但明度接近，对比较硬")
	}
	if avgLight < 0.16 {
		score -= 12
		notes = append(notes, "整体偏暗")
	}
	if avgLight > 0.88 {
		score -= 8
		notes = append(notes, "整体偏亮")
	}
	if len(items) == 3 {
		score += 4
	}
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	ordered := orderPaletteByLightness(colors)
	return paletteSuggestion{
		Name:   fmt.Sprintf("%s · %d 色", rule, len(items)),
		Rule:   rule,
		Colors: ordered,
		Score:  math.Round(score*10) / 10,
		Reason: rule + "，" + strings.Join(notes, "，"),
	}
}

// orderPaletteByLightness 由浅到深排列，读起来像「底色 → 主色 → 点缀」。
func orderPaletteByLightness(colors []string) []string {
	ordered := append([]string(nil), colors...)
	sort.SliceStable(ordered, func(left, right int) bool {
		return hexToHSL(ordered[left]).L > hexToHSL(ordered[right]).L
	})
	return ordered
}

// suggestPaletteCombinations 遍历 2–4 色组合，返回得分最高的若干组。
func suggestPaletteCombinations(colors []string, limit int) []paletteSuggestion {
	maxSize := 4
	if len(colors) < maxSize {
		maxSize = len(colors)
	}
	suggestions := make([]paletteSuggestion, 0, 16)
	var walk func(start int, picked []string)
	walk = func(start int, picked []string) {
		if len(picked) >= 2 {
			suggestions = append(suggestions, scorePaletteCombination(append([]string(nil), picked...)))
		}
		if len(picked) == maxSize {
			return
		}
		for index := start; index < len(colors); index++ {
			walk(index+1, append(picked, colors[index]))
		}
	}
	walk(0, nil)

	sort.SliceStable(suggestions, func(left, right int) bool {
		if suggestions[left].Score == suggestions[right].Score {
			return len(suggestions[left].Colors) > len(suggestions[right].Colors)
		}
		return suggestions[left].Score > suggestions[right].Score
	})
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}
	return suggestions
}

func paletteSuggestHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireSignedIn(db, w, r); !ok {
			return
		}
		var input paletteSuggestRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid request body"})
			return
		}
		colors, err := normalizePaletteColors(input.Colors, 2, 8)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"data": map[string]any{"suggestions": suggestPaletteCombinations(colors, 3), "colors": orderPaletteByLightness(colors)}})
	}
}

func paletteTemplatesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		style := strings.TrimSpace(r.URL.Query().Get("style"))
		query := `SELECT id,name,style,occasion,description,colors,is_builtin,created_at FROM palette_templates WHERE (is_builtin=TRUE OR user_id=?)`
		args := []any{account.ID}
		if style != "" {
			query += ` AND style=?`
			args = append(args, style)
		}
		query += ` ORDER BY is_builtin DESC,id`
		rows, err := db.QueryContext(r.Context(), query, args...)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load palette templates"})
			return
		}
		defer rows.Close()
		items := make([]paletteTemplate, 0)
		for rows.Next() {
			item, err := scanPaletteTemplate(rows.Scan)
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read palette template"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}

func createPaletteTemplateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		var input createPaletteTemplateRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid request body"})
			return
		}
		name, colors, err := validatePaletteTemplate(input.Name, input.Colors)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		style := strings.TrimSpace(input.Style)
		if len([]rune(style)) > 40 || len([]rune(strings.TrimSpace(input.Occasion))) > 40 || len([]rune(strings.TrimSpace(input.Description))) > 200 {
			writeJSON(w, 400, map[string]string{"error": "风格、题材或说明过长"})
			return
		}
		raw, _ := json.Marshal(colors)
		result, err := db.ExecContext(r.Context(), `INSERT INTO palette_templates (user_id,name,style,occasion,description,colors,is_builtin) VALUES (?,?,?,?,?,?,FALSE)`, account.ID, name, style, strings.TrimSpace(input.Occasion), strings.TrimSpace(input.Description), raw)
		if err != nil {
			if strings.Contains(err.Error(), "uk_palette_templates_owner_name") || strings.Contains(err.Error(), "Duplicate entry") {
				writeJSON(w, 409, map[string]string{"error": "已有同名配色，换个名字吧"})
				return
			}
			writeJSON(w, 500, map[string]string{"error": "could not save palette template"})
			return
		}
		id, _ := result.LastInsertId()
		recordAudit(r.Context(), db, account.ID, "palette.create", "palette_template", id, map[string]any{"name": name, "colors": colors})
		writeJSON(w, 201, map[string]any{"data": map[string]any{"id": id, "name": name, "colors": colors}})
	}
}

func deletePaletteTemplateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid palette template id"})
			return
		}
		var builtin bool
		if err := db.QueryRowContext(r.Context(), `SELECT is_builtin FROM palette_templates WHERE id=? AND user_id=?`, id, account.ID).Scan(&builtin); err != nil {
			writeJSON(w, 404, map[string]string{"error": "palette template not found"})
			return
		}
		if builtin {
			writeJSON(w, 400, map[string]string{"error": "内置配色不能删除"})
			return
		}
		if _, err = db.ExecContext(r.Context(), `DELETE FROM palette_templates WHERE id=? AND user_id=?`, id, account.ID); err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not delete palette template"})
			return
		}
		recordAudit(r.Context(), db, account.ID, "palette.delete", "palette_template", id, map[string]any{})
		w.WriteHeader(204)
	}
}

func stitchTermsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		category := strings.TrimSpace(r.URL.Query().Get("category"))
		where := "WHERE 1=1"
		args := []any{account.ID}
		if query != "" {
			where += " AND (t.code LIKE ? OR t.name_zh LIKE ? OR t.name_en LIKE ? OR t.description LIKE ?)"
			like := "%" + query + "%"
			args = append(args, like, like, like, like)
		}
		if category != "" {
			where += " AND t.category=?"
			args = append(args, category)
		}
		rows, err := db.QueryContext(r.Context(), `SELECT t.id,t.code,t.name_zh,t.name_en,t.category,t.description,t.instruction,t.caution,EXISTS(SELECT 1 FROM user_stitch_favorites f WHERE f.user_id=? AND f.stitch_term_id=t.id) FROM stitch_terms t `+where+` ORDER BY t.category,t.code`, args...)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not load stitch dictionary"})
			return
		}
		defer rows.Close()
		items := make([]stitchTerm, 0)
		for rows.Next() {
			var item stitchTerm
			if err := rows.Scan(&item.ID, &item.Code, &item.NameZH, &item.NameEN, &item.Category, &item.Description, &item.Instruction, &item.Caution, &item.IsFavorite); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read stitch term"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, map[string]any{"data": items})
	}
}

func favoriteStitchTermHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		termID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || termID < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid stitch term id"})
			return
		}
		result, err := db.ExecContext(r.Context(), `INSERT IGNORE INTO user_stitch_favorites (user_id,stitch_term_id) SELECT ?,id FROM stitch_terms WHERE id=?`, account.ID, termID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not favorite stitch term"})
			return
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			var exists int
			if db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM stitch_terms WHERE id=?`, termID).Scan(&exists) != nil || exists == 0 {
				writeJSON(w, 404, map[string]string{"error": "stitch term not found"})
				return
			}
		}
		w.WriteHeader(204)
	}
}

func unfavoriteStitchTermHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireSignedIn(db, w, r)
		if !ok {
			return
		}
		termID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || termID < 1 {
			writeJSON(w, 400, map[string]string{"error": "invalid stitch term id"})
			return
		}
		_, err = db.ExecContext(r.Context(), `DELETE FROM user_stitch_favorites WHERE user_id=? AND stitch_term_id=?`, account.ID, termID)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "could not unfavorite stitch term"})
			return
		}
		w.WriteHeader(204)
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
