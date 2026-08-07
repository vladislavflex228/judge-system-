package main

import (
	"context"
	"fmt"
	"judge-system/internal/config"
	"judge-system/internal/database"
	"judge-system/internal/handler"
	"judge-system/internal/middleware"
	"judge-system/internal/repository"
	"judge-system/internal/runner"
	"judge-system/internal/service"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		slog.Error("load config failed", slog.Any("err", err))
		return
	}

	db, err := database.CreatePool(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s", conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName))
	if err != nil {
		slog.Error("load data base failed", slog.Any("err", err))
		return
	}
	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("custom method not allowed"))
	})
	r.Use(middleware.RecoverMiddleware)
	r.Use(middleware.LoggingMiddleware)

	langRepo := repository.NewLanguageRepository(db)
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	testRepo := repository.NewTestRepository(db)
	subRepo := repository.NewSubmissionRepository(db)

	regService := service.NewRegService(userRepo)
	regHandler := handler.NewRegHandler(regService)

	r.Post("/registration", http.HandlerFunc(regHandler.Registration))

	logService := service.NewLoginService(userRepo)
	logHandler := handler.NewLoginHandler(logService)

	r.Post("/login", http.HandlerFunc(logHandler.Login))

	subService := service.NewSubmissionService(subRepo, taskRepo)
	subHandler := handler.NewSubmissionHandler(subService)

	cppRunner := runner.NewCppRunner()
	goRunner := runner.NewGoRunner()
	pyRunner := runner.NewPyRunner()

	runners := make(map[string]runner.LanguageRunner)
	runners["cpp"] = cppRunner
	runners["go"] = goRunner
	runners["python"] = pyRunner

	manager := runner.NewRunnerManager(runners)

	judgeService := service.NewJudgeService(subRepo, langRepo, taskRepo, testRepo, manager)
	judgeHandler := handler.NewTaskHandler(judgeService)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthorizationMiddleware)

		r.Get("/submission", http.HandlerFunc(subHandler.Get))
		r.Post("/submission", http.HandlerFunc(subHandler.Create))
		r.Get("/judge", http.HandlerFunc(judgeHandler.Judge))

	})

	http.ListenAndServe(":8081", r)
}
