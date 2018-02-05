// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"github.com/go-xorm/builder"
)

type KanbanRepo struct {
	ID   int64  `json:"id"`
	Name string `json:"full_name"`
}

type KanbanLabel struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type KanbanMilestone struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type KanbanAssignee struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type KanbanIssue struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	RepoID    string  `json:"repo_id"`
	Assignee  string  `json:"assignee"`
	Milestone string  `json:"milestone"`
	Closed    bool    `json:"closed"`
	Index     int64   `json:"index"`
	Labels    []int64 `json:"label_ids"`
}

type KanbanFilter struct {
	Repositories []*KanbanRepo      `json:"repositories"`
	Milestones   []*KanbanMilestone `json:"milestones"`
	Labels       []*KanbanLabel     `json:"labels"`
	Assignees    []*KanbanAssignee  `json:"assignees"`
}

type KanbanIssueOptions struct {
	Page          int64
	ReposIDs      []int64
	LabelsIDs     []int64
	MilestonesIDs []int64
	AssigneesIDs  []int64
	Stages        []string
	State         string
	IsClosed      bool
}

func getKanbanRepos(sess Engine, reposIDs []int64) ([]*KanbanRepo, error) {
	repos := make([]*KanbanRepo, 0)
	err := sess.Table("repository").
		Select("`repository`.id as `id`, (`user`.lower_name || '/' || `repository`.lower_name) AS `name`").
		Join("INNER", "user", "`repository`.owner_id = `user`.id").
		In("`repository`.id", reposIDs).
		OrderBy("repository.lower_name ASC").
		Find(&repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func getKanbanLabels(sess Engine, reposIDs []int64) ([]*KanbanLabel, error) {
	labels := make([]*KanbanLabel, 0)
	err := sess.Table("label").
		Select("`label`.id AS `id`, `label`.name AS name, `label`.color AS color").
		Join("INNER", "repository", "`repository`.id = `label`.repo_id").
		In("`label`.repo_id", reposIDs).
		OrderBy("`label`.name ASC").
		Find(&labels)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func getKanbanMilestones(sess Engine, reposIDs []int64) ([]*KanbanMilestone, error) {
	milestones := make([]*KanbanMilestone, 0)
	err := sess.Table("milestone").
		Select("DISTINCT `milestone`.id AS `id`, `milestone`.name AS name").
		Join("INNER", "issue", "`milestone`.id = `issue`.milestone_id").
		In("`issue`.repo_id", reposIDs).
		OrderBy("`milestone`.name ASC").
		Find(&milestones)
	if err != nil {
		return nil, err
	}
	return milestones, nil
}

func getKanbanAssignees(sess Engine, reposIDs []int64) ([]*KanbanAssignee, error) {
	assignees := make([]*KanbanAssignee, 0)
	err := sess.Table("user").
		Select("`user`.id AS `id`, `user`.lower_name AS name").
		Join("INNER", "issue", "`user`.id = `issue`.assignee_id").
		In("`issue`.repo_id", reposIDs).
		OrderBy("`user`.name ASC").
		Find(&assignees)
	if err != nil {
		return nil, err
	}
	return assignees, nil
}

func (user *User) GetKanbanFilters() (*KanbanFilter, error) {
	reposIds, err := user.GetAccessibleRepositoriesIds()
	if err != nil {
		return nil, err
	}
	sess := x.NewSession()
	defer sess.Close()

	repos, err := getKanbanRepos(sess, reposIds)
	if err != nil {
		return nil, err
	}

	labels, err := getKanbanLabels(sess, reposIds)
	if err != nil {
		return nil, err
	}

	milestones, err := getKanbanMilestones(sess, reposIds)
	if err != nil {
		return nil, err
	}

	assignees, err := getKanbanAssignees(sess, reposIds)
	if err != nil {
		return nil, err
	}

	filters := KanbanFilter{
		Repositories: repos,
		Milestones:   milestones,
		Labels:       labels,
		Assignees:    assignees,
	}

	return &filters, err
}

func (user *User) GetKanbanIssues(opts KanbanIssueOptions) ([]*KanbanIssue, error) {
	sess := x.NewSession()
	defer sess.Close()

	cond := builder.NewCond()

	accessibleReposIDs, err := user.GetAccessibleRepositoriesIds()
	if err != nil {
		return nil, err
	}
	if len(opts.ReposIDs) > 0 {
		reposIDs := getSlicesIntersection(accessibleReposIDs, opts.ReposIDs)
		opts.ReposIDs = make([]int64, len(reposIDs))
		for i := range reposIDs {
			opts.ReposIDs[i] = reposIDs[i].(int64)
		}

	} else {
		opts.ReposIDs = accessibleReposIDs
	}

	cond = cond.And(builder.In("`issue`.repo_id", opts.ReposIDs))

	if len(opts.AssigneesIDs) > 0 {
		cond = cond.And(builder.In("`issue`.assignee_id", opts.AssigneesIDs))
	}

	if len(opts.MilestonesIDs) > 0 {
		cond = cond.And(builder.In("`issue`.milestone_id", opts.MilestonesIDs))
	}

	if len(opts.LabelsIDs) > 0 {
		sess.Join("INNER", "issue_label", "`issue`.id = `issue_label`.issue_id")
		sess.In("`issue_label`.label_id", opts.LabelsIDs)
	}

	cond = cond.And(builder.Eq{"`issue`.is_closed": opts.IsClosed})
	if !opts.IsClosed {
		labelsIDs := make([]int64, 0)

		if len(opts.State) > 0 {
			err := sess.Table("label").
				Select("id").
				Where("name = ?", opts.State).
				Find(&labelsIDs)
			if err != nil {
				return nil, err
			}
			sess.Join("INNER", "issue_label", "`issue`.id = `issue_label`.issue_id")
			sess.In("`issue_label`.label_id", labelsIDs)
		} else if len(opts.Stages) > 0 {
			err := sess.Table("label").
				Select("id").
				In("name", opts.Stages).
				Find(&labelsIDs)
			if err != nil {
				return nil, err
			}
			sess.Join("LEFT", "issue_label", "`issue`.id = `issue_label`.issue_id")
			sess.NotIn("`issue_label`.label_id", labelsIDs)
			sess.Or("`issue_label`.label_id IS NULL")
		}
	}

	issues := make([]*KanbanIssue, 0)
	err = sess.Table("issue").
		Join("LEFT", "user", "`issue`.assignee_id = `user`.id").
		Join("LEFT", "milestone", "`issue`.milestone_id = `milestone`.id").
		Select("`issue`.id AS id, `issue`.is_closed AS closed, `issue`.name AS name"+
			", `issue`.repo_id AS repo_id, `issue`.index AS index"+
			", `user`.lower_name AS assignee, `milestone`.name AS milestone").
		Where(cond).
		OrderBy("`issue`.created_unix DESC").
		Limit(15, (int(opts.Page) - 1) * 15).
		Find(&issues)
	if err != nil {
		return nil, err
	}
	for _, issue := range issues {
		err = sess.Table("issue_label").
			Select("label_id").
			Where("issue_id = ?", issue.ID).
			Find(&issue.Labels)
		if err != nil {
			return nil, err
		}
	}
	return issues, nil
}
