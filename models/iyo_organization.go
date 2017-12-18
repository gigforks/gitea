package models

import (
	"fmt"

	"encoding/json"
	"net/http"

	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/cache"
	"github.com/markbates/goth/gothic"
)

type IyoCollaboration struct {
	ID                   int64      `xorm:"pk autoincr"`
	RepoID               int64      `xorm:"UNIQUE(s) INDEX NOT NULL"`
	OrganizationGlobalId string     `xorm:"UNIQUE(s) INDEX NOT NULL"`
	Mode                 AccessMode `xorm:"DEFAULT 2 NOT NULL"`
}

// IyoCollaborator represents an itsyou.online organization with collaboration details.
type IyoCollaborator struct {
	OrganizationGlobalId string
	IyoCollaboration     *IyoCollaboration
}

// ModeI18nKey loads and returns a translation for the UI
func (c *IyoCollaboration) ModeI18nKey() string {
	switch c.Mode {
	case AccessModeRead:
		return "repo.settings.collaboration.read"
	case AccessModeWrite:
		return "repo.settings.collaboration.write"
	case AccessModeAdmin:
		return "repo.settings.collaboration.admin"
	default:
		return "repo.settings.collaboration.undefined"
	}
}

// AddIyoCollaborator adds new Iyo organization collaboration to a repository with default access mode.
func (repo *Repository) AddIyoCollaborator(globalID string) error {
	collaboration := &IyoCollaboration{
		RepoID:               repo.ID,
		OrganizationGlobalId: globalID,
	}

	has, err := x.Get(collaboration)
	if err != nil {
		return err
	} else if has {
		return nil
	}
	collaboration.Mode = AccessModeWrite

	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.InsertOne(collaboration); err != nil {
		return err
	}

	if repo.Owner.IsOrganization() {
		err = repo.recalculateTeamAccesses(sess, 0)
	} else {
		err = repo.recalculateAccesses(sess)
	}
	if err != nil {
		return fmt.Errorf("recalculateAccesses 'team=%v': %v", repo.Owner.IsOrganization(), err)
	}

	return sess.Commit()
}

// GetIyoCollaborators returns the itsyou.online organization collaborators for a repository
func (repo *Repository) GetIyoCollaborators() ([]*IyoCollaborator, error) {
	return repo.getIyoCollaborators(x)
}

// ChangeIyoCollaborationAccessMode sets new access mode for the itsyou.online organization collaboration.
func (repo *Repository) ChangeIyoCollaborationAccessMode(globalID string, mode AccessMode) error {
	// Discard invalid input
	if mode <= AccessModeNone || mode > AccessModeOwner {
		return nil
	}

	collaboration := &IyoCollaboration{
		RepoID:               repo.ID,
		OrganizationGlobalId: globalID,
	}
	has, err := x.Get(collaboration)
	if err != nil {
		return fmt.Errorf("get collaboration: %v", err)
	} else if !has {
		return nil
	}

	if collaboration.Mode == mode {
		return nil
	}
	collaboration.Mode = mode

	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Id(collaboration.ID).AllCols().Update(collaboration); err != nil {
		return fmt.Errorf("update collaboration: %v", err)
	} else if _, err = sess.Exec("UPDATE iyo_collaboration SET mode = ? WHERE organization_global_id = ? AND repo_id = ?", mode, globalID, repo.ID); err != nil {
		return fmt.Errorf("update access table: %v", err)
	}

	return sess.Commit()
}

// DeleteIyoCollaboration removes the itsyou.online organization collaboration relation
// between the user and repository.
func (repo *Repository) DeleteIyoCollaboration(globalID string) (err error) {
	log.Warn("ORG NAME: ", globalID)
	collaboration := &IyoCollaboration{
		RepoID:               repo.ID,
		OrganizationGlobalId: globalID,
	}

	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	if has, err := sess.Delete(collaboration); err != nil || has == 0 {
		return err
	} else if err = repo.recalculateAccesses(sess); err != nil {
		return err
	}

	return sess.Commit()
}

func (repo *Repository) getIyoCollaborations(e Engine) ([]*IyoCollaboration, error) {
	collaborations := make([]*IyoCollaboration, 0)
	return collaborations, e.Find(&collaborations, &IyoCollaboration{RepoID: repo.ID})
}

func (repo *Repository) getIyoCollaborators(e Engine) ([]*IyoCollaborator, error) {
	collaborations, err := repo.getIyoCollaborations(e)
	if err != nil {
		return nil, fmt.Errorf("getIyoCollaborations: %v", err)
	}

	collaborators := make([]*IyoCollaborator, len(collaborations))
	for i, c := range collaborations {
		globalID := c.OrganizationGlobalId
		collaborators[i] = &IyoCollaborator{
			OrganizationGlobalId: globalID,
			IyoCollaboration:     c,
		}
	}
	return collaborators, nil
}

// GetUserOrganizations Get User organizations from goth session
func GetUserOrganizations(request *http.Request, user *User) []string {
	userOrgs := make([]string, 0)
	if user != nil && user.IsOAuth2() {
		loginSource, err := GetLoginSourceByID(user.LoginSource)
		if loginSource.OAuth2().Provider == "itsyou.online" && err == nil {
			sessionKey := loginSource.Name + gothic.SessionName
			session, err := gothic.Store.Get(request, sessionKey)
			if err == nil {
				sessionValue := session.Values[loginSource.Name]
				if sessionValue != nil {
					session := struct {
						Organizations []string `json:"Organizations"`
					}{}
					json.Unmarshal([]byte(sessionValue.(string)), &session)
					userOrgs = session.Organizations
				}
			}

			// The user may be authenticated through jwt using apis
			// so orgs are cached in memory not in session
			if len(userOrgs) == 0 {
				cachedOrgs, err := cache.Get("itsyou.online_" + user.LoginName)
				if err == nil {
					orgs, ok := cachedOrgs.([]string)
					if ok {
						userOrgs = orgs
					}
				}
			}
		}
	}
	return userOrgs
}
