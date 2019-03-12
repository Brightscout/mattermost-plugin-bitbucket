package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	"net/http"
	"reflect"

	"strings"

	"github.com/google/go-github/github"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/wbrefvem/go-bitbucket"
	bb_webhook "gopkg.in/go-playground/webhooks.v5/bitbucket"
)

func verifyWebhookSignature(secret []byte, signature string, body []byte) bool {
	const signaturePrefix = "sha1="
	const signatureLength = 45

	if len(signature) != signatureLength || !strings.HasPrefix(signature, signaturePrefix) {
		return false
	}

	actual := make([]byte, 20)
	hex.Decode(actual, []byte(signature[5:]))

	return hmac.Equal(signBody(secret, body), actual)
}

func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha1.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}

type GitHubEvent struct {
	// Repo *github.Repository `json:"repository,omitempty"`
	Repo *github.Repository `json:"repository,omitempty"`
}

// Hack to convert from github.PushEventRepository to github.Repository
func ConvertPushEventRepositoryToRepository(pushRepo *github.PushEventRepository) *github.Repository {
	repoName := pushRepo.GetFullName()
	private := pushRepo.GetPrivate()
	return &github.Repository{
		FullName: &repoName,
		Private:  &private,
	}
}

func (p *Plugin) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// config := p.getConfiguration()
	fmt.Printf("----- #### BB webhook.handleWebhook \nr = %+v\n", r)
	config := bitbucket.NewConfiguration()
	client := bitbucket.NewAPIClient(config)
	fmt.Printf("client = %+v\n", client)

	// ctx := context.Background()
	// var githubClient *bitbucket.APIClient

	// gitUser, _, err := githubClient.UsersApi.UsersUsernameGet(ctx, "jfrerich")
	// subtypes, _, err := githubClient.WebhooksApi.HookEventsGet(ctx)
	// IssueComment RepositoriesUsernameRepoSlugIssuesIssueIdCommentsCommentIdGet(ctx, commentId, username, repoSlug, issueId)
	// comment, _, err := githubClient.IssueTrackerApi.RepositoriesUsernameRepoSlugIssuesIssueIdCommentsCommentIdGet(ctx, "51109448", "jfrerich", "mattermost-plugin-readme", "1")
	// fmt.Printf("gitUser = %+v\n", gitUser)
	// fmt.Printf("subtypes = %+v\n", subtypes)
	// fmt.Printf("comment = %+v\n", comment)

	event := r.Header.Get("X-Event-Key")
	bitbucketEvent := bb_webhook.Event(event)

	fmt.Printf("bitbucketEvent = %+v\n", bitbucketEvent)

	fmt.Printf("---- r type= %+v\n\n", reflect.TypeOf(r).String())
	hook, _ := bb_webhook.New()
	payload, err := hook.Parse(r, bb_webhook.IssueCommentCreatedEvent)

	fmt.Printf("---- payload = %+v\n\n\n", payload)
	fmt.Printf("---- payload Type = %+v\n\n", reflect.TypeOf(payload).String())

	var handler func()

	switch payload.(type) {
	case bb_webhook.IssueCommentCreatedPayload:
		handler = func() {
			p.postIssueCommentCreatedEvent(payload)
			return
		}
	}

	if err != nil {
		mlog.Error(err.Error())
		return
	}

	handler()

}

