// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"github.com/Unknwon/paginater"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/util"
	"code.gitea.io/gitea/modules/search"
	"code.gitea.io/gitea/modules/indexer"
)

const (
	// tplSearch search page template
	tplSearch base.TplName = "search"
)

type SearchStats struct {
	RepositoriesCount  int64
	CodeCount          int
	IssuesCount        int64
	UsersCount         int64
	OrganizationsCount int64
}

type SearchResults struct {
	Repositories  []*models.Repository
	Organizations []*models.User
	Code          []*search.Result
	Issues        []*models.Issue
	Users         []*models.User
}

type IssueResult struct {
	*models.Issue
	Url    string
}

// SearchAll render search page
func SearchAll(ctx *context.Context) {

	if !ctx.IsSigned || !ctx.User.IsActive{
		ctx.Redirect(setting.AppSubURL + "/user/oauth2/Itsyou.online")
		return
	}

	var (
		total int64
		err error
	)
	searchResults := SearchResults{}
	searchStats := SearchStats{}
	viewType := ctx.Query("type")
	if len(viewType) == 0 {
		viewType = "repositories"
	}
	//TODO handle keyword cases
	keyword := ctx.Query("keyword")

	page := ctx.QueryInt("page")
	if page <= 1 {
		page = 1
	}

	// Search for repositories
	searchResults.Repositories, searchStats.RepositoriesCount, err = models.SearchRepositoryByName(&models.SearchRepoOptions{
		OwnerID: ctx.User.ID,
		Keyword: keyword,
		IyoOrganizations: ctx.User.GetUserOrganizations(),
		Page: page,
		PageSize:  setting.UI.IssuePagingNum,
		OrderBy:   models.SearchOrderByRecentUpdated,
		Private:   true,
	})
	if err != nil {
		ctx.Handle(500, "SearchAll", err)
		return
	}

	// Search for users
	searchResults.Users, searchStats.UsersCount, err = models.SearchUsers(&models.SearchUserOptions{
		Keyword: keyword,
		Type:     models.UserTypeIndividual,
		PageSize: setting.UI.IssuePagingNum,
		IsActive: util.OptionalBoolTrue,
		Page: page,
	})
	if err != nil {
		ctx.Handle(500, "SearchAll", err)
		return
	}

	// Search for Organizations
	searchResults.Organizations, searchStats.OrganizationsCount, err = models.SearchUsers(&models.SearchUserOptions{
		Keyword: keyword,
		Type:     models.UserTypeOrganization,
		PageSize: setting.UI.IssuePagingNum,
		IsActive: util.OptionalBoolTrue,
		Page: page,
	})
	if err != nil {
		ctx.Handle(500, "SearchAll", err)
		return
	}

	// Search for code in all accessed repos
	reposIds, err := models.GetAccessibleRepositories(ctx.User)
	if err != nil {
		ctx.Handle(500, "SearchAll", err)
		return
	}
	searchStats.CodeCount, searchResults.Code, err = search.PerformReposSearch(reposIds, keyword, page, setting.UI.IssuePagingNum)
	if err != nil {
		ctx.Handle(500, "SearchAll", err)
		return
	}

	// Search for issues in all accessed repos
	var issueIDs []int64
	searchStats.IssuesCount, issueIDs, err = indexer.SearchReposIssuesByKeyword(reposIds, keyword)
	if len(issueIDs) > 0 {
		searchResults.Issues, err = models.Issues(&models.IssuesOptions{
			Page:        page,
			PageSize:    setting.UI.IssuePagingNum,
			IssueIDs:    issueIDs,
		})
		if err != nil {
			ctx.Handle(500, "SearchAll", err)
			return
		}
	}

	if viewType == "repositories" {
		total = searchStats.RepositoriesCount
	} else if viewType == "users" {
		total = searchStats.UsersCount
	} else if viewType == "organizations" {
		total = searchStats.OrganizationsCount
	} else if viewType == "code" {
		total = int64(searchStats.CodeCount)
	} else if viewType == "issues" {
		total = searchStats.IssuesCount
	}

	pager := paginater.New(int(total), setting.UI.IssuePagingNum, page, 5)

	ctx.Data["Page"] = pager
	ctx.Data["SearchResults"] = searchResults
	ctx.Data["SearchStats"] = searchStats
	ctx.Data["Keyword"] = keyword
	ctx.Data["ViewType"] = viewType
	ctx.HTML(200, tplSearch)
}
