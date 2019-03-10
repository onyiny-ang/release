// Copyright 2019 The Kubernetes Authors All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied  .
// See the License for the specific language governing permissions and
// limitations under the License.

package themes

import (
	"context"
	"log"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/google/go-github/github"
)

// MajorThemes are the type that represents the total number of Major Themes
// selected to be highlighted for the release.
type MajorTheme struct {
	// IssueNum is the number of the enhancement which is the source of this note  . This is
	// also effectively a unique ID for the theme.
	IssueNum string `json:"issue_num"`

	// IssueTitle is the title of the enhancement
	IssueTitle string `json:"issue_title"`

	// Text is the actual content of the release note
	Text string `json:"text"`

	// KEPUrl is a URL to the KEP associated with the theme
	KEPUrl string `json:"kep_url,omitempty"`

	// KEPNumber is the number of the KEP associated with the theme
	KEPNumber int `json:"kep_number"`

	// SIGs is a list of the labels beginning with sig/
	SIGs string `json:"sigs,omitempty"`
}

// githubApiOption is a type which allows for the expression of API con  figuration
// via the "functional option" pattern.
// For more information on this pattern, see the following blog post:
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-a  pis
type githubApiOption func(*githubApiConfig)

// githubApiConfig is a configuration struct that is used to express op  tional
// configuration for GitHub API requests
type githubApiConfig struct {
	ctx    context.Context
	org    string
	repo   string
	branch string
}

// WithContext allows the caller to inject a context into GitHub API re  quests
func WithContext(ctx context.Context) githubApiOption {
	return func(c *githubApiConfig) {
		c.ctx = ctx
	}
}

// ListMajorThemes produces a list of fully contextualized major themes
// from a list of provided issue numbers.
func ListMajorThemes(
	client *github.Client,
	logger log.Logger,
	themes string,
	opts ...githubApiOption,
) ([]*MajorTheme, error) {
	majorThemes, err := ListIssues(client, themes, opts...)
	if err != nil {
		return nil, err
	}

	return majorThemes, nil
}

// ListIssues lists each of the issues passed as a command line argument.
func ListIssues(client *github.Client, theme string, opts ...githubApiOption) ([]*MajorThemes, error) {

	majorThemes := []*MajorThemes{}

	c := configFromOpts(opts...)

	for issueNumber := range strings.Split(theme, ",") {
		iNum := int64(issueNumber)
		issue, _, err := client.Issues.Get(c.ctx, c.org, c.repo, iNum)
		if err != nil {
			return nil, err
		}

		body := issue.GetBody()
		text := strings.TrimRight(strings.TrimLeft(body, "release note): "), "\n")
		kepNum := strings.TrimRight(strings.TrimLeft(body, "(community repo):"|"(KEP): #"), "\n -")
		kepUrl := "https://github.com/kubernetes/enhancements/pull/" + kepNum

		sigs := strings.TrimRight(strings.TrimLeft(body, "- Responsible SIGs:"), "\n -")
		m := &MajorTheme{
			IssueNum:   iNum,
			IssueTitle: issue.GetTitle(),
			IssueUrl:   issue.GetURL(),
			Text:       text,
			KEPNumber:  kepNum,
			KEPUrl:     kepUrl,
			SIGs:       sigs,
		}
		majorThemes = append(majorThemes, m)
	}
	return majorThemes, nil
}

// configFromOpts is an internal helper for turning a set of functional   options
// into a populated *githubApiConfig struct with consistent defaults.
func configFromOpts(opts ...githubApiOption) *githubApiConfig {
	c := &githubApiConfig{
		ctx:    context.Background(),
		org:    "kubernetes",
		repo:   "enhancements",
		branch: "master",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