func (p *Plugin) postIssueCommentCreatedEvent(pl interface{}) {
	// func (p *Plugin) postIssueCommentCreatedEvent(r string) (interface{}, error) {
	fmt.Println("_we made it_")
	r := pl.(bb_webhook.IssueCommentCreatedPayload)
	fmt.Printf("\nrelease = %+v\n\n", r)
	fmt.Printf("\nrelease.Actor = %+v\n\n", r.Actor)
	fmt.Printf("release.Actor.Type = %+v\n", r.Actor.Username)
	fmt.Printf("release = %+v\n", r)

	message := fmt.Sprintf("[\\[%s\\]](%s) New comment by [%s](%s) on [#%v %s]:\n\n%s",
		"junk1", "junk2", r.Actor.Username, r.Actor.Username, "junk5", "junk6", "junk7")
	// "junk1", "junk2", r.Actor.Username, r.Actor.Username, comment.Issue.Id, comment.Issue.Title, comment.Content.Raw)

	ctx := context.Background()
	var githubClient *bitbucket.APIClient
	githubClient = p.githubConnect()

	// fmt.Printf("r.Body = %+v\n", r.Body)
	// bodyBytes, err := ioutil.ReadAll(r.Body)
	// bodyString := string(bodyBytes)
	// fmt.Printf("bodyString = %+v\n", bodyString)

	// fmt.Printf("r.Body = %+v\n", r.Body)
	// fmt.Printf("r.Body.String() = %+v\n", r.Body.String())

	// IssueComment RepositoriesUsernameRepoSlugIssuesIssueIdCommentsCommentIdGet(ctx, commentId, username, repoSlug, issueId)
	comment, _, err := githubClient.IssueTrackerApi.RepositoriesUsernameRepoSlugIssuesIssueIdCommentsCommentIdGet(ctx, "51110066", "jfrerich", "mattermost-bitbucket-readme", "1")
	repoInfo, _, err := githubClient.RepositoriesApi.RepositoriesUsernameGet(ctx, "jfrerich", nil)
	fmt.Printf("repoInfo = %+v\n", repoInfo)
	fmt.Printf("repoInfo = %+v\n", repoInfo.Values)
	fmt.Printf("comment = %+v\n", comment)
	fmt.Printf("comment.User = %+v\n", comment.User)
	fmt.Printf("comment.Issue.ID = %+v\n", comment.Issue.Id)
	fmt.Printf("comment.Issue = %+v\n", comment.Issue)
	fmt.Printf("comment.User = %+v\n", comment.User.Username)
	fmt.Printf("comment.Links = %+v\n", comment.Links)
	fmt.Printf("comment.Content = %+v\n", comment.Content)
	fmt.Printf("comment.Content.Raw = %+v\n", comment.Content.Raw)
	fmt.Printf("err = %+v\n", err)
	//
	// message := fmt.Sprintf("[\\[%s\\]](%s) New comment by [%s](%s) on [#%v %s]:\n\n%s",
	// 	"junk1", "junk2", comment.User.Username, comment.User.Username, comment.Issue.Id, comment.Issue.Title, comment.Content.Raw)
	// repo.GetFullName(),
	// repo.GetHTMLURL(),
	// event.GetSender().GetLogin(),
	// event.GetSender().GetHTMLURL(),
	// event.GetIssue().GetNumber(),
	// event.GetIssue().GetTitle(), body)

	config := p.getConfiguration()
	// 	repo := event.GetRepo()
	//
	// repo := "mattermost-bitbucket-readme"
	// subs := p.GetSubscribedChannelsForRepository(repo)
	//
	// 	if subs == nil || len(subs) == 0 {
	// 		return
	// 	}
	//
	userID := ""
	if user, err := p.API.GetUserByUsername(config.Username); err != nil {
		mlog.Error(err.Error())
		return
	} else {
		userID = user.Id
	}
	//
	// 	if event.GetAction() != "created" {
	// 		return
	// 	}
	//
	// 	body := event.GetComment().GetBody()
	//
	// 	// Try to parse out email footer junk
	// 	if strings.Contains(body, "notifications@github.com") {
	// 		body = strings.Split(body, "\n\nOn")[0]
	// 	}
	//
	// 	message := fmt.Sprintf("[\\[%s\\]](%s) New comment by [%s](%s) on [#%v %s]:\n\n%s",
	// 		repo.GetFullName(), repo.GetHTMLURL(), event.GetSender().GetLogin(), event.GetSender().GetHTMLURL(), event.GetIssue().GetNumber(), event.GetIssue().GetTitle(), body)

	// message := fmt.Sprintf("[\\[\\]]() New comment by []:\n\n")
	// fmt.Printf("userID = %+v\n", userID)

	post := &model.Post{
		UserId: userID,
		Type:   "custom_git_comment",
		Props: map[string]interface{}{
			"from_webhook":      "true",
			"override_username": GITHUB_USERNAME,
			"override_icon_url": GITHUB_ICON_URL,
		},
	}
	//
	// 	labels := make([]string, len(event.GetIssue().Labels))
	// 	for i, v := range event.GetIssue().Labels {
	// 		labels[i] = v.GetName()
	// 	}
	//
	// 	for _, sub := range subs {
	// 		if !sub.IssueComments() {
	// 			continue
	// 		}
	//
	// 		label := sub.Label()
	//
	// 		contained := false
	// 		for _, v := range labels {
	// 			if v == label {
	// 				contained = true
	// 			}
	// 		}
	//
	// 		if !contained && label != "" {
	// 			continue
	// 		}
	//
	// 		if event.GetAction() == "created" {
	post.Message = message
	// 		}
	//
	// post.ChannelId = sub.ChannelID
	// post.ChannelId = "yo4c9uz7k7ddxqdbom5rxigpsy"
	post.ChannelId = "kneqppyxxpgz5quwj169xtzmoy"
	// post.ChannelId = "6eerzh9oc7bi3qkhz93qgwyp4a"
	if _, err := p.API.CreatePost(post); err != nil {
		mlog.Error(err.Error())
	}
	// 	}

}

