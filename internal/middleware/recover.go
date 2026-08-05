package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				/////////Лог , который видит разработчик
				slog.Error("HTTP handler panic recovered",
					slog.Any("error", err),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("stack", string(stack)))
				/////////
				/////////Информация , котрую видит пользователь в ответе
				w.Header().Set("Content-type", "application/json") //Браузеру надо понимать, в каком формате пришел ответ для отображения в качестве обычного текста
				w.WriteHeader(http.StatusInternalServerError)      // Статус всегда выносится отдельно от всего тела ответа

				json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
				////////
			}
		}()

		next.ServeHTTP(w, r)
	})
}
