package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (s *Service) GetHelloWorldBasic(w http.ResponseWriter, _ *http.Request) error {
	return httpx.Write(w, http.StatusOK, s.clock.Now())
}

func (s *Service) GetHelloWorld(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get("name")
	return httpx.Write(w, http.StatusOK, models.HelloWorldResponse{
		Response: "Hello " + name,
	})
}

func (s *Service) CreateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	id := r.URL.Query().Get("id")
	err := s.db.CreateUser(ctx, models.DBUser{
		Id:    id,
		Name:  "A name",
		Email: "An email",
	})
	if err != nil {
		return err
	}
	return httpx.Write(w, http.StatusCreated, nil)
}

/*func createToken(username string) (string, error) {
token := jwt.NewWithClaims(jwt.SigningMethodHS256,
jwt.MapClaims{
"username": username,
"exp": time.Now().Add(time.Hour * 24).Unix(),
})

tokenString, err := token.SignedString(secretKey)
if err != nil {
return "", err
}

return tokenString, nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  var u User
  json.NewDecoder(r.Body).Decode(&u)
  fmt.Printf("The user request value %v", u)

  if u.Username == "Chek" && u.Password == "123456" {
    tokenString, err := CreateToken(u.Username)
    if err != nil {
       w.WriteHeader(http.StatusInternalServerError)
       fmt.Errorf("No username found")
     }
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, tokenString)
    return
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprint(w, "Invalid credentials")
  }
}

*/

func (s *Service) signIn(w http.ResponseWriter, r *http.Request) error {
	req := models.SignInRequest{Email: r.PostFormValue("email"), Password: r.PostFormValue("password")}
	return httpx.Write(w, http.StatusOK)

}

func (s *Service) signUp(w http.ResponseWriter, r *http.Request) error {
	return httpx.Write(w, http.StatusOK)
}
