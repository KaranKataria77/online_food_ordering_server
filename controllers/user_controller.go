package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"online_food_ordering/consts"
	"online_food_ordering/model"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrUserNotFound = errors.New("user not found")
var secretKey = []byte(os.Getenv("JWT_PRIVATE_KEY"))

type LoginUser struct {
	Email     string `json:"email" validate:"required,email"`
	Mobile_No string `json:"mobile_no"`
}

func collectionExists(ctx context.Context, db *mongo.Database, collectionName string) (bool, error) {
	// ListCollections returns a cursor, and we can use the Next method to check if there are any collections
	cursor, err := db.ListCollections(ctx, bson.M{"name": collectionName})
	if err != nil {
		return false, err
	}
	defer cursor.Close(ctx)

	// Attempt to retrieve the next result from the cursor
	return cursor.Next(ctx), nil
}

func (server *Server) initUserCollection() {
	isExist, err := collectionExists(context.Background(), server.database, "users")
	if err != nil {
		log.Panic("Error while reading users collections")
	}
	if !isExist {
		collection = server.database.Collection("users")
		indexOptions := options.Index().SetUnique(true)
		index := mongo.IndexModel{
			Keys:    bson.M{"email": 1},
			Options: indexOptions,
		}
		_, err = collection.Indexes().CreateOne(context.TODO(), index)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (server *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create user route called")

	collection = server.database.Collection("users")

	var user model.User
	_ = json.NewDecoder(r.Body).Decode(&user)
	err := user.Validate()

	if err != nil {
		consts.SendErrorResponse(&w, http.StatusBadRequest, consts.ErrorRequiredFieldMissing, err)
	} else {
		userId := insertOneUser(user)
		jwtkey, err := createJWT(userId)
		fmt.Println("User Id after creation ", userId)
		if err != nil {
			consts.SendErrorResponse(&w, http.StatusInternalServerError, consts.ErrorCreatingJWTToken, err)
			return
		}
		setCookie(&w, jwtkey)
		fmt.Println("key is ", jwtkey)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user":      user,
			"isSuccess": true,
		})
	}
}

func (server *Server) UserLogin(w http.ResponseWriter, r *http.Request) {
	var user LoginUser
	collection = server.database.Collection("users")
	_ = json.NewDecoder(r.Body).Decode(&user)
	loggedInUser, loginErr := checkLoginUser(&user)
	if loginErr != nil {
		consts.SendErrorResponse(&w, http.StatusNotFound, consts.ErrorUserNotFound, loginErr)
		return
	}
	userMap, _ := loggedInUser.(map[string]interface{})
	userField, _ := userMap["user"].(model.User)

	fmt.Println("User Logged in is ", userField.ID)
	userId := userField.ID.Hex()
	jwtkey, err := createJWT(userId)
	if err != nil {
		fmt.Println("Error in createJWT", err)
	}
	setCookie(&w, jwtkey)
	json.NewEncoder(w).Encode(loggedInUser)
}

func (server *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	collection = server.database.Collection("users")
	token, cokkieErr := readCookie(r)
	fmt.Println("Reading token from cookies")
	if cokkieErr != nil {
		fmt.Println("Error reading in cookies ", cokkieErr)
		errorResponse := map[string]string{"error": "Unauthorized user"}
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

// services
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
func checkLoginUser(user *LoginUser) (interface{}, error) {
	filter := bson.D{{"email", user.Email}, {"mobile_no", user.Mobile_No}}

	var loggedInUser model.User
	err := collection.FindOne(context.Background(), filter).Decode(&loggedInUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return map[string]interface{}{"message": "no user found"}, err
		}
	}
	return map[string]interface{}{"user": loggedInUser}, nil
}
func insertOneUser(user model.User) string {
	fmt.Println("User collection created")
	inserted, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		fmt.Println("Error while creating user ", err.Error())
	}

	fmt.Println("One user inserted ID ", inserted.InsertedID)
	return inserted.InsertedID.(primitive.ObjectID).Hex()
}

func createJWT(id string) (string, error) {
	fmt.Println("Setting  up JWT", id)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id": id,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	fmt.Println("Token string ", tokenString)
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
	fmt.Println("Setting cookies", token)
	cookie := http.Cookie{
		Name:     "user_token",
		Value:    token,
		HttpOnly: true, // Set the cookie as HTTP-only for security purpose
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		Domain:   "localhost:3000",
		// Secure:   true, // Set to true when using HTTPS
	}
	http.SetCookie(*w, &cookie)
}

func readCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("user_token")
	fmt.Println("Reading cookies ", cookie)
	if err != nil {
		return "", err
	}
	fmt.Println("Access Token:", cookie.Value)
	return cookie.Value, nil
}

func getUserIdFromToken(r *http.Request) (string, error) {
	token, cokkieErr := readCookie(r)
	fmt.Println("Reading token from cookies")
	if cokkieErr != nil {
		fmt.Println("Error reading in cookies ", cokkieErr)
		errorResponse := errors.New("Unauthorized User")
		return "", errorResponse
	}
	id, tokenErr := verifyToken(token)
	if tokenErr != nil {
		fmt.Println("Invalid Token ", tokenErr)
		return "", tokenErr
	}
	return id, nil
}
