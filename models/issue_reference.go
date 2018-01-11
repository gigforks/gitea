package models

import (
	"regexp"
	"fmt"
	"strings"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"strconv"
	"code.gitea.io/gitea/modules/log"
)

type IssueReference struct {
	ID             int64   `xorm:"pk autoincr"`
	Url            string  `xorm:"NOT NULL"`
	IssueID        int64   `xorm:"NOT NULL"`
	CommentID      int64   `xorm:""`
	Title          string  `xorm:"NOT NULL"`
	Owner          string  `xorm:"NOT NULL"`
	Repo           string  `xorm:"NOT NULL"`
	Type           string  `xorm:"NOT NULL"`
	RefIssueNumber int64   `xorm:"NOT NULL"`
}

type jsonIssue struct {
	Title string `json:"title"`
}


// UpdateRefs updates references to other issues mentioned in an issue or comment
func (issue *Issue) updateRefs(e Engine, commentID int64) (error) {
	var content string
	if commentID > 0 {
		comment, err := GetCommentByID(commentID)
		if err != nil {
			return err
		}
		content = comment.Content
	} else {
		content = issue.Content
	}
	giteaIssueRefs := issue.getGiteaIssuesRefs(e, content)
	githubIssueRefs := issue.getGithubIssuesRefs(content)
	allRefIssues := append(giteaIssueRefs, githubIssueRefs...)

	// Get the current refs from database to add new refs and delete removed ones
	currentRefIssues := make([] *IssueReference, 0)
	err := e.Where("issue_id = ? and comment_id = ?", issue.ID, commentID).Find(&currentRefIssues)
	if err != nil {
		return err
	}

	// Delete removed refs
	issueRefsToDelete := make([]int64, 0)
	for _, currentRefIssue := range currentRefIssues {
		toDelete := true
		for _, refIssue := range allRefIssues {
			if len(refIssue.Title) == 0{
				continue
			}
			if isEqual(&refIssue, currentRefIssue) {
				toDelete = false
				break
			}
		}
		if toDelete {
			issueRefsToDelete = append(issueRefsToDelete, currentRefIssue.ID)
		}
	}
	_, err = e.In("id", issueRefsToDelete).Delete(new(IssueReference))
	if err != nil {
		return err
	}

	// Add new refs
	issueRefsToAdd := make([] IssueReference, 0)
	for _, refIssue := range allRefIssues {
		if len(refIssue.Title) <= 0{
			continue
		}
		toAdd := true
		for _, currentRefIssue := range currentRefIssues {
			if isEqual(&refIssue, currentRefIssue){
				toAdd = false
				break
			}
		}
		if toAdd {
			refIssue.CommentID = commentID
			issueRefsToAdd = append(issueRefsToAdd, refIssue)
		}
	}
	_, err = e.Insert(issueRefsToAdd)
	if err != nil {
		return err
	}

	return nil
}

//updateIssueRefs updates references to other issues mentioned in an issue
func (issue *Issue) updateIssueRefs(e Engine) (error) {
	err := issue.updateRefs(e, 0)
	if err != nil {
		return err
	}
	return nil
}

//updateIssueCommentRefs updates references to other issues mentioned in a comment
func (comment *Comment) updateIssueCommentRefs(e Engine) (error) {
	issue, err := GetIssueByID(comment.IssueID)
	if err != nil {
		return err
	}
	err = issue.updateRefs(e, comment.ID)
	if err != nil {
		return err
	}
	return nil
}

//deleteIssueCommentRefs deletes references to other issues mentioned in a comment when the comment is deleted
func (comment *Comment) DeleteIssueCommentRefs(e Engine) (error) {
	// Get the current refs from database to add new refs and delete removed ones
	_, err := e.Where("issue_id = ? and comment_id = ?", comment.IssueID, comment.ID).Delete(new(IssueReference))
	if err != nil {
		return err
	}
	return nil
}

