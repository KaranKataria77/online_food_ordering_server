package consts

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	ErrorUserNotFound         = "User not Found"
	ErrorNotAuthorized        = "User not Authorized"
	ErrorRequiredFieldMissing = "Required field are missing"
	ErrorCreatingJWTToken     = "Error while creating JWT token"
	ErrorInsertingNewUser     = "Error while creating new user"
)
const (
	ErrorUpdatingCart = "Error while updating cart"
)

const InternalServerError = "Something went wrong"

func SendErrorResponse(w *http.ResponseWriter, statusCode int, name string, err error) {
	log.Panic(name)
	errMessage := map[string]interface{}{
		"name":  name,
		"error": err.Error(),
	}
	(*w).WriteHeader(statusCode)
	json.NewEncoder(*w).Encode(errMessage)
}
