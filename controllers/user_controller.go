package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"online_food_ordering/model"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrUserNotFound = errors.New("user not found")
var secretKey = []byte(os.Getenv("JWT_PRIVATE_KEY"))

func (server *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-www-form-urlencode")
	fmt.Println("Create user route called")
	collection = server.database.Collection("users")

	var user model.User
	_ = json.NewDecoder(r.Body).Decode(&user)
	err := user.Validate()
	if err != nil {
		log.Panic("User data not valid ", err)
		return
	}
	jwtkey, err := createJWT(user.ID.Hex())
	if err != nil {
		fmt.Println("Error in createJWT", err)
	}
	setCookie(&w, jwtkey)
	fmt.Println("key is ", jwtkey)
	if err != nil {
		errorResponse := map[string]string{"error": err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
	} else {
		insertOneUser(user)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user":      user,
			"isSuccess": true,
		})
	}
}

func (server *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	collection = server.database.Collection("users")
	// vars := mux.Vars(r)
	// id := vars["id"]
	token, cokkieErr := readCookie(&w, r)
	if cokkieErr != nil {
		errorResponse := map[string]string{"error": "Unauthosized user"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	id, tokenErr := verifyToken(token)
	if tokenErr != nil {
		fmt.Println("Invalid Token ", tokenErr)
		return
	}
	fmt.Println("Id is ", id)
	var user model.User
	err := getUserByID(id, &user)
	if err != nil {
		if err == ErrUserNotFound {
			errorResponse := map[string]string{"error": "User not found"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		} else {
			errorResponse := map[string]string{"error": "Internal Server error"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		}
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (server *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	collection = server.database.Collection("users")
	vars := mux.Vars(r)
	id := vars["id"]
	var user model.User
	_ = json.NewDecoder(r.Body).Decode(&user)
	err := updateUserByID(id, &user)
	if err != nil {
		if err == ErrUserNotFound {
			errorResponse := map[string]string{"error": "User not found"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		} else {
			errorResponse := map[string]string{"error": "Something went wrong"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		}
		return
	}
	json.NewEncoder(w).Encode(user)
}

func updateUserByID(userId string, user *model.User) error {
	id, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{{"name", user.Name}, {"email", user.Email}, {"mobile_no", user.Mobile_No}}}}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	fmt.Println("Result after modified ", result.ModifiedCount)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}
func getUserByID(userId string, user *model.User) error {
	id, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.D{{"_id", id}}

	err := collection.FindOne(context.Background(), filter).Decode(user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}
func insertOneUser(user model.User) {
	fmt.Println("User collection created")
	inserted, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("One movie inserted ID ", inserted.InsertedID)
}

func createJWT(id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id": id,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("could not assert claim")
	}
	userId := claim["user_id"].(string)
	fmt.Println("UserId from token is ", userId)

	return userId, nil
}

func setCookie(w *http.ResponseWriter, token string) {
	cookie := http.Cookie{
		Name:     "access_token",
		Value:    token,
		HttpOnly: true, // Set the cookie as HTTP-only for security purpose
		Expires:  time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(*w, &cookie)
}

func readCookie(w *http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return "", err
	}
	fmt.Println("Access Token:", cookie.Value)
	return cookie.Value, nil
}