//getGiteaIssuesRefs parses the gitea issues referenced in the issue/comment content
func (issue *Issue) getGiteaIssuesRefs(e Engine, content string) ([] IssueReference) {
	giteaRegex, _ := regexp.Compile("([^ ]*/[^ ]*)?#([0-9]+)")
	giteaMatches := giteaRegex.FindAllStringSubmatch(content, -1)
	giteaIssueRefs := make([]IssueReference, len(giteaMatches))
	for i, giteaMatch := range giteaMatches {
		var url string
		var repoName string
		var ownerName string
		repoFullName := string(giteaMatch[1])
		issueNumber, _ := strconv.ParseInt(giteaMatch[2], 10, 0)
		title := "ISSUE #" + string(issueNumber)
		if len(repoFullName) > 0 {
			fullName := strings.Split(repoFullName, "/")
			ownerName, repoName = fullName[0], fullName[1]
			url += "/" + repoFullName + "/"
			repository, err := GetRepositoryByOwnerAndName(ownerName, repoName)
			if err != nil {
				continue
			}
			giteaIssue, err := GetIssueByIndex(repository.ID, issueNumber)
			if err != nil {
				continue
			}
			title = giteaIssue.Title
		} else {
			issue.loadAttributes(e)
			issue.Repo.getOwner(e)
			ownerName = issue.Repo.Owner.LowerName
			repoName = issue.Repo.Name
			giteaIssue, err := GetIssueByIndex(issue.Repo.ID, issueNumber)
			if err != nil {
				continue
			}
			title = giteaIssue.Title
		}
		url += giteaMatch[2]
		giteaIssueRefs[i] = IssueReference{
			Url: url,
			IssueID: issue.ID,
			Title: title,
			Owner: ownerName,
			Repo: repoName,
			Type: "gitea",
			RefIssueNumber: issueNumber,
		}
	}

	return giteaIssueRefs
}

//getGiteaIssuesRefs parses the github issues referenced in the issue/comment content
func (issue *Issue) getGithubIssuesRefs(content string) ([] IssueReference) {
	githubRegex, _ := regexp.Compile("github.com/([^ ]+)/([^ ]+)/issues/([0-9]+)")
	githubMatches := githubRegex.FindAllStringSubmatch(content, -1)
	githubIssueRefs := make([]IssueReference, len(githubMatches))
	for i, githubMatch := range githubMatches {
		ownerName := string(githubMatch[1])
		repoName := string(githubMatch[2])
		issueNumber, _ := strconv.ParseInt(githubMatch[3], 10, 0)
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", ownerName, repoName, issueNumber)
		res, err := http.Get(url)
		if err != nil {
			continue
		}
		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			continue
		}
		title := fmt.Sprintf("ISSUE %d", issueNumber)
		if res.StatusCode == 200 {
			issueDetails := jsonIssue{}
			err = json.Unmarshal(body, &issueDetails)
			if err == nil {
				title = issueDetails.Title
			}
		}

		githubIssueRefs[i] = IssueReference{
			Url: "https://" + string(githubMatch[0]),
			IssueID: issue.ID,
			Title: title,
			Owner: ownerName,
			Repo: repoName,
			Type: "github",
			RefIssueNumber: issueNumber,
		}
	}
	return githubIssueRefs
}

//GetRefIssues lists all not duplicated issue references to an issue
func (issue *Issue) GetRefIssues() ([] *IssueReference) {
	// Get the current refs from database
	currentRefIssues := make([] *IssueReference, 0)
	err := x.Where("issue_id = ?", issue.ID).Find(&currentRefIssues)
	if err != nil {
		log.Info("Could not load issues references for issue #%d: %s", issue.ID, err)
	}
	filteredRefIssues := make([] *IssueReference, 0)
	for _, currentRefIssue := range currentRefIssues {
		addIssue := true
		for _, filteredRefIssue := range filteredRefIssues {
			if isEqual(currentRefIssue, filteredRefIssue){
				addIssue = false
				break
			}
		}
		if addIssue {
			filteredRefIssues = append(filteredRefIssues, currentRefIssue)
		}
	}
	return filteredRefIssues
}

// Compare two issueReferences
func isEqual(issue1, issue2 *IssueReference) (bool) {
	if issue1.Type == issue2.Type && issue1.RefIssueNumber == issue2.RefIssueNumber &&
		issue1.Repo == issue2.Repo && issue1.Owner == issue2.Owner {
		return true
	}
	return false
}
