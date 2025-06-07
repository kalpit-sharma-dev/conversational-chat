package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/banking/ai-agents-banking/src/middleware"
	"github.com/gorilla/mux"

	"github.com/banking/ai-agents-banking/src/config"
	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/handlers"
	"github.com/banking/ai-agents-banking/src/services"
)

func main() {
	// Initialize configuration
	cfg := config.New()

	// Initialize DAOs
	sessionDAO := dao.NewSessionDAO()
	accountDAO := dao.NewAccountDAO()
	payeeDAO := dao.NewPayeeDAO()
	transferDAO := dao.NewTransferDAO()
	loanDAO := dao.NewLoanDAO()

	// Initialize Services
	sessionService := services.NewSessionService(sessionDAO)
	intentService := services.NewIntentRecognitionService()
	conversationService := services.NewConversationService()
	llamaService := services.NewLlamaService(cfg.LlamaURL)
	agentService := services.NewAgentService(accountDAO, payeeDAO, transferDAO, loanDAO)
	toolRegistry := services.NewToolRegistry(15 * time.Minute)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(sessionService)
	chatHandler := handlers.NewChatHandler(
		sessionService,
		agentService,
		conversationService,
		intentService,
		llamaService,
		toolRegistry,
	)
	agentsHandler := handlers.NewAgentsHandler(agentService, sessionService)
	healthHandler := handlers.NewHealthHandler(agentService, conversationService, sessionService)
	conversationHandler := handlers.NewConversationHandler(conversationService, sessionService)

	// Register banking tools
	registerBankingTools(toolRegistry, agentService)

	// Initialize middleware
	corsMiddleware := middleware.NewCORSMiddleware()
	authMiddleware := middleware.NewAuthMiddleware(sessionService)
	loggingMiddleware := middleware.NewLoggingMiddleware()

	// Initialize REST API handlers
	transferHandler := handlers.NewTransferHandler(transferDAO, accountDAO)
	accountHandler := handlers.NewAccountHandler(accountDAO)
	payeeHandler := handlers.NewPayeeHandler(payeeDAO)
	loanHandler := handlers.NewLoanHandler(loanDAO)

	// Create Gorilla Mux router
	r := mux.NewRouter()

	// Add global middleware
	r.Use(corsMiddleware.MiddlewareFunc)
	r.Use(loggingMiddleware.MiddlewareFunc)

	// Public routes (no authentication required)
	r.HandleFunc("/auth", authHandler.ServeHTTP).Methods("POST", "OPTIONS")
	r.HandleFunc("/health", healthHandler.ServeHTTP).Methods("GET", "OPTIONS")

	// API v1 subrouter
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(authMiddleware.MiddlewareFunc) // All API v1 routes require authentication

	// Chat routes
	chatRoutes := api.PathPrefix("/chat").Subrouter()
	chatRoutes.HandleFunc("", chatHandler.ServeHTTP).Methods("GET", "POST", "OPTIONS")
	//chatRoutes.HandleFunc("/stream", chatHandler.Stream).Methods("POST")
	//chatRoutes.HandleFunc("/poll/{sessionId}", chatHandler.Poll).Methods("GET")

	// Agent routes
	api.HandleFunc("/agents", agentsHandler.ServeHTTP).Methods("GET")
	api.HandleFunc("/agents/{agentName}", agentsHandler.GetAgentDetails).Methods("GET")

	// Conversation routes
	conversationRoutes := api.PathPrefix("/conversation").Subrouter()
	conversationRoutes.HandleFunc("/history/{sessionId}", conversationHandler.ServeHTTP).Methods("GET")
	conversationRoutes.HandleFunc("/clear/{sessionId}", conversationHandler.ClearHistory).Methods("DELETE")

	// Banking specific routes
	bankingRoutes := api.PathPrefix("/banking").Subrouter()

	// Transfer routes
	transferRoutes := bankingRoutes.PathPrefix("/transfers").Subrouter()
	transferRoutes.HandleFunc("", transferHandler.ListTransfers).Methods("GET")
	transferRoutes.HandleFunc("", transferHandler.CreateTransfer).Methods("POST")
	transferRoutes.HandleFunc("/{transferId}", transferHandler.GetTransfer).Methods("GET")

	// Account routes
	accountRoutes := bankingRoutes.PathPrefix("/accounts").Subrouter()
	accountRoutes.HandleFunc("", accountHandler.ListAccounts).Methods("GET")
	accountRoutes.HandleFunc("/{accountId}", accountHandler.GetAccount).Methods("GET")
	accountRoutes.HandleFunc("/{accountId}/balance", accountHandler.GetBalance).Methods("GET")

	// Payee routes
	payeeRoutes := bankingRoutes.PathPrefix("/payees").Subrouter()
	payeeRoutes.HandleFunc("", payeeHandler.ListPayees).Methods("GET")
	payeeRoutes.HandleFunc("", payeeHandler.CreatePayee).Methods("POST")
	payeeRoutes.HandleFunc("/{payeeId}", payeeHandler.GetPayee).Methods("GET")
	payeeRoutes.HandleFunc("/{payeeId}", payeeHandler.UpdatePayee).Methods("PUT")
	payeeRoutes.HandleFunc("/{payeeId}", payeeHandler.DeletePayee).Methods("DELETE")

	// Loan routes
	loanRoutes := bankingRoutes.PathPrefix("/loans").Subrouter()
	loanRoutes.HandleFunc("/products", loanHandler.ListLoanProducts).Methods("GET")
	loanRoutes.HandleFunc("/applications", loanHandler.ListApplications).Methods("GET")
	loanRoutes.HandleFunc("/applications", loanHandler.CreateApplication).Methods("POST")
	loanRoutes.HandleFunc("/applications/{applicationId}", loanHandler.GetApplication).Methods("GET")
	loanRoutes.HandleFunc("/eligibility", loanHandler.CheckEligibility).Methods("POST")
	loanRoutes.HandleFunc("/calculate-emi", loanHandler.CalculateEMI).Methods("POST")

	// Backward compatibility routes (without /api/v1 prefix)
	r.Handle("/chat", authMiddleware.MiddlewareFunc(http.HandlerFunc(chatHandler.ServeHTTP))).Methods("GET", "POST", "OPTIONS")
	r.Handle("/agents", authMiddleware.MiddlewareFunc(http.HandlerFunc(agentsHandler.ServeHTTP))).Methods("GET")

	// Add route listing endpoint for development
	if cfg.Environment == "development" {
		r.HandleFunc("/routes", listRoutesHandler).Methods("GET")
	}

	// Start cleanup routine
	go sessionService.StartCleanupRoutine()

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Port
	}

	// Ensure port has colon prefix
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🏦 Banking Agents Server starting on port %s", port)
	log.Printf("🔧 Configuration:")
	log.Printf("   Environment: %s", cfg.Environment)
	log.Printf("   LLaMA URL: %s", cfg.LlamaURL)
	log.Printf("   Log Level: %s", cfg.LogLevel)
	log.Printf("")
	log.Printf("📋 Available API endpoints:")
	log.Printf("   POST   /auth - Authentication")
	log.Printf("   GET    /health - Health check")
	if cfg.Environment == "development" {
		log.Printf("   GET    /routes - List all routes (dev only)")
	}
	log.Printf("")
	log.Printf("   API v1 (Protected routes):")
	log.Printf("   POST   /api/v1/chat - Chat with banking assistant")
	log.Printf("   POST   /api/v1/chat/stream - Start streaming chat")
	log.Printf("   GET    /api/v1/chat/poll/{sessionId} - Poll streaming session")
	log.Printf("   GET    /api/v1/agents - List available agents")
	log.Printf("   GET    /api/v1/agents/{agentName} - Get agent details")
	log.Printf("   GET    /api/v1/conversation/history/{sessionId} - Get conversation history")
	log.Printf("   DELETE /api/v1/conversation/clear/{sessionId} - Clear conversation")
	log.Printf("")
	log.Printf("   Banking API:")
	log.Printf("   GET    /api/v1/banking/accounts - List accounts")
	log.Printf("   GET    /api/v1/banking/accounts/{accountId}/balance - Get balance")
	log.Printf("   GET    /api/v1/banking/transfers - List transfers")
	log.Printf("   POST   /api/v1/banking/transfers - Create transfer")
	log.Printf("   GET    /api/v1/banking/payees - List payees")
	log.Printf("   POST   /api/v1/banking/payees - Create payee")
	log.Printf("   GET    /api/v1/banking/loans/products - List loan products")
	log.Printf("   POST   /api/v1/banking/loans/applications - Apply for loan")
	log.Printf("")
	log.Printf("🚀 Server is ready!")

	log.Fatal(srv.ListenAndServe())
}