func (p *Plugin) permissionToRepo(userID string, ownerAndRepo string) bool {

	fmt.Println("---- #### BB webhook.permissionToRepo -----")
	// if userID == "" {
	// 	return false
	// }
	//
	// // config := p.getConfiguration()
	// config := bitbucket.NewConfiguration()
	// ctx := context.Background()
	// _, owner, repo := parseOwnerAndRepo(ownerAndRepo, config.EnterpriseBaseURL)
	//
	// if owner == "" {
	// 	return false
	// }
	// if err := p.checkOrg(owner); err != nil {
	// 	return false
	// }

	// info, apiErr := p.getGitHubUserInfo(userID)
	// if apiErr != nil {
	// 	return false
	// }
	// // var githubClient *github.Client
	// var githubClient *bitbucket.APIClient
	// githubClient = p.githubConnect(*info.Token)
	//
	// if result, _, err := githubClient.Repositories.Get(ctx, owner, repo); result == nil || err != nil {
	// 	if err != nil {
	// 		mlog.Error(err.Error())
	// 	}
	// 	return false
	// }
	return true
}

func (p *Plugin) postPullRequestEvent(event *github.PullRequestEvent) {
	config := p.getConfiguration()
	repo := event.GetRepo()

	subs := p.GetSubscribedChannelsForRepository(repo)
	if subs == nil || len(subs) == 0 {
		return
	}

	action := event.GetAction()
	if action != "opened" && action != "labeled" && action != "closed" {
		return
	}

	userID := ""
	if user, err := p.API.GetUserByUsername(config.Username); err != nil {
		mlog.Error(err.Error())
		return
	} else {
		userID = user.Id
	}

	pr := event.GetPullRequest()
	prUser := pr.GetUser()
	eventLabel := event.GetLabel().GetName()
	labels := make([]string, len(pr.Labels))
	for i, v := range pr.Labels {
		labels[i] = v.GetName()
	}

	newPRMessage := fmt.Sprintf(`
#### %s
##### [%s#%v](%s)
#new-pull-request by [%s](%s) on [%s](%s)

%s
`, pr.GetTitle(), repo.GetFullName(), pr.GetNumber(), pr.GetHTMLURL(), prUser.GetLogin(), prUser.GetHTMLURL(), pr.GetCreatedAt().String(), pr.GetHTMLURL(), pr.GetBody())

	fmtCloseMessage := ""
	if pr.GetMerged() {
		fmtCloseMessage = "[%s] Pull request [#%v %s](%s) was merged by [%s](%s)"
	} else {
		fmtCloseMessage = "[%s] Pull request [#%v %s](%s) was closed by [%s](%s)"
	}
	closedPRMessage := fmt.Sprintf(fmtCloseMessage, repo.GetFullName(), pr.GetNumber(), pr.GetTitle(), pr.GetHTMLURL(), event.GetSender().GetLogin(), event.GetSender().GetHTMLURL())

	post := &model.Post{
		UserId: userID,
		Type:   "custom_git_pr",
		Props: map[string]interface{}{
			"from_webhook":      "true",
			"override_username": GITHUB_USERNAME,
			"override_icon_url": GITHUB_ICON_URL,
		},
	}

	for _, sub := range subs {
		if !sub.Pulls() {
			continue
		}

		label := sub.Label()

		contained := false
		for _, v := range labels {
			if v == label {
				contained = true
			}
		}

		if !contained && label != "" {
			continue
		}

		if action == "labeled" {
			if label != "" && label == eventLabel {
				post.Message = fmt.Sprintf("#### %s\n##### [%s#%v](%s)\n#pull-request-labeled `%s` by [%s](%s) on [%s](%s)\n\n%s", pr.GetTitle(), repo.GetFullName(), pr.GetNumber(), pr.GetHTMLURL(), eventLabel, event.GetSender().GetLogin(), event.GetSender().GetHTMLURL(), pr.GetUpdatedAt().String(), pr.GetHTMLURL(), pr.GetBody())
			} else {
				continue
			}
		}

		if action == "opened" {
			post.Message = newPRMessage
		}

		if action == "closed" {
			post.Message = closedPRMessage
		}

		post.ChannelId = sub.ChannelID
		if _, err := p.API.CreatePost(post); err != nil {
			mlog.Error(err.Error())
		}
	}
}

