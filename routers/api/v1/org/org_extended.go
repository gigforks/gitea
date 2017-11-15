package org

import (
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/routers/api/v1/convert"
	api "code.gitea.io/sdk/gitea"
)

func ListAllOrgs(ctx *context.APIContext) {
	lenOrgs := int(models.CountOrganizations())

	orgSearchOpts := &models.SearchUserOptions{
		Keyword:  "_",
		Type:     models.UserTypeOrganization,
		PageSize: lenOrgs,
		Page:     0,
	}
	// ALL IN ONE PAGE.
	if users, _, err := models.SearchUsers(orgSearchOpts); err == nil {
		results := make([]*api.User, len(users))
		for i := range users {
			results[i] = &api.User{
				ID:        users[i].ID,
				UserName:  users[i].Name,
				AvatarURL: users[i].AvatarLink(),
				FullName:  users[i].FullName,
			}
			if ctx.IsSigned {
				results[i].Email = users[i].Email
			}
		}
		ctx.JSON(200, results)
	}

}

func CreateOrganization(ctx *context.APIContext, form api.CreateOrgOption) {
	orgInfo := &models.User{
		Name:        form.UserName,
		Description: form.Description,
		IsActive:    true,
		Type:        models.UserTypeOrganization,
	}
	currentUser := ctx.User
	if err := models.CreateOrganization(orgInfo, currentUser); err != nil {
		ctx.JSON(500, "Couldn't create organization.")
	}
	ctx.JSON(200, convert.ToOrganization(orgInfo))
}
