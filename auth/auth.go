package auth

import (
	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"log"
)

var client *auth.Client

func Init() {
	opt := option.WithCredentialsFile("./serviceAccountKey.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebase: %v\n", err)
	}

	client, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting auth client: %v\n", err)
	}
}

func Verify(idToken string) (*auth.Token, error) {
	token, err := client.VerifyIDToken(idToken)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func UserFromUID(uid string) (*auth.UserRecord, error) {
	return client.GetUser(context.Background(), uid)
}

func UserFromEmail(email string) (*auth.UserRecord, error) {
	return client.GetUserByEmail(context.Background(), email)
}