func (p *Plugin) postIssueEvent(event *github.IssuesEvent) {
	config := p.getConfiguration()
	repo := event.GetRepo()

	subs := p.GetSubscribedChannelsForRepository(repo)
	if subs == nil || len(subs) == 0 {
		return
	}

	action := event.GetAction()
	if action != "opened" && action != "labeled" && action != "closed" {
		return
	}

	userID := ""
	if user, err := p.API.GetUserByUsername(config.Username); err != nil {
		mlog.Error(err.Error())
		return
	} else {
		userID = user.Id
	}

	issue := event.GetIssue()
	issueUser := issue.GetUser()
	eventLabel := event.GetLabel().GetName()
	labels := make([]string, len(issue.Labels))
	for i, v := range issue.Labels {
		labels[i] = v.GetName()
	}

	newIssueMessage := fmt.Sprintf(`
#### %s
##### [%s#%v](%s)
#new-issue by [%s](%s) on [%s](%s)

%s
`, issue.GetTitle(), repo.GetFullName(), issue.GetNumber(), issue.GetHTMLURL(), issueUser.GetLogin(), issueUser.GetHTMLURL(), issue.GetCreatedAt().String(), issue.GetHTMLURL(), issue.GetBody())

	closedIssueMessage := fmt.Sprintf("\\[%s] Issue [%s](%s) closed by [%s](%s)",
		repo.GetFullName(), issue.GetTitle(), issue.GetHTMLURL(), event.GetSender().GetLogin(), event.GetSender().GetHTMLURL())

	post := &model.Post{
		UserId: userID,
		Type:   "custom_git_issue",
		Props: map[string]interface{}{
			"from_webhook":      "true",
			"override_username": GITHUB_USERNAME,
			"override_icon_url": GITHUB_ICON_URL,
		},
	}

	for _, sub := range subs {
		if !sub.Issues() {
			continue
		}

		label := sub.Label()

		contained := false
		for _, v := range labels {
			if v == label {
				contained = true
			}
		}

		if !contained && label != "" {
			continue
		}

		if action == "labeled" {
			if label != "" && label == eventLabel {
				post.Message = fmt.Sprintf("#### %s\n##### [%s#%v](%s)\n#issue-labeled `%s` by [%s](%s) on [%s](%s)\n\n%s", issue.GetTitle(), repo.GetFullName(), issue.GetNumber(), issue.GetHTMLURL(), eventLabel, event.GetSender().GetLogin(), event.GetSender().GetHTMLURL(), issue.GetUpdatedAt().String(), issue.GetHTMLURL(), issue.GetBody())
			} else {
				continue
			}
		}

		if action == "opened" {
			post.Message = newIssueMessage
		}

		if action == "closed" {
			post.Message = closedIssueMessage
		}

		post.ChannelId = sub.ChannelID
		if _, err := p.API.CreatePost(post); err != nil {
			mlog.Error(err.Error())
		}
	}
}

func (p *Plugin) postPushEvent(event *github.PushEvent) {
	config := p.getConfiguration()
	repo := event.GetRepo()

	subs := p.GetSubscribedChannelsForRepository(ConvertPushEventRepositoryToRepository(repo))

	if subs == nil || len(subs) == 0 {
		return
	}

	userID := ""
	if user, err := p.API.GetUserByUsername(config.Username); err != nil {
		mlog.Error(err.Error())
		return
	} else {
		userID = user.Id
	}

	forced := event.GetForced()
	branch := strings.Replace(event.GetRef(), "refs/heads/", "", 1)
	commits := event.Commits
	compare_url := event.GetCompare()
	pusher := event.GetSender()

	if len(commits) == 0 {
		return
	}

	fmtMessage := ``
	if forced {
		fmtMessage = "[%s](%s) force-pushed [%d new commits](%s) to [\\[%s:%s\\]](%s/tree/%s):\n"
	} else {
		fmtMessage = "[%s](%s) pushed [%d new commits](%s) to [\\[%s:%s\\]](%s/tree/%s):\n"
	}
	newPushMessage := fmt.Sprintf(fmtMessage, pusher.GetLogin(), pusher.GetHTMLURL(), len(commits), compare_url, repo.GetName(), branch, repo.GetHTMLURL(), branch)
	for _, commit := range commits {
		newPushMessage += fmt.Sprintf("[`%s`](%s) %s - %s\n",
			commit.GetID()[:6], commit.GetURL(), commit.GetMessage(), commit.GetCommitter().GetName())
	}

	post := &model.Post{
		UserId: userID,
		Type:   "custom_git_push",
		Props: map[string]interface{}{
			"from_webhook":      "true",
			"override_username": GITHUB_USERNAME,
			"override_icon_url": GITHUB_ICON_URL,
		},
		Message: newPushMessage,
	}

	for _, sub := range subs {
		if !sub.Pushes() {
			continue
		}

		post.ChannelId = sub.ChannelID
		if _, err := p.API.CreatePost(post); err != nil {
			mlog.Error(err.Error())
		}
	}
}

