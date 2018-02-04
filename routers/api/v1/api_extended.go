package v1

import (
	"code.gitea.io/gitea/routers/api/v1/org"
	"code.gitea.io/gitea/routers/api/v1/repo"
	"code.gitea.io/gitea/routers/api/v1/user"
	api "code.gitea.io/sdk/gitea"
	"github.com/go-macaron/binding"

	macaron "gopkg.in/macaron.v1"
)

func registerExtendedRoutes(m *macaron.Macaron) {
	bind := binding.Bind

	m.Group("/users", func() {
		m.Get("", user.ListAllUsers)
	})

	m.Post("/user/org", reqToken(), bind(api.CreateOrgOption{}), user.AddMyUserToOrganization)
	m.Post("/users/:username/org", reqToken(), bind(api.CreateOrgOption{}), user.AddUserToOrganization)
	m.Delete("/user/org", reqToken(), bind(api.CreateOrgOption{}), user.DeleteMyUserFromOrganization)
	m.Delete("/users/:username/org", reqToken(), bind(api.CreateOrgOption{}), user.DeleteUserFromOrganization)

	m.Delete("/collaborators/:collaborator", bind(api.AddCollaboratorOption{}), repo.DeleteCollaborator)

	m.Get("/orgs/", reqToken(), org.ListAllOrgs)
	m.Post("/orgs", reqToken(), bind(api.CreateOrgOption{}), org.CreateOrganization)

	m.Get("/orgs/search", repo.SearchOrgs)
	m.Get("/kanban/filters", reqToken(), user.GetFilters)
	m.Get("/kanban/issues", reqToken(), user.GetKanbanIssues)
	m.Post("/token-by-jwt", bind(user.TokenByJwtOption{}), user.GetTokenByJWT)
}
