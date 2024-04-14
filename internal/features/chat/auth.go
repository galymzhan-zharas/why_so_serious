package chat

import (
	"context"
	"log"
	"net/http"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/chat/database"
	"github.com/NuEventTeam/chat/pkg"
	"github.com/golang-jwt/jwt"
)

var secret = "my-32-character-ultra-secure-and-ultra-long-secret"

type User struct {
	ID           int64   `json:"id"`
	Email        string  `json:"email"`
	Password     string  `json:"password"`
	Firstname    string  `json:"firstname"`
	Lastname     string  `json:"lastname"`
	ProfileImage *string `json:"profileImage"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}

	err := pkg.ReadJSON(w, r, &input)
	if err != nil {
		pkg.BadRequestResponse(w, r, err)
		return
	}

	query := `insert into users(email,password,firstname,lastname) values($1,$2,$3,$4)`

	db := database.DB.GetDb()

	_, err = db.Exec(context.Background(), query, input.Email, input.Password, input.Firstname, input.Lastname)
	if err != nil {
		pkg.ServerErrorResponse(w, r, err)
		return
	}
	err = pkg.WriteJSON(w, http.StatusCreated, pkg.Envelope{"user": input}, nil)
	if err != nil {
		pkg.ServerErrorResponse(w, r, err)
	}

}

func LoginUser(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := pkg.ReadJSON(w, r, &input)
	if err != nil {
		pkg.BadRequestResponse(w, r, err)
		return
	}

	query := qb.Select("id,password,firstname,lastname,email,profile_image").From("users").Where(sq.Eq{"email": input.Email})
	stmt, param, err := query.ToSql()
	if err != nil {
		pkg.BadRequestResponse(w, r, err)
		return
	}

	db := database.DB.GetDb()

	var p string
	var user User

	err = db.QueryRow(context.Background(), stmt, param...).Scan(&user.ID, &p, &user.Firstname, &user.Lastname, &user.Email, &user.ProfileImage)
	if err != nil {
		pkg.BadRequestResponse(w, r, err)
		return
	}

	if p != input.Password {
		pkg.InvalidCredentialsResponse(w, r)
	}
	token, err := GetJWT(user.ID)
	if err != nil {
		pkg.BadRequestResponse(w, r, err)
		return
	}
	err = pkg.WriteJSON(w, http.StatusCreated, pkg.Envelope{"authentication": token, "user": user}, nil)
	if err != nil {
		pkg.ServerErrorResponse(w, r, err)
	}
}

func GetJWT(userID int64) (string, error) {
	var (
		key = []byte(secret)
	)
	log.Println(userID)
	expireTime := time.Now().Add(time.Hour * 24 * 30)
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["userId"] = userID
	claims["exp"] = expireTime.Unix()

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil

}