func (p *Plugin) postCreateEvent(event *github.CreateEvent) {
	config := p.getConfiguration()
	repo := event.GetRepo()

	subs := p.GetSubscribedChannelsForRepository(repo)

	if subs == nil || len(subs) == 0 {
		return
	}

	userID := ""
	if user, err := p.API.GetUserByUsername(config.Username); err != nil {
		mlog.Error(err.Error())
		return
	} else {
		userID = user.Id
	}

	typ := event.GetRefType()
	sender := event.GetSender()
	name := event.GetRef()

	if typ != "tag" && typ != "branch" {
		return
	}

	newCreateMessage := fmt.Sprintf("[%s](%s) just created %s [\\[%s:%s\\]](%s/tree/%s)",
		sender.GetLogin(), sender.GetHTMLURL(), typ, repo.GetName(), name, repo.GetHTMLURL(), name)

	post := &model.Post{
		UserId: userID,
		Type:   "custom_git_create",
		Props: map[string]interface{}{
			"from_webhook":      "true",
			"override_username": GITHUB_USERNAME,
			"override_icon_url": GITHUB_ICON_URL,
		},
		Message: newCreateMessage,
	}

	for _, sub := range subs {
		if !sub.Creates() {
			continue
		}

		post.ChannelId = sub.ChannelID
		if _, err := p.API.CreatePost(post); err != nil {
			mlog.Error(err.Error())
		}
	}
}

func (p *Plugin) postDeleteEvent(event *github.DeleteEvent) {
	config := p.getConfiguration()
	repo := event.GetRepo()

	subs := p.GetSubscribedChannelsForRepository(repo)

	if subs == nil || len(subs) == 0 {
		return
	}

	userID := ""
	if user, err := p.API.GetUserByUsername(config.Username); err != nil {
		mlog.Error(err.Error())
		return
	} else {
		userID = user.Id
	}

	typ := event.GetRefType()
	sender := event.GetSender()
	name := event.GetRef()

	if typ != "tag" && typ != "branch" {
		return
	}

	newDeleteMessage := fmt.Sprintf("[%s](%s) just deleted %s \\[%s:%s]",
		sender.GetLogin(), sender.GetHTMLURL(), typ, repo.GetName(), name)

	post := &model.Post{
		UserId: userID,
		Type:   "custom_git_delete",
		Props: map[string]interface{}{
			"from_webhook":      "true",
			"override_username": GITHUB_USERNAME,
			"override_icon_url": GITHUB_ICON_URL,
		},
		Message: newDeleteMessage,
	}

	for _, sub := range subs {
		if !sub.Deletes() {
			continue
		}

		post.ChannelId = sub.ChannelID
		if _, err := p.API.CreatePost(post); err != nil {
			mlog.Error(err.Error())
		}
	}
}

// func (p *Plugin) postIssueCommentEvent(event *github.IssueCommentEvent) {
// 	config := p.getConfiguration()
// 	repo := event.GetRepo()
//
// 	subs := p.GetSubscribedChannelsForRepository(repo)
//
// 	if subs == nil || len(subs) == 0 {
// 		return
// 	}
//
// 	userID := ""
// 	if user, err := p.API.GetUserByUsername(config.Username); err != nil {
// 		mlog.Error(err.Error())
// 		return
// 	} else {
// 		userID = user.Id
// 	}
//
// 	if event.GetAction() != "created" {
// 		return
// 	}
//
// 	body := event.GetComment().GetBody()
//
// 	// Try to parse out email footer junk
// 	if strings.Contains(body, "notifications@github.com") {
// 		body = strings.Split(body, "\n\nOn")[0]
// 	}
//
// 	message := fmt.Sprintf("[\\[%s\\]](%s) New comment by [%s](%s) on [#%v %s]:\n\n%s",
// 		repo.GetFullName(), repo.GetHTMLURL(), event.GetSender().GetLogin(), event.GetSender().GetHTMLURL(), event.GetIssue().GetNumber(), event.GetIssue().GetTitle(), body)
//
// 	post := &model.Post{
// 		UserId: userID,
// 		Type:   "custom_git_comment",
// 		Props: map[string]interface{}{
// 			"from_webhook":      "true",
// 			"override_username": GITHUB_USERNAME,
// 			"override_icon_url": GITHUB_ICON_URL,
// 		},
// 	}
//
// 	labels := make([]string, len(event.GetIssue().Labels))
// 	for i, v := range event.GetIssue().Labels {
// 		labels[i] = v.GetName()
// 	}
//
// 	for _, sub := range subs {
// 		if !sub.IssueComments() {
// 			continue
// 		}
//
// 		label := sub.Label()
//
// 		contained := false
// 		for _, v := range labels {
// 			if v == label {
// 				contained = true
// 			}
// 		}
//
// 		if !contained && label != "" {
// 			continue
// 		}
//
// 		if event.GetAction() == "created" {
// 			post.Message = message
// 		}
//
// 		post.ChannelId = sub.ChannelID
// 		if _, err := p.API.CreatePost(post); err != nil {
// 			mlog.Error(err.Error())
// 		}
// 	}
// }

