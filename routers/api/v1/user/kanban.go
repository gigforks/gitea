// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"fmt"
	"crypto/ecdsa"
	"os"
	"github.com/dgrijalva/jwt-go"
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
	api "code.gitea.io/sdk/gitea"
	"strconv"
	"strings"
)

var jwtPubKey *ecdsa.PublicKey

const (
	iyoPubKey = `-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAES5X8XrfKdx9gYayFITc89wad4usrk0n2
7MjiGYvqalizeSWTHEpnd7oea9IQ8T5oJjMVH5cc0H5tFSKilFFeh//wngxIyny6
6+Vq5t5B0V0Ehy01+2ceEon2Y0XDkIKv
-----END PUBLIC KEY-----`
	tokenName = "kanban"
)

type TokenByJwtOption struct {
	Jwt string `json:"jwt" binding:"Required"`
}

func init() {
	var err error

	jwtPubKey, err = jwt.ParseECPublicKeyFromPEM([]byte(iyoPubKey))
	if err != nil {
		fmt.Printf("failed to parse pub key:%v\n", err)
		os.Exit(1)
	}
}

func GetTokenByJWT(ctx *context.APIContext, form TokenByJwtOption) {
	userName, err := getUserName(form.Jwt)
	if err != nil {
		ctx.Error(500, "Failed to get username", err)
		return
	}

	// Search user should be filtered by loginSource also
	// But in our case -for now- we have only one login source
	user := &models.User{
		LoginName:   userName,
		LoginType:   models.LoginOAuth2,
	}
	userExists, err := models.GetUser(user)
	if err != nil {
		ctx.Error(500, "AccessToken", err.Error())
		return
	}

	if !userExists {
		ctx.Error(500, "AccessToken", "User doesn't exist")
		return
	}

	// Set user organizations from itsyou.online
	user.UpdateMembership()

	token, err := getKanbanToken(user)
	if err != nil {
		ctx.Error(500, "AccessToken", fmt.Errorf("Access Token Error: %v", err.Error()))
		return
	}

	ctx.JSON(200, &api.AccessToken{
		Name: token.Name,
		Sha1: token.Sha1,
	})

}

// getUserName gets the user name from jwt if it is valid
func getUserName(jwtString string) (string, error) {
	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodES384 {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtPubKey, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("Failed to get claims from %v", token.Claims)
	}

	userName, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("Failed to get username from %v", token.Claims)
	}

	return userName, nil
}

// getKanbanToken gets the user kanban token or create new one if not exists
func getKanbanToken(user *models.User) (*models.AccessToken, error) {
	tokens, err := models.ListAccessTokens(user.ID)
	if err != nil {
		return nil, err
	}
	if len(tokens) > 0 {
		for _, token := range tokens {
			if token.Name == tokenName {
				return token, nil
			}
		}
	}
	token := &models.AccessToken{
		UID:  user.ID,
		Name: "kanban",
	}
	if err := models.NewAccessToken(token); err != nil {
		return nil, err
	}
	return token, err
}

func processQueryInt64(items string) ([]int64, error) {
	itemsIDs := make([]int64, 0)
	if len(items) > 0 {
		itemsSliceIDs := strings.Split(items, ",")
		for _, itemID := range itemsSliceIDs {
			itemID = strings.Trim(itemID, " ")
			parsedItemID, err := strconv.ParseInt(itemID, 10, 64)
			if err != nil {
				return nil, err
			}
			itemsIDs = append(itemsIDs, parsedItemID)
		}
	}
	return itemsIDs, nil
}

func GetFilters(ctx *context.APIContext) {
	kanbanFilters, err := ctx.User.GetKanbanFilters()
	if err != nil {
		ctx.Error(500, "Failed to get user repos", err)
		return
	}
	ctx.JSON(200, kanbanFilters)
}

func GetKanbanIssues(ctx *context.APIContext) {
	issueOptions := models.KanbanIssueOptions{}
	var err error

	page := ctx.Query("page")
	if len(page) > 0 {
		issueOptions.Page, err = strconv.ParseInt(page, 10, 64)
		if err != nil {
			ctx.Error(500, "Failed to get user issues", err)
			return
		}
	} else {
		issueOptions.Page = 1
	}

	issueOptions.ReposIDs, err = processQueryInt64(ctx.Query("repos"))
	if err != nil {
		ctx.Error(500, "Failed to get user issues", err)
		return
	}

	issueOptions.MilestonesIDs, err = processQueryInt64(ctx.Query("milestones"))
	if err != nil {
		ctx.Error(500, "Failed to get user issues", err)
		return
	}

	issueOptions.LabelsIDs, err = processQueryInt64(ctx.Query("labels"))
	if err != nil {
		ctx.Error(500, "Failed to get user issues", err)
		return
	}

	issueOptions.AssigneesIDs, err = processQueryInt64(ctx.Query("assignees"))
	if err != nil {
		ctx.Error(500, "Failed to get user issues", err)
		return
	}

	isClosed := strings.Trim(ctx.Query("closed"), " ")
	if len(isClosed) > 0 {
		issueOptions.IsClosed = true
	} else {
		issueOptions.IsClosed = false
	}

	stages := ctx.Query("stages")
	if len(stages) > 0 {
		issueOptions.Stages = make([]string, 0)
		stages := strings.Split(stages, ",")
		for _, stage := range stages {
			stage = strings.Trim(stage, " ")
			issueOptions.Stages = append(issueOptions.Stages, stage)
		}
	}

	state := ctx.Query("state")
	if len(state) > 0 {
		issueOptions.State = strings.Trim(state, " ")
	}

	issues, err := ctx.User.GetKanbanIssues(issueOptions)
	if err != nil {
		ctx.Error(500, "Failed to get user issues", err)
		return
	}
	ctx.JSON(200, issues)
}
