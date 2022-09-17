package gcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"kamogawa/cache/gcecache"
	"kamogawa/config"
	"kamogawa/types"
	"kamogawa/types/gcp/gcetypes"
	"log"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

//   "error": {
//     "code": 401,
//     "message": "Request had invalid authentication credentials. Expected OAuth 2 access token, login cookie or other valid authentication credential. See https://developers.google.com/identity/sign-in/web/devconsole-project.",
//     "status": "UNAUTHENTICATED",
//     "details": [
//       {
//         "@type": "type.googleapis.com/google.rpc.ErrorInfo",
//         "reason": "ACCESS_TOKEN_TYPE_UNSUPPORTED",
//         "metadata": {
//           "service": "cloudresourcemanager.googleapis.com",
//           "method": "google.cloudresourcemanager.v1.Projects.ListProjects"
//         }
//       }
//     ]
//   }
// }

func GCPListProjects(db *gorm.DB, user types.User, useCache bool) (*gcetypes.ListProjectResponse, *gcetypes.ErrorResponse) {
	if config.CacheEnabled && useCache {
		responseSuccess, err := gcecache.ReadProjectsCache(db, user)
		if err == nil {
			return responseSuccess, &gcetypes.ErrorResponse{}
		}
	}

	responseSuccess, responseError := GCPListProjectsMain(db, user)
	if responseSuccess == nil {
		return nil, responseError
	}

	if config.CacheEnabled {
		gcecache.WriteProjectsCache(db, user, responseSuccess)
	}

	return responseSuccess, responseError
}

func GCPListProjectsMain(db *gorm.DB, user types.User) (*gcetypes.ListProjectResponse, *gcetypes.ErrorResponse) {
	apiProjectsUrl := "https://cloudresourcemanager.googleapis.com/v1/projects?filter=lifecycleState:ACTIVE"
	log.Printf("User %v\n", user.AccessToken)
	if !user.AccessToken.Valid {
		panic("Access Token expected but not found %v\n")
	}
	var bearer = "Bearer " + user.AccessToken.String

	fmt.Printf("Token %v\n", bearer)

	req, err := http.NewRequest("GET", apiProjectsUrl, nil)
	if err != nil {
		panic("Error while making request to project")
	}
	req.Header.Add("Authorization", bearer)
	// Send req using http Client
	client := &http.Client{}
	apiProjectsResp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERROR] -", err)
	}
	defer apiProjectsResp.Body.Close()

	buf := &bytes.Buffer{}
	tee := io.TeeReader(apiProjectsResp.Body, buf)
	reader1, _ := ioutil.ReadAll(tee)
	reader2, _ := ioutil.ReadAll(buf)

	var responseSuccess gcetypes.ListProjectResponse
	err = json.Unmarshal(reader1, &responseSuccess)
	if err != nil {
		panic(err)
	}
	var responseError gcetypes.ErrorResponse
	err = json.Unmarshal(reader2, &responseError)
	if err != nil {
		panic(err)
	}
	if responseError.Error.Code == 401 {
		if strings.HasPrefix(responseError.Error.Message, "Request had invalid authentication credentials.") {
			return nil, &responseError
		}
	}
	fmt.Printf("response raw %v \n", string(reader1))
	fmt.Printf("Fetched %v projects\n", len(responseSuccess.Projects))
	return &responseSuccess, &responseError
}