func (p *Plugin) postPullRequestReviewEvent(event *github.PullRequestReviewEvent) {
	config := p.getConfiguration()
	repo := event.GetRepo()

	subs := p.GetSubscribedChannelsForRepository(repo)
	if subs == nil || len(subs) == 0 {
		return
	}

	userID := ""
	if user, err := p.API.GetUserByUsername(config.Username); err != nil {
		mlog.Error(err.Error())
		return
	} else {
		userID = user.Id
	}

	action := event.GetAction()
	if action != "submitted" {
		return
	}

	state := event.GetReview().GetState()
	fmtReviewMessage := ""
	switch state {
	case "APPROVED":
		fmtReviewMessage = "[\\[%s\\]](%s) [%s](%s) approved [#%v %s](%s):\n\n%s"
	case "COMMENTED":
		fmtReviewMessage = "[\\[%s\\]](%s) [%s](%s) commented on [#%v %s](%s):\n\n%s"
	case "CHANGES_REQUESTED":
		fmtReviewMessage = "[\\[%s\\]](%s) [%s](%s) requested changes on [#%v %s](%s):\n\n%s"
	default:
		return
	}

	newReviewMessage := fmt.Sprintf(fmtReviewMessage, repo.GetFullName(), repo.GetHTMLURL(), event.GetSender().GetLogin(), event.GetSender().GetHTMLURL(), event.GetPullRequest().GetNumber(), event.GetPullRequest().GetTitle(), event.GetPullRequest().GetHTMLURL(), event.GetReview().GetBody())

	post := &model.Post{
		UserId:  userID,
		Type:    "custom_git_pull_review",
		Message: newReviewMessage,
		Props: map[string]interface{}{
			"from_webhook":      "true",
			"override_username": GITHUB_USERNAME,
			"override_icon_url": GITHUB_ICON_URL,
		},
	}

	labels := make([]string, len(event.GetPullRequest().Labels))
	for i, v := range event.GetPullRequest().Labels {
		labels[i] = v.GetName()
	}

	for _, sub := range subs {
		if !sub.PullReviews() {
			continue
		}

		label := sub.Label()

		contained := false
		for _, v := range labels {
			if v == label {
				contained = true
			}
		}

		if !contained && label != "" {
			continue
		}

		post.ChannelId = sub.ChannelID
		if _, err := p.API.CreatePost(post); err != nil {
			mlog.Error(err.Error())
		}
	}
}

func (p *Plugin) postPullRequestReviewCommentEvent(event *github.PullRequestReviewCommentEvent) {
	config := p.getConfiguration()
	repo := event.GetRepo()

	subs := p.GetSubscribedChannelsForRepository(repo)
	if subs == nil || len(subs) == 0 {
		return
	}

	userID := ""
	if user, err := p.API.GetUserByUsername(config.Username); err != nil {
		mlog.Error(err.Error())
		return
	} else {
		userID = user.Id
	}

	newReviewMessage := fmt.Sprintf("[\\[%s\\]](%s) New review comment by [%s](%s) on [#%v %s](%s):\n\n%s\n%s",
		repo.GetFullName(), repo.GetHTMLURL(), event.GetSender().GetLogin(), event.GetSender().GetHTMLURL(), event.GetPullRequest().GetNumber(), event.GetPullRequest().GetTitle(), event.GetPullRequest().GetHTMLURL(), event.GetComment().GetDiffHunk(), event.GetComment().GetBody())

	post := &model.Post{
		UserId:  userID,
		Type:    "custom_git_pull_review_comment",
		Message: newReviewMessage,
		Props: map[string]interface{}{
			"from_webhook":      "true",
			"override_username": GITHUB_USERNAME,
			"override_icon_url": GITHUB_ICON_URL,
		},
	}

	labels := make([]string, len(event.GetPullRequest().Labels))
	for i, v := range event.GetPullRequest().Labels {
		labels[i] = v.GetName()
	}

	for _, sub := range subs {
		if !sub.PullReviews() {
			continue
		}

		label := sub.Label()

		contained := false
		for _, v := range labels {
			if v == label {
				contained = true
			}
		}

		if !contained && label != "" {
			continue
		}

		post.ChannelId = sub.ChannelID
		if _, err := p.API.CreatePost(post); err != nil {
			mlog.Error(err.Error())
		}
	}
}

