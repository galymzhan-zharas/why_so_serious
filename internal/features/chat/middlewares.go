package chat

import (
	"context"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
)

func Authorize(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// authHeader, ok := r.Header["Authorization"]
		// if !ok {
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	w.Write([]byte("Malformed Token"))
		// 	return
		// }

		// token := strings.Split(authHeader[0], " ")[1]

		// userId, userAgent, err := pkg.ParseJWT(token, "my-32-character-ultra-secure-and-ultra-long-secret")
		// if err != nil {
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	w.Write([]byte("Malformed Token"))
		// 	log.Println(err)
		// 	return
		// }

		userId := rand.Int63()
		userAgent := uuid.New().String()
		ctx := context.WithValue(r.Context(), "userId", userId)
		ctx = context.WithValue(ctx, "userAgent", userAgent)

		next.ServeHTTP(w, r.WithContext(ctx))

	})

}
