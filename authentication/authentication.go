package authentication

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"stock/config"
	"stock/function"
	"stock/responses"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx"
)

const (
	sqlselect = `select password from users where username = $1`
)

func Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	conn, err := pgx.Connect(context.Background(), os.Getenv(config.Conf.DbConnect))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var hashedpass string
	err = conn.QueryRow(context.Background(), sqlselect, username).Scan(&hashedpass)
	if err != nil {
		err = responses.ErrUnauthorized
		responses.SendResponse(w, err, nil, nil)
		return

	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedpass), []byte(password))
	if err != nil {
		err = responses.ErrForbidden
		responses.SendResponse(w, err, nil, nil)
		return
	}

	token, err := function.CreateToken(username)
	if err != nil {
		panic(err)
	}
	RoleOfUser := function.SelectRoleOfUser(function.SelectUserID(username))

	item := &responses.UserLogin{
		AccessToken: token,
		Role:        RoleOfUser,
	}

	responses.SendResponse(w, err, item, nil)
}
