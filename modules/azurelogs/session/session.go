package session

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/netip"
	"os"
	"os/user"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
)

const appLogFile = "app.log"

var S *Session

func initConfig() error {
	// Setup logging to app.log file
	logFile, err := os.OpenFile(appLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Set the session's log target to the file instead of stderr
	S.LogTarget = logFile

	// Create slog logger that writes to app.log
	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Set this as the default logger
	slog.SetDefault(logger)

	// Also set it on the session (this will now use the file as target)
	S.SetLogging("info")

	// Initialize Azure authentication using modern non-deprecated libraries
	if err = InitializeAzureAuthentication(S); err != nil {
		slog.Warn("Azure authentication failed (this is expected in demo environments without proper credentials)", "error", err)
		// Don't return error - we want the UI to work even without Azure credentials
	}

	return nil
}

func Init(queryPath *string) (*Session, error) {
	S = &Session{}
	S.Azure = &AZSession{}
	S.LogTarget = os.Stderr

	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user home directory: %w", err)
	}

	S.HomeDir = usr.HomeDir

	if err = initConfig(); err != nil {
		return nil, fmt.Errorf("failed to initialize Azure session configuration: %w", err)
	}

	hOptions := slog.HandlerOptions{AddSource: false}
	hOptions.Level = ProgramLevel
	err = readQueryFile(*queryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read query file %s: %w", *queryPath, err)
	}

	return S, nil
}

var ProgramLevel = new(slog.LevelVar)

func (s *Session) SetLogging(level string) {
	hOptions := slog.HandlerOptions{AddSource: false}

	switch strings.ToUpper(level) {
	case "ERROR":
		ProgramLevel.Set(slog.LevelError)
	case "WARN":
		ProgramLevel.Set(slog.LevelWarn)

	case "INFO":
		ProgramLevel.Set(slog.LevelInfo)

	case "DEBUG":
		ProgramLevel.Set(slog.LevelDebug)
	}

	hOptions.Level = ProgramLevel

	s.Logger = slog.New(slog.NewTextHandler(s.LogTarget, &hOptions))
}

type Session struct {
	App struct {
		SemVer string
	}

	Logger    *slog.Logger
	LogLevel  string
	LogTarget *os.File
	Output    string

	HTTPClient *http.Client
	Host       netip.Addr

	HomeDir         string
	ConfigRoot      string
	ConfigPath      string
	UseTestData     bool
	Azure           *AZSession
	LogQueryClients map[string]*azquery.LogsClient
	QueriesPath     string
	QueryFile       QueryFile
}

type AZClientSecretCredential struct {
	ClientID     string
	ClientSecret string
	TenantID     string
}

type AZSession struct {
	Credential             azcore.TokenCredential
	ClientSecretCredential AZClientSecretCredential
}

// InitializeAzureAuthentication sets up Azure authentication using modern SDK
func InitializeAzureAuthentication(sess *Session) error {
	var err error

	// Read credentials from environment variables
	sess.Azure.ClientSecretCredential.ClientID = os.Getenv("AZURE_CLIENT_ID")
	sess.Azure.ClientSecretCredential.ClientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	sess.Azure.ClientSecretCredential.TenantID = os.Getenv("AZURE_TENANT_ID")

	// Prefer client secret credential if all required environment variables are set
	if sess.Azure.ClientSecretCredential.ClientID != "" &&
		sess.Azure.ClientSecretCredential.ClientSecret != "" &&
		sess.Azure.ClientSecretCredential.TenantID != "" {

		sess.Logger.Info("Using Azure Client Secret credential authentication")
		sess.Azure.Credential, err = azidentity.NewClientSecretCredential(
			sess.Azure.ClientSecretCredential.TenantID,
			sess.Azure.ClientSecretCredential.ClientID,
			sess.Azure.ClientSecretCredential.ClientSecret,
			&azidentity.ClientSecretCredentialOptions{})
		if err != nil {
			sess.Logger.Error("Failed to create client secret credential", "error", err)
			return err
		}
		return nil
	}

	// Fall back to DefaultAzureCredential which tries multiple authentication methods:
	// 1. Environment variables (AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
	// 2. Workload Identity (if running in Azure Kubernetes Service with a managed identity)
	// 3. Managed Identity (if running on Azure VM/App Service/Function App/etc.)
	// 4. Azure CLI (if user is logged in via `az login`)
	// 5. Azure PowerShell (if user is logged in via `Connect-AzAccount`)
	// 6. Interactive browser (if running interactively)
	sess.Logger.Info("Using Azure DefaultAzureCredential authentication chain")
	sess.Azure.Credential, err = azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{})
	if err != nil {
		sess.Logger.Error("Failed to create default Azure credential", "error", err)
		return err
	}

	sess.Logger.Info("Azure authentication initialized successfully")
	return nil
}

func GetLogsClient(sess *Session, subscriptionID string) (*azquery.LogsClient, error) {
	// Initialize the cache map if it doesn't exist
	if sess.LogQueryClients == nil {
		sess.LogQueryClients = make(map[string]*azquery.LogsClient)
	}

	// Check if we already have a cached client for this subscription ID
	if client, exists := sess.LogQueryClients[subscriptionID]; exists {
		sess.Logger.Info("Using cached LogsClient", "subscriptionID", subscriptionID)
		return client, nil
	}

	// Ensure Azure credentials are available
	if sess.Azure.Credential == nil {
		sess.Logger.Error("Azure credentials not initialized", "subscriptionID", subscriptionID)
		return nil, fmt.Errorf("azure credentials not initialized for subscription %s: please set up authentication first", subscriptionID)
	}

	// Create a new client for this subscription ID using modern Azure SDK
	sess.Logger.Info("Creating new LogsClient", "subscriptionID", subscriptionID)
	client, err := azquery.NewLogsClient(sess.Azure.Credential, nil)
	if err != nil {
		sess.Logger.Error("failed to create logs client", "subscriptionID", subscriptionID, "error", err)
		return nil, fmt.Errorf("failed to create Azure Logs client for subscription %s: %w", subscriptionID, err)
	}

	// Cache the client for future use
	sess.LogQueryClients[subscriptionID] = client
	sess.Logger.Info("Cached new LogsClient", "subscriptionID", subscriptionID)

	return client, nil
}
