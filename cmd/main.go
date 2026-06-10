package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"bank-api/internal/handlers"
	"bank-api/internal/middleware"
	"bank-api/internal/repository"
	"bank-api/internal/service"
)

func main() {
	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using environment variables")
	}

	// Настройка логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("🚀 Starting Bank API Service...")

	// Подключение к БД
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/bank?sslmode=disable"
		logger.Warn("Using default database URL")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal("❌ Failed to connect to database: ", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("❌ Failed to ping database: ", err)
	}

	logger.Info("✅ Connected to database")

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	cardRepo := repository.NewCardRepository(db)
	creditRepo := repository.NewCreditRepository(db)

	// HMAC секрет (в реальном проекте из env)
	hmacSecret := []byte(os.Getenv("HMAC_SECRET"))
	if len(hmacSecret) == 0 {
		hmacSecret = []byte("default-hmac-secret-key-change-in-production")
		logger.Warn("Using default HMAC secret")
	}

	// Инициализация сервисов
	userService := service.NewUserService(userRepo, logger)
	accountService := service.NewAccountService(accountRepo, logger)
	transferService := service.NewTransferService(accountRepo, transactionRepo, logger)
	cardService := service.NewCardService(cardRepo, accountRepo, hmacSecret, logger)
	creditService := service.NewCreditService(creditRepo, accountRepo, transactionRepo, getCentralBankRate, logger)

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(userService, logger)
	accountHandler := handlers.NewAccountHandler(accountService, logger)
	transferHandler := handlers.NewTransferHandler(transferService, logger)
	cardHandler := handlers.NewCardHandler(cardService, logger)
	creditHandler := handlers.NewCreditHandler(creditService, logger)

	// Настройка маршрутизации
	router := mux.NewRouter()

	// Публичные маршруты
	router.HandleFunc("/health", healthCheck).Methods("GET")
	router.HandleFunc("/register", userHandler.Register).Methods("POST")
	router.HandleFunc("/login", userHandler.Login).Methods("POST")

	// Защищенные маршруты (требуют JWT)
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	// Маршруты для счетов
	protected.HandleFunc("/accounts", accountHandler.CreateAccount).Methods("POST")
	protected.HandleFunc("/accounts", accountHandler.GetAccounts).Methods("GET")
	protected.HandleFunc("/accounts/{id}", accountHandler.GetAccountByID).Methods("GET")

	// Маршруты для переводов
	protected.HandleFunc("/transfer", transferHandler.Transfer).Methods("POST")
	protected.HandleFunc("/transactions", transferHandler.GetTransactionHistory).Methods("GET")

	// Маршруты для карт
	protected.HandleFunc("/accounts/{account_id}/cards", cardHandler.CreateCard).Methods("POST")
	protected.HandleFunc("/accounts/{account_id}/cards", cardHandler.GetCards).Methods("GET")

	// Маршруты для кредитов
	protected.HandleFunc("/credits", creditHandler.ApplyForCredit).Methods("POST")
	protected.HandleFunc("/credits/{credit_id}/schedule", creditHandler.GetPaymentSchedule).Methods("GET")

	// Запуск шедулера для обработки просроченных платежей
	go startScheduler(creditService, logger)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.WithField("port", port).Info("🌐 Server starting...")
	logger.Info("📍 Available endpoints:")
	logger.Info("   === PUBLIC ===")
	logger.Info("   GET    /health")
	logger.Info("   POST   /register")
	logger.Info("   POST   /login")
	logger.Info("   === PROTECTED (JWT required) ===")
	logger.Info("   POST   /api/accounts")
	logger.Info("   GET    /api/accounts")
	logger.Info("   GET    /api/accounts/{id}")
	logger.Info("   POST   /api/transfer")
	logger.Info("   GET    /api/transactions?account_id={id}")
	logger.Info("   POST   /api/accounts/{account_id}/cards")
	logger.Info("   GET    /api/accounts/{account_id}/cards")
	logger.Info("   POST   /api/credits")
	logger.Info("   GET    /api/credits/{credit_id}/schedule")

	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Fatal("❌ Server failed: ", err)
	}
}

// Health check handler
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

// getCentralBankRate получает ключевую ставку ЦБ РФ через SOAP
// Временно заглушка, в реальном проекте нужно реализовать SOAP запрос
func getCentralBankRate() (float64, error) {
	// TODO: Реализовать SOAP запрос к API ЦБ РФ
	// URL: https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx
	// Метод: KeyRate

	// Временная заглушка - возвращаем текущую ставку ЦБ (~7.5%)
	// В реальном проекте раскомментировать код ниже

	/*
		soapRequest := buildSOAPRequest()
		rawBody, err := sendSOAPRequest(soapRequest)
		if err != nil {
			return 7.5, err
		}
		rate, err := parseXMLResponse(rawBody)
		if err != nil {
			return 7.5, err
		}
		return rate, nil
	*/

	log.Println("Using default central bank rate (7.5%)")
	return 7.5, nil
}

// startScheduler запускает шедулер для обработки просроченных платежей
func startScheduler(creditService *service.CreditService, logger *logrus.Logger) {
	logger.Info("🕐 Scheduler started - will run every 12 hours")

	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	// Запускаем сразу при старте
	logger.Info("Running initial overdue payments check...")
	if err := creditService.ProcessOverduePayments(); err != nil {
		logger.Error("Failed to process overdue payments: ", err)
	}

	// Затем каждые 12 часов
	for range ticker.C {
		logger.Info("Running scheduled payment processing...")
		if err := creditService.ProcessOverduePayments(); err != nil {
			logger.Error("Failed to process overdue payments: ", err)
		}
	}
}

// Функции для SOAP запроса к ЦБ РФ (закомментированы, раскомментировать при реализации)
/*
func buildSOAPRequest() string {
	fromDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
		<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
			<soap12:Body>
				<KeyRate xmlns="http://web.cbr.ru/">
					<fromDate>%s</fromDate>
					<ToDate>%s</ToDate>
				</KeyRate>
			</soap12:Body>
		</soap12:Envelope>`, fromDate, toDate)
}

func sendSOAPRequest(soapRequest string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(
		"POST",
		"https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
		bytes.NewBuffer([]byte(soapRequest)),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SOAP request failed: %v", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func parseXMLResponse(rawBody []byte) (float64, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(rawBody); err != nil {
		return 0, fmt.Errorf("failed to parse XML: %v", err)
	}

	krElements := doc.FindElements("//diffgram/KeyRate/KR")
	if len(krElements) == 0 {
		return 7.5, nil
	}

	rateElement := krElements[0].FindElement("./Rate")
	if rateElement == nil {
		return 7.5, nil
	}

	var rate float64
	fmt.Sscanf(rateElement.Text(), "%f", &rate)
	return rate, nil
}
*/
