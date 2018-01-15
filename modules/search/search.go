// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package search

import (
	"bytes"
	gotemplate "html/template"
	"strings"

	"code.gitea.io/gitea/modules/highlight"
	"code.gitea.io/gitea/modules/indexer"
	"code.gitea.io/gitea/modules/util"
	"code.gitea.io/gitea/models"
	"fmt"
)

// Result a search result to display
type Result struct {
	Filename       string
	HighlightClass string
	LineNumbers    []int
	FormattedLines gotemplate.HTML
	FileURL 	   string
}

func indices(content string, selectionStartIndex, selectionEndIndex int) (int, int) {
	startIndex := selectionStartIndex
	numLinesBefore := 0
	for ; startIndex > 0; startIndex-- {
		if content[startIndex-1] == '\n' {
			if numLinesBefore == 1 {
				break
			}
			numLinesBefore++
		}
	}

	endIndex := selectionEndIndex
	numLinesAfter := 0
	for ; endIndex < len(content); endIndex++ {
		if content[endIndex] == '\n' {
			if numLinesAfter == 1 {
				break
			}
			numLinesAfter++
		}
	}

	return startIndex, endIndex
}

func writeStrings(buf *bytes.Buffer, strs ...string) error {
	for _, s := range strs {
		_, err := buf.WriteString(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchResult(result *indexer.RepoSearchResult, startIndex, endIndex int) (*Result, error) {
	startLineNum := 1 + strings.Count(result.Content[:startIndex], "\n")

	var formattedLinesBuffer bytes.Buffer

	contentLines := strings.SplitAfter(result.Content[startIndex:endIndex], "\n")
	lineNumbers := make([]int, len(contentLines))
	index := startIndex
	for i, line := range contentLines {
		var err error
		if index < result.EndIndex &&
			result.StartIndex < index+len(line) &&
			result.StartIndex < result.EndIndex {
			openActiveIndex := util.Max(result.StartIndex-index, 0)
			closeActiveIndex := util.Min(result.EndIndex-index, len(line))
			err = writeStrings(&formattedLinesBuffer,
				`<li>`,
				line[:openActiveIndex],
				`<span class='active'>`,
				line[openActiveIndex:closeActiveIndex],
				`</span>`,
				line[closeActiveIndex:],
				`</li>`,
			)
		} else {
			err = writeStrings(&formattedLinesBuffer,
				`<li>`,
				line,
				`</li>`,
			)
		}
		if err != nil {
			return nil, err
		}

		lineNumbers[i] = startLineNum + i
		index += len(line)
	}
	var (
		fileUrl string
	)
	if result.RepoID > 0 {
		repo, err := models.GetRepositoryByID(int64(result.RepoID))
		if err != nil {
			return nil, err
		}
		fileUrl = fmt.Sprintf("%s/src/branch/%s/%s", repo.HTMLURL(), repo.DefaultBranch, result.Filename)
	}
	return &Result{
		Filename:       result.Filename,
		HighlightClass: highlight.FileNameToHighlightClass(result.Filename),
		LineNumbers:    lineNumbers,
		FormattedLines: gotemplate.HTML(formattedLinesBuffer.String()),
		FileURL:        fileUrl,
	}, nil
}

// PerformSearch perform a search on a repository
func PerformSearch(repoID int64, keyword string, page, pageSize int) (int, []*Result, error) {
	if len(keyword) == 0 {
		return 0, nil, nil
	}

	total, results, err := indexer.SearchRepoByKeyword(repoID, keyword, page, pageSize)
	if err != nil {
		return 0, nil, err
	}

	displayResults := make([]*Result, len(results))

	for i, result := range results {
		startIndex, endIndex := indices(result.Content, result.StartIndex, result.EndIndex)
		displayResults[i], err = searchResult(result, startIndex, endIndex)
		if err != nil {
			return 0, nil, err
		}
	}
	return int(total), displayResults, nil
}

// PerformReposSearch perform a search on all repositories that user can access
func PerformReposSearch(reposIds []int64, keyword string, page, pageSize int) (int, []*Result, error) {
	if len(keyword) == 0 {
		return 0, nil, nil
	}

	total, results, err := indexer.SearchReposByKeyword(reposIds, keyword, page, pageSize)
	if err != nil {
		return 0, nil, err
	}

	displayResults := make([]*Result, len(results))

	for i, result := range results {
		startIndex, endIndex := indices(result.Content, result.StartIndex, result.EndIndex)
		displayResults[i], err = searchResult(result, startIndex, endIndex)
		if err != nil {
			return 0, nil, err
		}
	}
	return int(total), displayResults, nil
}
