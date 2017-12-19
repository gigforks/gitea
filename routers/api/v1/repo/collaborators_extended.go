package repo

import (
	"strings"

	"code.gitea.io/gitea/modules/context"
)

// func DeleteCollaborator(ctx *context.APIContext, form api.AddCollaboratorOption) {
// 	collaborator, err := models.GetUserByName(ctx.Params(":collaborator"))
// 	if err != nil {
// 		if models.IsErrUserNotExist(err) {
// 			ctx.Error(422, "", err)
// 		} else {
// 			ctx.Error(500, "GetUserByName", err)
// 		}
// 		return
// 	}
//
// 	if err = ctx.Repo.Repository.DeleteCollaboration(collaborator.ID); err != nil {
// 		ctx.Status(204)
// 	}
//
// }

// SearchOrgs returns Itsyou.Online organization names the user is member of.
func SearchOrgs(ctx *context.APIContext) {
	type OrgName struct {
		Name string
	}

	resp := make([]*OrgName, 0)

	if !ctx.IsSigned {
		ctx.JSON(200, map[string]interface{}{
			"ok":   true,
			"data": resp,
		})
		return
	}
	q := ctx.Query("q")
	userOrgs := ctx.User.GetUserOrganizations()

	// if organizations are set on the context, try to add them to the response.
	// else just return the empty list
	if len(userOrgs) > 0 {
		for _, org := range userOrgs {
			if strings.HasPrefix(org, q) {
				resp = append(resp, &OrgName{Name: org})
			}
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": resp,
	})
}