func (p *Plugin) handleCommentMentionNotification(event *github.IssueCommentEvent) {
	body := event.GetComment().GetBody()

	// Try to parse out email footer junk
	if strings.Contains(body, "notifications@github.com") {
		body = strings.Split(body, "\n\nOn")[0]
	}

	mentionedUsernames := parseGitHubUsernamesFromText(body)

	message := fmt.Sprintf("[%s](%s) mentioned you on [%s#%v](%s):\n>%s", event.GetSender().GetLogin(), event.GetSender().GetHTMLURL(), event.GetRepo().GetFullName(), event.GetIssue().GetNumber(), event.GetComment().GetHTMLURL(), body)

	post := &model.Post{
		UserId:  p.BotUserID,
		Message: message,
		Type:    "custom_git_mention",
		Props: map[string]interface{}{
			"from_webhook":      "true",
			"override_username": GITHUB_USERNAME,
			"override_icon_url": GITHUB_ICON_URL,
		},
	}

	for _, username := range mentionedUsernames {
		// Don't notify user of their own comment
		if username == event.GetSender().GetLogin() {
			continue
		}

		// Notifications for issue authors are handled separately
		if username == event.GetIssue().GetUser().GetLogin() {
			continue
		}

		userID := p.getGitHubToUserIDMapping(username)
		if userID == "" {
			continue
		}

		if event.GetRepo().GetPrivate() && !p.permissionToRepo(userID, event.GetRepo().GetFullName()) {
			continue
		}

		channel, err := p.API.GetDirectChannel(userID, p.BotUserID)
		if err != nil {
			continue
		}

		post.ChannelId = channel.Id
		_, err = p.API.CreatePost(post)
		if err != nil {
			mlog.Error("Error creating mention post: " + err.Error())
		}

		p.sendRefreshEvent(userID)
	}
}

func (p *Plugin) handleCommentAuthorNotification(event *github.IssueCommentEvent) {
	author := event.GetIssue().GetUser().GetLogin()
	if author == event.GetSender().GetLogin() {
		return
	}

	authorUserID := p.getGitHubToUserIDMapping(author)
	if authorUserID == "" {
		return
	}

	if event.GetRepo().GetPrivate() && !p.permissionToRepo(authorUserID, event.GetRepo().GetFullName()) {
		return
	}

	splitURL := strings.Split(event.GetIssue().GetHTMLURL(), "/")
	if len(splitURL) < 2 {
		return
	}

	message := ""
	switch splitURL[len(splitURL)-2] {
	case "pull":
		message = "[%s](%s) commented on your pull request [%s#%v](%s)"
	case "issues":
		message = "[%s](%s) commented on your issue [%s#%v](%s)"
	}

	message = fmt.Sprintf(message, event.GetSender().GetLogin(), event.GetSender().GetHTMLURL(), event.GetRepo().GetFullName(), event.GetIssue().GetNumber(), event.GetIssue().GetHTMLURL())

	p.CreateBotDMPost(authorUserID, message, "custom_git_author")
	p.sendRefreshEvent(authorUserID)
}

func (p *Plugin) handlePullRequestNotification(event *github.PullRequestEvent) {
	author := event.GetPullRequest().GetUser().GetLogin()
	sender := event.GetSender().GetLogin()
	repoName := event.GetRepo().GetFullName()
	isPrivate := event.GetRepo().GetPrivate()

	requestedReviewer := ""
	requestedUserID := ""
	message := ""
	authorUserID := ""
	assigneeUserID := ""

	switch event.GetAction() {
	case "review_requested":
		requestedReviewer = event.GetRequestedReviewer().GetLogin()
		if requestedReviewer == sender {
			return
		}
		requestedUserID = p.getGitHubToUserIDMapping(requestedReviewer)
		message = "[%s](%s) requested your review on [%s#%v](%s)"
		if isPrivate && !p.permissionToRepo(requestedUserID, repoName) {
			requestedUserID = ""
		}
	case "closed":
		if author == sender {
			return
		}
		if event.GetPullRequest().GetMerged() {
			message = "[%s](%s) merged your pull request [%s#%v](%s)"
		} else {
			message = "[%s](%s) closed your pull request [%s#%v](%s)"
		}
		authorUserID = p.getGitHubToUserIDMapping(author)
		if isPrivate && !p.permissionToRepo(authorUserID, repoName) {
			authorUserID = ""
		}
	case "reopened":
		if author == sender {
			return
		}
		message = "[%s](%s) reopened your pull request [%s#%v](%s)"
		authorUserID = p.getGitHubToUserIDMapping(author)
		if isPrivate && !p.permissionToRepo(authorUserID, repoName) {
			authorUserID = ""
		}
	case "assigned":
		assignee := event.GetPullRequest().GetAssignee().GetLogin()
		if assignee == sender {
			return
		}
		message = "[%s](%s) assigned you to pull request [%s#%v](%s)"
		assigneeUserID = p.getGitHubToUserIDMapping(assignee)
		if isPrivate && !p.permissionToRepo(assigneeUserID, repoName) {
			assigneeUserID = ""
		}
	}

	if len(message) > 0 {
		message = fmt.Sprintf(message, event.GetSender().GetLogin(), event.GetSender().GetHTMLURL(), repoName, event.GetNumber(), event.GetPullRequest().GetHTMLURL())
	}

	if len(requestedUserID) > 0 {
		p.CreateBotDMPost(requestedUserID, message, "custom_git_review_request")
		p.sendRefreshEvent(requestedUserID)
	}

	p.postIssueNotification(message, authorUserID, assigneeUserID)
}