func registerBankingTools(registry *services.ToolRegistry, agentService *services.AgentService) {
	// Register fund transfer tool
	registry.RegisterTool(&services.FundTransferTool{
		BaseBankingTool: services.BaseBankingTool{
			AgentService: agentService,
		},
	})

	// Register balance check tool
	registry.RegisterTool(&services.BalanceCheckTool{
		BaseBankingTool: services.BaseBankingTool{
			AgentService: agentService,
		},
	})

	// Register add payee tool
	registry.RegisterTool(&services.AddPayeeTool{
		BaseBankingTool: services.BaseBankingTool{
			AgentService: agentService,
		},
	})

	// Register fixed deposit tool
	registry.RegisterTool(&services.FixedDepositTool{
		BaseBankingTool: services.BaseBankingTool{
			AgentService: agentService,
		},
	})

	// Register recurring deposit tool
	registry.RegisterTool(&services.RecurringDepositTool{
		BaseBankingTool: services.BaseBankingTool{
			AgentService: agentService,
		},
	})

	// Register interest rates tool
	registry.RegisterTool(&services.InterestRatesTool{
		BaseBankingTool: services.BaseBankingTool{
			AgentService: agentService,
		},
	})

	// Register weather tool
	registry.RegisterTool(&services.WeatherTool{
		CacheTTL: 15 * time.Minute,
		Cache:    make(map[string]interface{}),
	})
}

func listRoutesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	routes := `{
  "routes": [
    {"method": "POST", "path": "/auth", "description": "Authentication", "protected": false},
    {"method": "GET", "path": "/health", "description": "Health check", "protected": false},
    {"method": "GET", "path": "/routes", "description": "List all routes (dev only)", "protected": false},
    {"method": "POST", "path": "/api/v1/chat", "description": "Chat with banking assistant", "protected": true},
    {"method": "POST", "path": "/api/v1/chat/stream", "description": "Start streaming chat", "protected": true},
    {"method": "GET", "path": "/api/v1/chat/poll/{sessionId}", "description": "Poll streaming session", "protected": true},
    {"method": "GET", "path": "/api/v1/agents", "description": "List available agents", "protected": true},
    {"method": "GET", "path": "/api/v1/agents/{agentName}", "description": "Get agent details", "protected": true},
    {"method": "GET", "path": "/api/v1/conversation/history/{sessionId}", "description": "Get conversation history", "protected": true},
    {"method": "DELETE", "path": "/api/v1/conversation/clear/{sessionId}", "description": "Clear conversation", "protected": true},
    {"method": "GET", "path": "/api/v1/banking/accounts", "description": "List accounts", "protected": true},
    {"method": "GET", "path": "/api/v1/banking/accounts/{accountId}", "description": "Get account details", "protected": true},
    {"method": "GET", "path": "/api/v1/banking/accounts/{accountId}/balance", "description": "Get account balance", "protected": true},
    {"method": "GET", "path": "/api/v1/banking/transfers", "description": "List transfers", "protected": true},
    {"method": "POST", "path": "/api/v1/banking/transfers", "description": "Create transfer", "protected": true},
    {"method": "GET", "path": "/api/v1/banking/transfers/{transferId}", "description": "Get transfer details", "protected": true},
    {"method": "GET", "path": "/api/v1/banking/payees", "description": "List payees", "protected": true},
    {"method": "POST", "path": "/api/v1/banking/payees", "description": "Create payee", "protected": true},
    {"method": "GET", "path": "/api/v1/banking/payees/{payeeId}", "description": "Get payee details", "protected": true},
    {"method": "PUT", "path": "/api/v1/banking/payees/{payeeId}", "description": "Update payee", "protected": true},
    {"method": "DELETE", "path": "/api/v1/banking/payees/{payeeId}", "description": "Delete payee", "protected": true},
    {"method": "GET", "path": "/api/v1/banking/loans/products", "description": "List loan products", "protected": true},
    {"method": "GET", "path": "/api/v1/banking/loans/applications", "description": "List loan applications", "protected": true},
    {"method": "POST", "path": "/api/v1/banking/loans/applications", "description": "Apply for loan", "protected": true},
    {"method": "GET", "path": "/api/v1/banking/loans/applications/{applicationId}", "description": "Get loan application", "protected": true},
    {"method": "POST", "path": "/api/v1/banking/loans/eligibility", "description": "Check loan eligibility", "protected": true},
    {"method": "POST", "path": "/api/v1/banking/loans/calculate-emi", "description": "Calculate EMI", "protected": true}
  ],
  "total": 26,
  "server_info": {
    "framework": "Gorilla Mux",
    "version": "1.0.0",
    "environment": "development"
  }
}`
	w.Write([]byte(routes))
}
