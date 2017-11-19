// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
)

const tplSignInOauth base.TplName = "user/auth/signin_oauth"


// RenderSignInOauth render sign in oauth page
func RenderSignInOauth(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("sign_in")

	// Check auto-login.
	if checkAutoLogin(ctx) {
		return
	}

	orderedOAuth2Names, oauth2Providers, err := models.GetActiveOAuth2Providers()
	if err != nil {
		ctx.Handle(500, "UserSignIn", err)
		return
	}
	ctx.Data["OrderedOAuth2Names"] = orderedOAuth2Names
	ctx.Data["OAuth2Providers"] = oauth2Providers
	ctx.Data["PageIsSignIn"] = true
	ctx.Data["PageIsLoginOauth"] = true

	ctx.HTML(200, tplSignInOauth)
}