func (p *Plugin) handleIssueNotification(event *github.IssuesEvent) {
	author := event.GetIssue().GetUser().GetLogin()
	sender := event.GetSender().GetLogin()
	if author == sender {
		return
	}
	repoName := event.GetRepo().GetFullName()
	isPrivate := event.GetRepo().GetPrivate()

	message := ""
	authorUserID := ""
	assigneeUserID := ""

	switch event.GetAction() {
	case "closed":
		message = "[%s](%s) closed your issue [%s#%v](%s)"
		authorUserID = p.getGitHubToUserIDMapping(author)
		if isPrivate && !p.permissionToRepo(authorUserID, repoName) {
			authorUserID = ""
		}
	case "reopened":
		message = "[%s](%s) reopened your issue [%s#%v](%s)"
		authorUserID = p.getGitHubToUserIDMapping(author)
		if isPrivate && !p.permissionToRepo(authorUserID, repoName) {
			authorUserID = ""
		}
	case "assigned":
		assignee := event.GetAssignee().GetLogin()
		if assignee == sender {
			return
		}
		message = "[%s](%s) assigned you to issue [%s#%v](%s)"
		assigneeUserID = p.getGitHubToUserIDMapping(assignee)
		if isPrivate && !p.permissionToRepo(assigneeUserID, repoName) {
			assigneeUserID = ""
		}
	}

	if len(message) > 0 {
		message = fmt.Sprintf(message, sender, event.GetSender().GetHTMLURL(), repoName, event.GetIssue().GetNumber(), event.GetIssue().GetHTMLURL())
	}

	p.postIssueNotification(message, authorUserID, assigneeUserID)
}

func (p *Plugin) postIssueNotification(message, authorUserID, assigneeUserID string) {
	if len(authorUserID) > 0 {
		p.CreateBotDMPost(authorUserID, message, "custom_git_author")
		p.sendRefreshEvent(authorUserID)
	}

	if len(assigneeUserID) > 0 {
		p.CreateBotDMPost(assigneeUserID, message, "custom_git_assigned")
		p.sendRefreshEvent(assigneeUserID)
	}
}

func (p *Plugin) handlePullRequestReviewNotification(event *github.PullRequestReviewEvent) {
	author := event.GetPullRequest().GetUser().GetLogin()
	if author == event.GetSender().GetLogin() {
		return
	}

	if event.GetAction() != "submitted" {
		return
	}

	authorUserID := p.getGitHubToUserIDMapping(author)
	if authorUserID == "" {
		return
	}

	if event.GetRepo().GetPrivate() && !p.permissionToRepo(authorUserID, event.GetRepo().GetFullName()) {
		return
	}

	message := ""
	switch event.GetReview().GetState() {
	case "approved":
		message = "[%s](%s) approved your pull request [%s#%v](%s)"
	case "changes_requested":
		message = "[%s](%s) requested changes on your pull request [%s#%v](%s)"
	case "commented":
		message = "[%s](%s) commented on your pull request [%s#%v](%s)"
	}

	message = fmt.Sprintf(message, event.GetSender().GetLogin(), event.GetSender().GetHTMLURL(), event.GetRepo().GetFullName(), event.GetPullRequest().GetNumber(), event.GetPullRequest().GetHTMLURL())

	p.CreateBotDMPost(authorUserID, message, "custom_git_review")
	p.sendRefreshEvent(authorUserID)
}
