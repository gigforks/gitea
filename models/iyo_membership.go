package models

import "code.gitea.io/gitea/modules/auth/iyo"

type IyoMembership struct {
	ID           int64      `xorm:"pk autoincr"`
	UserID       int64      `xorm:"UNIQUE(s) INDEX NOT NULL"`
	Organization string     `xorm:"UNIQUE(s) INDEX NOT NULL"`
}


// UpdateMembership updates user's membership of iyo organizations after retrieving it from itsyou.online
func (user *User) UpdateMembership() error {

	// Get current user orgs from itsyou.online
	source, _ := GetLoginSourceByID(user.LoginSource)
	sourceCfg := source.OAuth2()
	provider := iyo.New(sourceCfg.ClientID, sourceCfg.ClientSecret, "", "")
	iyoOrgs, _ := provider.GetUserOrganizations(user.LoginName)

	// Delete old organizations that has been removed
	_, err := x.Where("user_id =? ", user.ID).
		NotIn("organization", iyoOrgs).
		Delete(new(IyoMembership))
	if err != nil {
		return err
	}

	// Get the current user orgs from database
	currentOrgs := make([]*IyoMembership, 0)
	err = x.Select("organization").
		Where("user_id =? ", user.ID).
		Find(&currentOrgs)
	if err != nil {
		return err
	}

	// Save only new orgs
	newOrgs := make([]*IyoMembership, 0)
	for _, iyoOrg := range iyoOrgs {

		alreadyExist := false
		for _, currentOrg := range currentOrgs {
			if iyoOrg == currentOrg.Organization {
				alreadyExist = true
				break
			}
		}

		if !alreadyExist {
			newOrgs = append(newOrgs, &IyoMembership{
				UserID:          user.ID,
				Organization:    iyoOrg,
			})
		}
	}

	_, err = x.Insert(newOrgs)
	if err != nil {
		return err
	}

	return nil
}

// GetUserOrganizations Get User organizations from IyoMembership table
func (user *User) GetUserOrganizations() []string {
	userOrgs := make([]string, 0)
	if user != nil && user.IsOAuth2() {
		loginSource, err := GetLoginSourceByID(user.LoginSource)
		if loginSource.OAuth2().Provider == "itsyou.online" && err == nil {
			userMemberships := make([]*IyoMembership, 0)
			err = x.Select("organization").
				Where("user_id =? ", user.ID).
				Find(&userMemberships)
			if err == nil {
				for _, org := range userMemberships {
					userOrgs = append(userOrgs, org.Organization)
				}
			}
		}
	}
	return userOrgs
}
